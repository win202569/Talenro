-- Private staging projection only: committed/active rows are necessary, not a
-- substitute for Coordinator-bound v7 readiness, incarnation and lease gates.
-- Two rows preserve ambiguity detection; never select an arbitrary first row.
-- name: ListExactCertificateProjection :many
SELECT c.node_id, c.issuer_id, c.serial_bytes, c.leaf_der_sha256, c.public_key_sha256,
       c.certificate_id, c.lineage_id, c.identity_epoch,
       i.identity_epoch AS inventory_identity_epoch,
       i.lineage_id AS inventory_lineage_id, i.inventory_version,
       c.authority_operation_id, c.authority_epoch, c.authority_sequence,
       c.status, c.valid_from, c.valid_until, c.leaf_der,
       (t.purpose IS NOT NULL)::boolean AS has_server_trust,
       COALESCE(t.purpose, '')::text AS trust_purpose,
       COALESCE(t.listener_kind, '')::text AS trust_listener_kind,
       COALESCE(t.trust_domain, '')::text AS trust_domain,
       COALESCE(t.authority_operation_id, '00000000-0000-0000-0000-000000000000'::uuid)::uuid AS trust_operation_id,
       COALESCE(t.authority_epoch, 0)::bigint AS trust_authority_epoch,
       COALESCE(t.authority_sequence, 0)::bigint AS trust_authority_sequence,
       COALESCE(t.bundle_version, 0)::bigint AS trust_bundle_version,
       COALESCE(t.bundle_digest, '\x'::bytea)::bytea AS trust_bundle_digest
FROM nodecontrol.node_certificates c
JOIN nodecontrol.node_inventory i
  ON i.node_id = c.node_id
 AND i.identity_epoch = c.identity_epoch
 AND i.lineage_id = c.lineage_id
JOIN nodecontrol.control_plane_authority_fences cf
  ON cf.operation_id = c.authority_operation_id
 AND cf.authority_epoch = c.authority_epoch
 AND cf.authority_sequence = c.authority_sequence
 AND cf.effect_kind = 'certificate_activate'
 AND cf.scope_kind = 'node'
 AND cf.scope_digest = sqlc.arg(node_scope_digest)::bytea
 AND cf.provider_status = 'committed'
 AND cf.visibility_state = 'active'
LEFT JOIN nodecontrol.control_plane_trust_bundle_high_waters t
  ON sqlc.arg(require_agent_trust)::boolean
 AND t.purpose = 'agent_server'
 AND t.listener_kind = 'agent'
 AND t.trust_domain = sqlc.arg(trust_domain)::text
 AND t.authority_epoch = c.authority_epoch
LEFT JOIN nodecontrol.control_plane_authority_fences tf
  ON tf.operation_id = t.authority_operation_id
 AND tf.authority_epoch = t.authority_epoch
 AND tf.authority_sequence = t.authority_sequence
 AND tf.effect_kind = 'trust_bundle_publish'
 AND tf.scope_kind = 'global_node_trust'
 AND tf.scope_digest = sqlc.arg(trust_scope_digest)::bytea
 AND tf.provider_status = 'committed'
 AND tf.visibility_state = 'active'
WHERE c.node_id = sqlc.arg(node_id)::uuid
  AND c.issuer_id = sqlc.arg(issuer_id)::text
  AND c.serial_bytes = sqlc.arg(serial_bytes)::bytea
  AND c.leaf_der_sha256 = sqlc.arg(leaf_der_digest)::bytea
  AND c.public_key_sha256 = sqlc.arg(public_key_digest)::bytea
  AND c.authority_epoch = sqlc.arg(authority_epoch)::bigint
  AND c.status IN ('active', 'recovery_pending', 'recovery_limited')
  AND (NOT sqlc.arg(require_agent_trust)::boolean OR tf.operation_id IS NOT NULL)
LIMIT 2;

-- Private fixed-base staging only. Not a canonical envelope or authorization.
-- Invalid pointers retain a diagnostic row; no alternative current artifact.
-- name: ListExactDesiredProjection :many
WITH args AS (
  SELECT sqlc.arg(node_id)::uuid AS node_id,
         sqlc.arg(authority_epoch)::bigint AS authority_epoch,
         sqlc.arg(node_scope_digest)::bytea AS node_scope_digest,
         sqlc.arg(trust_scope_digest)::bytea AS trust_scope_digest
), c AS (
  SELECT i.node_id,
    COALESCE(
      ((i.active_root_publish_id IS NULL AND i.active_root_version IS NULL) OR
       (i.active_root_publish_id IS NOT NULL AND i.active_root_publish_id <> '00000000-0000-0000-0000-000000000000'::uuid AND i.active_root_version > 0))
      AND ((i.active_metadata_publish_id IS NULL AND i.active_metadata_version IS NULL) OR
       (i.active_metadata_publish_id IS NOT NULL AND i.active_metadata_publish_id <> '00000000-0000-0000-0000-000000000000'::uuid AND i.active_metadata_version > 0))
      AND (i.active_desired_generation IS NULL OR i.active_desired_generation > 0), false)::boolean AS pointer_groups_valid,
    (i.active_desired_generation IS NOT NULL)::boolean AS has_desired,
    COALESCE(i.active_desired_generation, 0)::bigint AS active_desired_generation,
    COALESCE(i.active_root_publish_id, '00000000-0000-0000-0000-000000000000'::uuid)::uuid AS active_root_publish_id,
    COALESCE(i.active_root_version, 0)::bigint AS active_root_version,
    COALESCE(i.active_metadata_publish_id, '00000000-0000-0000-0000-000000000000'::uuid)::uuid AS active_metadata_publish_id,
    COALESCE(i.active_metadata_version, 0)::bigint AS active_metadata_version,
    COALESCE(
      i.active_root_version BETWEEN 1 AND 33 AND i.active_metadata_version > 0
      AND d.node_id IS NOT NULL AND d.authority_epoch = a.authority_epoch
      AND s.signing_id IS NOT NULL AND df.operation_id IS NOT NULL
      AND r.publish_id IS NOT NULL AND rf.operation_id IS NOT NULL
      AND m.publish_id IS NOT NULL AND mf.operation_id IS NOT NULL
      AND octet_length(d.canonical_payload) BETWEEN 1 AND 65536
      AND octet_length(d.payload_digest) = 32 AND octet_length(d.signature) = 64
      AND octet_length(d.resource_envelope_digest) = 32 AND octet_length(d.signing_key_id) = 32
      AND isfinite(d.issued_at) AND isfinite(d.effective_deadline) AND isfinite(d.valid_until)
      AND r.status = 'active' AND r.terminal_at IS NOT NULL AND isfinite(r.terminal_at)
      AND r.failure_reason IS NULL
      AND octet_length(r.canonical_payload) BETWEEN 1 AND 65536
      AND octet_length(r.published_envelope) BETWEEN 1 AND 262144
      AND octet_length(r.payload_digest) = 32 AND octet_length(r.published_envelope_digest) = 32
      AND m.status = 'active' AND m.terminal_at IS NOT NULL AND isfinite(m.terminal_at)
      AND m.failure_reason IS NULL
      AND octet_length(m.canonical_payload) BETWEEN 1 AND 65536
      AND octet_length(m.published_envelope) BETWEEN 1 AND 262144
      AND octet_length(m.payload_digest) = 32 AND octet_length(m.published_envelope_digest) = 32, false)::boolean AS current_valid,
    d.node_id AS desired_node_id,
    d.signing_id AS desired_signing_id,
    d.authority_operation_id AS desired_authority_operation_id,
    d.root_publish_id AS desired_root_publish_id,
    d.metadata_publish_id AS desired_metadata_publish_id,
    d.generation AS desired_generation,
    d.inventory_version AS desired_inventory_version,
    d.resource_envelope_version AS desired_resource_envelope_version,
    d.root_version AS desired_root_version,
    d.metadata_version AS desired_metadata_version,
    d.authority_epoch AS desired_authority_epoch,
    d.authority_sequence AS desired_authority_sequence,
    d.resource_envelope_digest AS desired_resource_envelope_digest,
    d.signing_key_id AS desired_signing_key_id,
    d.payload_digest AS desired_payload_digest,
    d.signature AS desired_signature,
    d.canonical_payload AS desired_payload,
    d.issued_at AS desired_issued_at,
    d.effective_deadline AS desired_effective_deadline,
    d.valid_until AS desired_valid_until,
    r.publish_id AS root_publish_id,
    r.authority_operation_id AS root_authority_operation_id,
    r.reserved_version AS root_version,
    r.base_root_version AS root_base_root_version,
    r.base_metadata_version AS root_base_metadata_version,
    r.authority_epoch AS root_authority_epoch,
    r.authority_sequence AS root_authority_sequence,
    r.canonical_payload AS root_payload,
    r.published_envelope AS root_published_envelope,
    r.payload_digest AS root_payload_digest,
    r.published_envelope_digest AS root_envelope_digest,
    m.publish_id AS metadata_publish_id,
    m.authority_operation_id AS metadata_authority_operation_id,
    m.reserved_version AS metadata_version,
    m.base_root_version AS metadata_base_root_version,
    m.base_metadata_version AS metadata_base_metadata_version,
    m.authority_epoch AS metadata_authority_epoch,
    m.authority_sequence AS metadata_authority_sequence,
    m.canonical_payload AS metadata_payload,
    m.published_envelope AS metadata_published_envelope,
    m.payload_digest AS metadata_payload_digest,
    m.published_envelope_digest AS metadata_envelope_digest
  FROM nodecontrol.node_inventory i
  CROSS JOIN args a
  LEFT JOIN nodecontrol.node_desired_states d
    ON d.node_id = i.node_id AND d.generation = i.active_desired_generation
   AND d.root_publish_id = i.active_root_publish_id AND d.root_version = i.active_root_version
   AND d.metadata_publish_id = i.active_metadata_publish_id AND d.metadata_version = i.active_metadata_version
  LEFT JOIN nodecontrol.node_state_signing_intents s
    ON s.signing_id = d.signing_id AND s.node_id = d.node_id
   AND s.reserved_generation = d.generation
   AND s.authority_operation_id = d.authority_operation_id
   AND s.authority_epoch = d.authority_epoch AND s.authority_sequence = d.authority_sequence
   AND s.signing_kind = 'desired' AND s.status = 'active'
   AND s.canonical_payload = d.canonical_payload AND s.payload_digest = d.payload_digest
   AND s.signature = d.signature AND s.expected_key_id = d.signing_key_id
   AND s.root_version = d.root_version AND s.metadata_version = d.metadata_version
   AND s.captured_inventory_version = d.inventory_version
   AND s.signature_verified_at IS NOT NULL AND isfinite(s.signature_verified_at)
   AND s.terminal_at IS NOT NULL AND isfinite(s.terminal_at) AND s.failure_reason IS NULL
  LEFT JOIN nodecontrol.control_plane_authority_fences df
    ON df.operation_id = d.authority_operation_id
   AND df.authority_epoch = d.authority_epoch AND df.authority_sequence = d.authority_sequence
   AND df.effect_kind = 'desired_activate' AND df.scope_kind = 'node'
   AND df.scope_digest = a.node_scope_digest
   AND df.provider_status = 'committed' AND df.visibility_state = 'active'
  LEFT JOIN nodecontrol.node_root_metadata_publish_intents r
    ON r.publish_id = i.active_root_publish_id AND r.reserved_version = i.active_root_version
   AND r.publish_kind = 'root' AND r.authority_epoch = a.authority_epoch
   AND r.base_root_version = r.reserved_version - 1
  LEFT JOIN nodecontrol.control_plane_authority_fences rf
    ON rf.operation_id = r.authority_operation_id
   AND rf.authority_epoch = r.authority_epoch AND rf.authority_sequence = r.authority_sequence
   AND rf.effect_kind = 'root_publish' AND rf.scope_kind = 'global_node_trust'
   AND rf.scope_digest = a.trust_scope_digest
   AND rf.provider_status = 'committed' AND rf.visibility_state = 'active'
  LEFT JOIN nodecontrol.node_root_metadata_publish_intents m
    ON m.publish_id = i.active_metadata_publish_id AND m.reserved_version = i.active_metadata_version
   AND m.publish_kind = 'metadata' AND m.authority_epoch = a.authority_epoch
   AND m.base_metadata_version = m.reserved_version - 1 AND m.base_root_version = i.active_root_version
  LEFT JOIN nodecontrol.control_plane_authority_fences mf
    ON mf.operation_id = m.authority_operation_id
   AND mf.authority_epoch = m.authority_epoch AND mf.authority_sequence = m.authority_sequence
   AND mf.effect_kind = 'metadata_publish' AND mf.scope_kind = 'global_node_trust'
   AND mf.scope_digest = a.trust_scope_digest
   AND mf.provider_status = 'committed' AND mf.visibility_state = 'active'
  WHERE i.node_id = a.node_id
)
SELECT c.node_id, c.pointer_groups_valid, c.has_desired, c.current_valid,
       c.active_desired_generation, c.active_root_publish_id, c.active_root_version,
       c.active_metadata_publish_id, c.active_metadata_version,
       (v.version IS NOT NULL)::boolean AS has_rotation,
       COALESCE(v.version, 0)::bigint AS rotation_expected_version,
       COALESCE(v.version IS NOT NULL AND h.publish_id IS NOT NULL AND hf.operation_id IS NOT NULL
         AND h.status = 'active' AND h.terminal_at IS NOT NULL AND isfinite(h.terminal_at)
      AND h.failure_reason IS NULL
      AND octet_length(h.canonical_payload) BETWEEN 1 AND 65536
      AND octet_length(h.published_envelope) BETWEEN 1 AND 262144
      AND octet_length(h.payload_digest) = 32 AND octet_length(h.published_envelope_digest) = 32, false)::boolean AS rotation_valid,
       (CASE WHEN c.current_valid THEN COALESCE(c.desired_node_id, '00000000-0000-0000-0000-000000000000'::uuid) ELSE '00000000-0000-0000-0000-000000000000'::uuid END)::uuid AS desired_node_id,
       (CASE WHEN c.current_valid THEN COALESCE(c.desired_signing_id, '00000000-0000-0000-0000-000000000000'::uuid) ELSE '00000000-0000-0000-0000-000000000000'::uuid END)::uuid AS desired_signing_id,
       (CASE WHEN c.current_valid THEN COALESCE(c.desired_authority_operation_id, '00000000-0000-0000-0000-000000000000'::uuid) ELSE '00000000-0000-0000-0000-000000000000'::uuid END)::uuid AS desired_authority_operation_id,
       (CASE WHEN c.current_valid THEN COALESCE(c.desired_root_publish_id, '00000000-0000-0000-0000-000000000000'::uuid) ELSE '00000000-0000-0000-0000-000000000000'::uuid END)::uuid AS desired_root_publish_id,
       (CASE WHEN c.current_valid THEN COALESCE(c.desired_metadata_publish_id, '00000000-0000-0000-0000-000000000000'::uuid) ELSE '00000000-0000-0000-0000-000000000000'::uuid END)::uuid AS desired_metadata_publish_id,
       (CASE WHEN c.current_valid THEN COALESCE(c.desired_generation, 0::bigint) ELSE 0::bigint END)::bigint AS desired_generation,
       (CASE WHEN c.current_valid THEN COALESCE(c.desired_inventory_version, 0::bigint) ELSE 0::bigint END)::bigint AS desired_inventory_version,
       (CASE WHEN c.current_valid THEN COALESCE(c.desired_resource_envelope_version, 0::bigint) ELSE 0::bigint END)::bigint AS desired_resource_envelope_version,
       (CASE WHEN c.current_valid THEN COALESCE(c.desired_root_version, 0::bigint) ELSE 0::bigint END)::bigint AS desired_root_version,
       (CASE WHEN c.current_valid THEN COALESCE(c.desired_metadata_version, 0::bigint) ELSE 0::bigint END)::bigint AS desired_metadata_version,
       (CASE WHEN c.current_valid THEN COALESCE(c.desired_authority_epoch, 0::bigint) ELSE 0::bigint END)::bigint AS desired_authority_epoch,
       (CASE WHEN c.current_valid THEN COALESCE(c.desired_authority_sequence, 0::bigint) ELSE 0::bigint END)::bigint AS desired_authority_sequence,
       (CASE WHEN c.current_valid THEN COALESCE(c.desired_resource_envelope_digest, '\x'::bytea) ELSE '\x'::bytea END)::bytea AS desired_resource_envelope_digest,
       (CASE WHEN c.current_valid THEN COALESCE(c.desired_signing_key_id, '\x'::bytea) ELSE '\x'::bytea END)::bytea AS desired_signing_key_id,
       (CASE WHEN c.current_valid THEN COALESCE(c.desired_payload_digest, '\x'::bytea) ELSE '\x'::bytea END)::bytea AS desired_payload_digest,
       (CASE WHEN c.current_valid THEN COALESCE(c.desired_signature, '\x'::bytea) ELSE '\x'::bytea END)::bytea AS desired_signature,
       (CASE WHEN c.current_valid THEN COALESCE(c.desired_payload, '\x'::bytea) ELSE '\x'::bytea END)::bytea AS desired_payload,
       c.desired_issued_at,
       c.desired_effective_deadline,
       c.desired_valid_until,
       (CASE WHEN c.current_valid THEN COALESCE(c.root_publish_id, '00000000-0000-0000-0000-000000000000'::uuid) ELSE '00000000-0000-0000-0000-000000000000'::uuid END)::uuid AS root_publish_id,
       (CASE WHEN c.current_valid THEN COALESCE(c.root_authority_operation_id, '00000000-0000-0000-0000-000000000000'::uuid) ELSE '00000000-0000-0000-0000-000000000000'::uuid END)::uuid AS root_authority_operation_id,
       (CASE WHEN c.current_valid THEN COALESCE(c.root_version, 0::bigint) ELSE 0::bigint END)::bigint AS root_version,
       (CASE WHEN c.current_valid THEN COALESCE(c.root_base_root_version, 0::bigint) ELSE 0::bigint END)::bigint AS root_base_root_version,
       (CASE WHEN c.current_valid THEN COALESCE(c.root_base_metadata_version, 0::bigint) ELSE 0::bigint END)::bigint AS root_base_metadata_version,
       (CASE WHEN c.current_valid THEN COALESCE(c.root_authority_epoch, 0::bigint) ELSE 0::bigint END)::bigint AS root_authority_epoch,
       (CASE WHEN c.current_valid THEN COALESCE(c.root_authority_sequence, 0::bigint) ELSE 0::bigint END)::bigint AS root_authority_sequence,
       (CASE WHEN c.current_valid THEN COALESCE(c.root_payload, '\x'::bytea) ELSE '\x'::bytea END)::bytea AS root_payload,
       (CASE WHEN c.current_valid THEN COALESCE(c.root_published_envelope, '\x'::bytea) ELSE '\x'::bytea END)::bytea AS root_published_envelope,
       (CASE WHEN c.current_valid THEN COALESCE(c.root_payload_digest, '\x'::bytea) ELSE '\x'::bytea END)::bytea AS root_payload_digest,
       (CASE WHEN c.current_valid THEN COALESCE(c.root_envelope_digest, '\x'::bytea) ELSE '\x'::bytea END)::bytea AS root_envelope_digest,
       (CASE WHEN c.current_valid THEN COALESCE(c.metadata_publish_id, '00000000-0000-0000-0000-000000000000'::uuid) ELSE '00000000-0000-0000-0000-000000000000'::uuid END)::uuid AS metadata_publish_id,
       (CASE WHEN c.current_valid THEN COALESCE(c.metadata_authority_operation_id, '00000000-0000-0000-0000-000000000000'::uuid) ELSE '00000000-0000-0000-0000-000000000000'::uuid END)::uuid AS metadata_authority_operation_id,
       (CASE WHEN c.current_valid THEN COALESCE(c.metadata_version, 0::bigint) ELSE 0::bigint END)::bigint AS metadata_version,
       (CASE WHEN c.current_valid THEN COALESCE(c.metadata_base_root_version, 0::bigint) ELSE 0::bigint END)::bigint AS metadata_base_root_version,
       (CASE WHEN c.current_valid THEN COALESCE(c.metadata_base_metadata_version, 0::bigint) ELSE 0::bigint END)::bigint AS metadata_base_metadata_version,
       (CASE WHEN c.current_valid THEN COALESCE(c.metadata_authority_epoch, 0::bigint) ELSE 0::bigint END)::bigint AS metadata_authority_epoch,
       (CASE WHEN c.current_valid THEN COALESCE(c.metadata_authority_sequence, 0::bigint) ELSE 0::bigint END)::bigint AS metadata_authority_sequence,
       (CASE WHEN c.current_valid THEN COALESCE(c.metadata_payload, '\x'::bytea) ELSE '\x'::bytea END)::bytea AS metadata_payload,
       (CASE WHEN c.current_valid THEN COALESCE(c.metadata_published_envelope, '\x'::bytea) ELSE '\x'::bytea END)::bytea AS metadata_published_envelope,
       (CASE WHEN c.current_valid THEN COALESCE(c.metadata_payload_digest, '\x'::bytea) ELSE '\x'::bytea END)::bytea AS metadata_payload_digest,
       (CASE WHEN c.current_valid THEN COALESCE(c.metadata_envelope_digest, '\x'::bytea) ELSE '\x'::bytea END)::bytea AS metadata_envelope_digest,
       (CASE WHEN c.current_valid THEN COALESCE(h.publish_id, '00000000-0000-0000-0000-000000000000'::uuid) ELSE '00000000-0000-0000-0000-000000000000'::uuid END)::uuid AS rotation_publish_id,
       (CASE WHEN c.current_valid THEN COALESCE(h.authority_operation_id, '00000000-0000-0000-0000-000000000000'::uuid) ELSE '00000000-0000-0000-0000-000000000000'::uuid END)::uuid AS rotation_authority_operation_id,
       (CASE WHEN c.current_valid THEN COALESCE(h.reserved_version, 0::bigint) ELSE 0::bigint END)::bigint AS rotation_version,
       (CASE WHEN c.current_valid THEN COALESCE(h.base_root_version, 0::bigint) ELSE 0::bigint END)::bigint AS rotation_base_root_version,
       (CASE WHEN c.current_valid THEN COALESCE(h.base_metadata_version, 0::bigint) ELSE 0::bigint END)::bigint AS rotation_base_metadata_version,
       (CASE WHEN c.current_valid THEN COALESCE(h.authority_epoch, 0::bigint) ELSE 0::bigint END)::bigint AS rotation_authority_epoch,
       (CASE WHEN c.current_valid THEN COALESCE(h.authority_sequence, 0::bigint) ELSE 0::bigint END)::bigint AS rotation_authority_sequence,
       (CASE WHEN c.current_valid AND octet_length(h.canonical_payload) BETWEEN 1 AND 65536 THEN COALESCE(h.canonical_payload, '\x'::bytea) ELSE '\x'::bytea END)::bytea AS rotation_payload,
       (CASE WHEN c.current_valid AND octet_length(h.published_envelope) BETWEEN 1 AND 262144 THEN COALESCE(h.published_envelope, '\x'::bytea) ELSE '\x'::bytea END)::bytea AS rotation_published_envelope,
       (CASE WHEN c.current_valid AND octet_length(h.payload_digest) BETWEEN 1 AND 32 THEN COALESCE(h.payload_digest, '\x'::bytea) ELSE '\x'::bytea END)::bytea AS rotation_payload_digest,
       (CASE WHEN c.current_valid AND octet_length(h.published_envelope_digest) BETWEEN 1 AND 32 THEN COALESCE(h.published_envelope_digest, '\x'::bytea) ELSE '\x'::bytea END)::bytea AS rotation_envelope_digest
FROM c
CROSS JOIN args a
LEFT JOIN LATERAL generate_series(2::bigint, CASE WHEN c.current_valid AND c.active_root_version BETWEEN 1 AND 33 THEN c.active_root_version ELSE 1::bigint END) AS v(version) ON true
LEFT JOIN nodecontrol.node_root_metadata_publish_intents h
  ON h.publish_kind = 'root' AND h.reserved_version = v.version
 AND h.base_root_version = v.version - 1 AND h.authority_epoch = a.authority_epoch
 AND h.reason = 'root_rotation'
LEFT JOIN nodecontrol.control_plane_authority_fences hf
    ON hf.operation_id = h.authority_operation_id
   AND hf.authority_epoch = h.authority_epoch AND hf.authority_sequence = h.authority_sequence
   AND hf.effect_kind = 'root_publish' AND hf.scope_kind = 'global_node_trust'
   AND hf.scope_digest = a.trust_scope_digest
   AND hf.provider_status = 'committed' AND hf.visibility_state = 'active'
ORDER BY rotation_expected_version
LIMIT 33;

