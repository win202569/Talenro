-- name: InsertAuthorityFencePending :execrows
WITH candidate AS (
  SELECT jsonb_populate_record(
    NULL::nodecontrol.control_plane_authority_fences,
    jsonb_build_object(
      'operation_id', sqlc.arg(operation_id)::uuid,
      'effect_kind', sqlc.arg(effect_kind)::text,
      'scope_kind', sqlc.arg(scope_kind)::text,
      'authority_epoch', sqlc.arg(authority_epoch)::bigint,
      'authority_sequence', sqlc.arg(authority_sequence)::bigint,
      'scope_digest', '\x' || encode(sqlc.arg(scope_digest)::bytea, 'hex'),
      'provider_reservation_digest', '\x' || encode(sqlc.arg(provider_reservation_digest)::bytea, 'hex'),
      'provider_status', 'reserved',
      'visibility_state', 'fence_pending',
      'reserved_at', sqlc.arg(reserved_at)::pg_catalog.timestamptz,
      'authority_protocol_profile', 'legacy_v6'
    )
  ) AS value
)
INSERT INTO nodecontrol.control_plane_authority_fences
SELECT (value).* FROM candidate
ON CONFLICT (operation_id) DO NOTHING;

-- name: GetAuthorityFenceForUpdate :one
SELECT operation_id, effect_kind, scope_kind, authority_epoch, authority_sequence,
       scope_digest, provider_reservation_digest, effect_digest, provider_status,
       provider_receipt_digest, db_system_id, db_timeline, required_lsn, abort_reason,
       visibility_state, reserved_at, effect_bound_at, terminal_at
FROM nodecontrol.control_plane_authority_fences
WHERE operation_id = $1
FOR UPDATE;

-- name: BindAuthorityFenceEffect :execrows
UPDATE nodecontrol.control_plane_authority_fences
SET effect_digest = $2, db_system_id = $3, db_timeline = $4,
    required_lsn = $5, effect_bound_at = $6
WHERE operation_id = $1 AND provider_status = 'reserved' AND visibility_state = 'fence_pending'
  AND effect_digest IS NULL;

-- name: ActivateCommittedAuthorityFence :execrows
UPDATE nodecontrol.control_plane_authority_fences
SET provider_status = 'committed', provider_receipt_digest = $2, visibility_state = 'active', terminal_at = $3
WHERE operation_id = $1 AND provider_status = 'reserved' AND visibility_state = 'fence_pending'
  AND effect_digest = $4 AND db_system_id = $5 AND db_timeline = $6 AND required_lsn = $7;

-- name: AbortAuthorityFence :execrows
UPDATE nodecontrol.control_plane_authority_fences
SET provider_status = 'aborted', provider_receipt_digest = $2, abort_reason = $3,
    visibility_state = 'aborted', terminal_at = $4
WHERE operation_id = $1 AND provider_status = 'reserved' AND visibility_state = 'fence_pending';

-- name: GetAuthorityFence :one
SELECT operation_id, effect_kind, scope_kind, authority_epoch, authority_sequence,
       scope_digest, provider_reservation_digest, effect_digest, provider_status,
       provider_receipt_digest, db_system_id, db_timeline, required_lsn, abort_reason,
       visibility_state, reserved_at, effect_bound_at, terminal_at
FROM nodecontrol.control_plane_authority_fences WHERE operation_id = $1;

-- name: LockAuthorityFence :one
SELECT * FROM nodecontrol.control_plane_authority_fences WHERE operation_id = $1 FOR UPDATE;

-- name: GetStoredAuthorityFence :one
SELECT * FROM nodecontrol.control_plane_authority_fences WHERE operation_id = $1;

-- name: GetAuthorityFenceHead :one
WITH latest_epoch AS (
  SELECT COALESCE(MAX(authority_epoch), 0)::bigint AS authority_epoch
  FROM nodecontrol.control_plane_authority_fences
), epoch_rows AS (
  SELECT f.* FROM nodecontrol.control_plane_authority_fences f
  JOIN latest_epoch e ON f.authority_epoch = e.authority_epoch
), aggregate_head AS (
  SELECT COUNT(*)::bigint AS record_count,
         COALESCE(MAX(authority_sequence), 0)::bigint AS latest_reserved_sequence,
         COALESCE(MAX(authority_sequence) FILTER (
           WHERE provider_status = 'committed' AND visibility_state = 'active'), 0)::bigint AS latest_committed_sequence,
         COUNT(*) FILTER (WHERE provider_status = 'reserved')::bigint AS pending_count
  FROM epoch_rows
), latest_record AS (
  SELECT provider_reservation_digest FROM epoch_rows ORDER BY authority_sequence DESC LIMIT 1
), latest_committed AS (
  SELECT operation_id, provider_receipt_digest, db_system_id, db_timeline, required_lsn
  FROM epoch_rows
  WHERE provider_status = 'committed' AND visibility_state = 'active'
  ORDER BY authority_sequence DESC LIMIT 1
)
SELECT e.authority_epoch, a.record_count, a.latest_reserved_sequence,
       a.latest_committed_sequence, a.pending_count, r.provider_reservation_digest,
       c.operation_id AS latest_committed_operation_id,
       c.provider_receipt_digest, c.db_system_id, c.db_timeline, c.required_lsn,
       (a.record_count <> a.latest_reserved_sequence) AS has_sequence_gap
FROM latest_epoch e CROSS JOIN aggregate_head a
LEFT JOIN latest_record r ON true
LEFT JOIN latest_committed c ON true;

-- name: ListPendingAuthorityFences :many
SELECT operation_id, effect_kind, scope_kind, scope_digest, authority_epoch,
       authority_sequence, provider_reservation_digest, effect_digest,
       db_system_id, db_timeline, required_lsn, reserved_at, effect_bound_at
FROM nodecontrol.control_plane_authority_fences
WHERE authority_epoch = sqlc.arg(authority_epoch)
  AND provider_status = 'reserved' AND visibility_state = 'fence_pending'
ORDER BY authority_sequence;

-- name: GetCommittedNodeCheckpoint :one
SELECT authority_epoch, authority_sequence, provider_receipt_digest
FROM nodecontrol.control_plane_authority_fences
WHERE authority_epoch = sqlc.arg(authority_epoch)
  AND provider_status = 'committed' AND visibility_state = 'active'
  AND (scope_kind = 'global_node_trust'
       OR (scope_kind = 'node' AND scope_digest = sqlc.arg(node_scope_digest)))
ORDER BY authority_sequence DESC
LIMIT 1;

-- name: InsertClaimV1AuthorityFencePending :execrows
INSERT INTO nodecontrol.control_plane_authority_fences
  (operation_id, effect_kind, scope_kind, authority_epoch, authority_sequence, scope_digest,
   provider_reservation_digest, provider_status, visibility_state, reserved_at,
   authority_protocol_profile, protocol_activation_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, 'reserved', 'fence_pending', $8, 'claim_v1', $9)
ON CONFLICT (operation_id) DO NOTHING;

-- name: ClaimAuthorityAbort :execrows
UPDATE nodecontrol.control_plane_authority_fences
SET abort_reason = $2, abort_claimed_at = $3
WHERE operation_id = $1
  AND authority_protocol_profile = 'claim_v1'
  AND provider_status = 'reserved'
  AND visibility_state = 'fence_pending'
  AND effect_digest IS NULL
  AND abort_reason IS NULL
  AND abort_claimed_at IS NULL;

-- name: BeginStagingImport :one
SELECT * FROM nodecontrol.begin_staging_import(
  sqlc.arg(capability_digest),
  sqlc.arg(manifest_digest),
  sqlc.arg(exclusion_lease_digest),
  sqlc.arg(acquisition_head_digest),
  sqlc.arg(route_closed_digest),
  sqlc.arg(projection_body_jcs),
  sqlc.arg(projection_digest),
  sqlc.arg(projection_objects)
);

/* The branch registry below is retained as a literal review aid; sqlc uses the
   ten closed per-owner queries that follow it.
-- closed persisted-outcome owner registry
WITH grant_create AS (
  SELECT 'grant_create'::text AS owner_kind,
         create_authority_effect_commitment_jcs AS commitment_jcs,
         create_authority_effect_commitment_digest AS commitment_digest,
         create_authority_provider_head_jcs AS provider_head_jcs,
         create_authority_provider_head_digest AS provider_head_digest,
         create_authority_checkpoint_anchor_jcs AS checkpoint_anchor_jcs,
         create_authority_checkpoint_anchor_digest AS checkpoint_anchor_digest,
         create_authority_effect_reason AS effect_reason,
         create_authority_attestation_expires_at AS attestation_expires_at,
         create_authority_activation_deadline AS activation_deadline,
         create_authority_expected_provider_identity_digest AS expected_provider_identity_digest,
         create_authority_activation_evidence_jcs AS activation_evidence_jcs,
         create_authority_activation_evidence_digest AS activation_evidence_digest,
         create_authority_effect_resolution_jcs AS effect_resolution_jcs,
         create_authority_effect_resolution_digest AS effect_resolution_digest
  FROM nodecontrol.node_enrollment_grants AS owner WHERE owner.authority_operation_id = $1 FOR UPDATE
), grant_claim AS (
  SELECT 'grant_claim'::text,
         claim_authority_effect_commitment_jcs, claim_authority_effect_commitment_digest,
         claim_authority_provider_head_jcs, claim_authority_provider_head_digest,
         claim_authority_checkpoint_anchor_jcs, claim_authority_checkpoint_anchor_digest,
         claim_authority_effect_reason, claim_authority_attestation_expires_at,
         claim_authority_activation_deadline, claim_authority_expected_provider_identity_digest,
         claim_authority_activation_evidence_jcs, claim_authority_activation_evidence_digest,
         claim_authority_effect_resolution_jcs, claim_authority_effect_resolution_digest
  FROM nodecontrol.node_enrollment_grants AS owner WHERE owner.claim_authority_operation_id = $1 FOR UPDATE
), certificate_activate AS (
  SELECT 'certificate_activate'::text,
         activation_authority_effect_commitment_jcs, activation_authority_effect_commitment_digest,
         activation_authority_provider_head_jcs, activation_authority_provider_head_digest,
         activation_authority_checkpoint_anchor_jcs, activation_authority_checkpoint_anchor_digest,
         activation_authority_effect_reason, activation_authority_attestation_expires_at,
         activation_authority_activation_deadline, activation_authority_expected_provider_identity_digest,
         activation_authority_activation_evidence_jcs, activation_authority_activation_evidence_digest,
         activation_authority_effect_resolution_jcs, activation_authority_effect_resolution_digest
  FROM nodecontrol.node_certificate_issuances AS owner WHERE owner.authority_operation_id = $1 FOR UPDATE
), certificate_revoke AS (
  SELECT 'certificate_revoke'::text,
         revoke_authority_effect_commitment_jcs, revoke_authority_effect_commitment_digest,
         revoke_authority_provider_head_jcs, revoke_authority_provider_head_digest,
         revoke_authority_checkpoint_anchor_jcs, revoke_authority_checkpoint_anchor_digest,
         revoke_authority_effect_reason, revoke_authority_attestation_expires_at,
         revoke_authority_activation_deadline, revoke_authority_expected_provider_identity_digest,
         revoke_authority_activation_evidence_jcs, revoke_authority_activation_evidence_digest,
         revoke_authority_effect_resolution_jcs, revoke_authority_effect_resolution_digest
  FROM nodecontrol.node_certificates AS owner WHERE owner.revoke_authority_operation_id = $1 FOR UPDATE
), state_transition AS (
  SELECT authority_effect_kind,
         authority_effect_commitment_jcs, authority_effect_commitment_digest,
         authority_provider_head_jcs, authority_provider_head_digest,
         authority_checkpoint_anchor_jcs, authority_checkpoint_anchor_digest,
         authority_effect_reason, authority_attestation_expires_at,
         authority_activation_deadline, authority_expected_provider_identity_digest,
         authority_activation_evidence_jcs, authority_activation_evidence_digest,
         authority_effect_resolution_jcs, authority_effect_resolution_digest
  FROM nodecontrol.node_state_transitions AS owner WHERE owner.authority_operation_id = $1 FOR UPDATE
), incident_open AS (
  SELECT 'security_incident_open'::text,
         open_authority_effect_commitment_jcs, open_authority_effect_commitment_digest,
         open_authority_provider_head_jcs, open_authority_provider_head_digest,
         open_authority_checkpoint_anchor_jcs, open_authority_checkpoint_anchor_digest,
         open_authority_effect_reason, open_authority_attestation_expires_at,
         open_authority_activation_deadline, open_authority_expected_provider_identity_digest,
         open_authority_activation_evidence_jcs, open_authority_activation_evidence_digest,
         open_authority_effect_resolution_jcs, open_authority_effect_resolution_digest
  FROM nodecontrol.node_security_incidents AS owner WHERE owner.authority_operation_id = $1 FOR UPDATE
), incident_resolve AS (
  SELECT 'security_incident_resolve'::text,
         resolve_authority_effect_commitment_jcs, resolve_authority_effect_commitment_digest,
         resolve_authority_provider_head_jcs, resolve_authority_provider_head_digest,
         resolve_authority_checkpoint_anchor_jcs, resolve_authority_checkpoint_anchor_digest,
         resolve_authority_effect_reason, resolve_authority_attestation_expires_at,
         resolve_authority_activation_deadline, resolve_authority_expected_provider_identity_digest,
         resolve_authority_activation_evidence_jcs, resolve_authority_activation_evidence_digest,
         resolve_authority_effect_resolution_jcs, resolve_authority_effect_resolution_digest
  FROM nodecontrol.node_security_incidents AS owner WHERE owner.resolution_authority_operation_id = $1 FOR UPDATE
), resource_activate AS (
  SELECT 'resource_envelope_activate'::text,
         activation_authority_effect_commitment_jcs, activation_authority_effect_commitment_digest,
         activation_authority_provider_head_jcs, activation_authority_provider_head_digest,
         activation_authority_checkpoint_anchor_jcs, activation_authority_checkpoint_anchor_digest,
         activation_authority_effect_reason, activation_authority_attestation_expires_at,
         activation_authority_activation_deadline, activation_authority_expected_provider_identity_digest,
         activation_authority_activation_evidence_jcs, activation_authority_activation_evidence_digest,
         activation_authority_effect_resolution_jcs, activation_authority_effect_resolution_digest
  FROM nodecontrol.node_resource_envelopes AS owner WHERE owner.authority_operation_id = $1 FOR UPDATE
), signing_activate AS (
  SELECT CASE signing_kind WHEN 'desired' THEN 'desired_activate'::text ELSE 'recovery_activate'::text END,
         activation_authority_effect_commitment_jcs, activation_authority_effect_commitment_digest,
         activation_authority_provider_head_jcs, activation_authority_provider_head_digest,
         activation_authority_checkpoint_anchor_jcs, activation_authority_checkpoint_anchor_digest,
         activation_authority_effect_reason, activation_authority_attestation_expires_at,
         activation_authority_activation_deadline, activation_authority_expected_provider_identity_digest,
         activation_authority_activation_evidence_jcs, activation_authority_activation_evidence_digest,
         activation_authority_effect_resolution_jcs, activation_authority_effect_resolution_digest
  FROM nodecontrol.node_state_signing_intents AS owner WHERE owner.authority_operation_id = $1 FOR UPDATE
), publish_activate AS (
  SELECT CASE publish_kind WHEN 'root' THEN 'root_publish'::text ELSE 'metadata_publish'::text END,
         activation_authority_effect_commitment_jcs, activation_authority_effect_commitment_digest,
         activation_authority_provider_head_jcs, activation_authority_provider_head_digest,
         activation_authority_checkpoint_anchor_jcs, activation_authority_checkpoint_anchor_digest,
         activation_authority_effect_reason, activation_authority_attestation_expires_at,
         activation_authority_activation_deadline, activation_authority_expected_provider_identity_digest,
         activation_authority_activation_evidence_jcs, activation_authority_activation_evidence_digest,
         activation_authority_effect_resolution_jcs, activation_authority_effect_resolution_digest
  FROM nodecontrol.node_root_metadata_publish_intents AS owner WHERE owner.authority_operation_id = $1 FOR UPDATE
)
SELECT * FROM grant_create
UNION ALL SELECT * FROM grant_claim
UNION ALL SELECT * FROM certificate_activate
UNION ALL SELECT * FROM certificate_revoke
UNION ALL SELECT * FROM state_transition
UNION ALL SELECT * FROM incident_open
UNION ALL SELECT * FROM incident_resolve
UNION ALL SELECT * FROM resource_activate
UNION ALL SELECT * FROM signing_activate
UNION ALL SELECT * FROM publish_activate;
*/

-- name: LockEnrollmentGrantCreateOutcome :one
SELECT create_authority_effect_commitment_jcs AS commitment_jcs,
       create_authority_effect_commitment_digest AS commitment_digest,
       create_authority_provider_head_jcs AS provider_head_jcs,
       create_authority_provider_head_digest AS provider_head_digest,
       create_authority_checkpoint_anchor_jcs AS checkpoint_anchor_jcs,
       create_authority_checkpoint_anchor_digest AS checkpoint_anchor_digest,
       create_authority_effect_reason AS effect_reason,
       create_authority_attestation_expires_at AS attestation_expires_at,
       create_authority_activation_deadline AS activation_deadline,
       create_authority_expected_provider_identity_digest AS expected_provider_identity_digest,
       create_authority_activation_evidence_jcs AS activation_evidence_jcs,
       create_authority_activation_evidence_digest AS activation_evidence_digest,
       create_authority_effect_resolution_jcs AS effect_resolution_jcs,
       create_authority_effect_resolution_digest AS effect_resolution_digest
FROM nodecontrol.node_enrollment_grants
WHERE authority_operation_id = $1 FOR UPDATE;

-- name: LockEnrollmentGrantClaimOutcome :one
SELECT claim_authority_effect_commitment_jcs AS commitment_jcs,
       claim_authority_effect_commitment_digest AS commitment_digest,
       claim_authority_provider_head_jcs AS provider_head_jcs,
       claim_authority_provider_head_digest AS provider_head_digest,
       claim_authority_checkpoint_anchor_jcs AS checkpoint_anchor_jcs,
       claim_authority_checkpoint_anchor_digest AS checkpoint_anchor_digest,
       claim_authority_effect_reason AS effect_reason,
       claim_authority_attestation_expires_at AS attestation_expires_at,
       claim_authority_activation_deadline AS activation_deadline,
       claim_authority_expected_provider_identity_digest AS expected_provider_identity_digest,
       claim_authority_activation_evidence_jcs AS activation_evidence_jcs,
       claim_authority_activation_evidence_digest AS activation_evidence_digest,
       claim_authority_effect_resolution_jcs AS effect_resolution_jcs,
       claim_authority_effect_resolution_digest AS effect_resolution_digest
FROM nodecontrol.node_enrollment_grants
WHERE claim_authority_operation_id = $1 FOR UPDATE;

-- name: LockCertificateIssuanceActivationOutcome :one
SELECT activation_authority_effect_commitment_jcs AS commitment_jcs,
       activation_authority_effect_commitment_digest AS commitment_digest,
       activation_authority_provider_head_jcs AS provider_head_jcs,
       activation_authority_provider_head_digest AS provider_head_digest,
       activation_authority_checkpoint_anchor_jcs AS checkpoint_anchor_jcs,
       activation_authority_checkpoint_anchor_digest AS checkpoint_anchor_digest,
       activation_authority_effect_reason AS effect_reason,
       activation_authority_attestation_expires_at AS attestation_expires_at,
       activation_authority_activation_deadline AS activation_deadline,
       activation_authority_expected_provider_identity_digest AS expected_provider_identity_digest,
       activation_authority_activation_evidence_jcs AS activation_evidence_jcs,
       activation_authority_activation_evidence_digest AS activation_evidence_digest,
       activation_authority_effect_resolution_jcs AS effect_resolution_jcs,
       activation_authority_effect_resolution_digest AS effect_resolution_digest
FROM nodecontrol.node_certificate_issuances
WHERE authority_operation_id = $1 FOR UPDATE;

-- name: LockCertificateRevocationOutcome :one
SELECT revoke_authority_effect_commitment_jcs AS commitment_jcs,
       revoke_authority_effect_commitment_digest AS commitment_digest,
       revoke_authority_provider_head_jcs AS provider_head_jcs,
       revoke_authority_provider_head_digest AS provider_head_digest,
       revoke_authority_checkpoint_anchor_jcs AS checkpoint_anchor_jcs,
       revoke_authority_checkpoint_anchor_digest AS checkpoint_anchor_digest,
       revoke_authority_effect_reason AS effect_reason,
       revoke_authority_attestation_expires_at AS attestation_expires_at,
       revoke_authority_activation_deadline AS activation_deadline,
       revoke_authority_expected_provider_identity_digest AS expected_provider_identity_digest,
       revoke_authority_activation_evidence_jcs AS activation_evidence_jcs,
       revoke_authority_activation_evidence_digest AS activation_evidence_digest,
       revoke_authority_effect_resolution_jcs AS effect_resolution_jcs,
       revoke_authority_effect_resolution_digest AS effect_resolution_digest
FROM nodecontrol.node_certificates
WHERE revoke_authority_operation_id = $1 FOR UPDATE;

-- name: LockStateTransitionOutcome :one
SELECT authority_effect_commitment_jcs AS commitment_jcs,
       authority_effect_commitment_digest AS commitment_digest,
       authority_provider_head_jcs AS provider_head_jcs,
       authority_provider_head_digest AS provider_head_digest,
       authority_checkpoint_anchor_jcs AS checkpoint_anchor_jcs,
       authority_checkpoint_anchor_digest AS checkpoint_anchor_digest,
       authority_effect_reason AS effect_reason,
       authority_attestation_expires_at AS attestation_expires_at,
       authority_activation_deadline AS activation_deadline,
       authority_expected_provider_identity_digest AS expected_provider_identity_digest,
       authority_activation_evidence_jcs AS activation_evidence_jcs,
       authority_activation_evidence_digest AS activation_evidence_digest,
       authority_effect_resolution_jcs AS effect_resolution_jcs,
       authority_effect_resolution_digest AS effect_resolution_digest
FROM nodecontrol.node_state_transitions
WHERE authority_operation_id = $1 FOR UPDATE;

-- name: LockSecurityIncidentOpenOutcome :one
SELECT open_authority_effect_commitment_jcs AS commitment_jcs,
       open_authority_effect_commitment_digest AS commitment_digest,
       open_authority_provider_head_jcs AS provider_head_jcs,
       open_authority_provider_head_digest AS provider_head_digest,
       open_authority_checkpoint_anchor_jcs AS checkpoint_anchor_jcs,
       open_authority_checkpoint_anchor_digest AS checkpoint_anchor_digest,
       open_authority_effect_reason AS effect_reason,
       open_authority_attestation_expires_at AS attestation_expires_at,
       open_authority_activation_deadline AS activation_deadline,
       open_authority_expected_provider_identity_digest AS expected_provider_identity_digest,
       open_authority_activation_evidence_jcs AS activation_evidence_jcs,
       open_authority_activation_evidence_digest AS activation_evidence_digest,
       open_authority_effect_resolution_jcs AS effect_resolution_jcs,
       open_authority_effect_resolution_digest AS effect_resolution_digest
FROM nodecontrol.node_security_incidents
WHERE authority_operation_id = $1 FOR UPDATE;

-- name: LockSecurityIncidentResolveOutcome :one
SELECT resolve_authority_effect_commitment_jcs AS commitment_jcs,
       resolve_authority_effect_commitment_digest AS commitment_digest,
       resolve_authority_provider_head_jcs AS provider_head_jcs,
       resolve_authority_provider_head_digest AS provider_head_digest,
       resolve_authority_checkpoint_anchor_jcs AS checkpoint_anchor_jcs,
       resolve_authority_checkpoint_anchor_digest AS checkpoint_anchor_digest,
       resolve_authority_effect_reason AS effect_reason,
       resolve_authority_attestation_expires_at AS attestation_expires_at,
       resolve_authority_activation_deadline AS activation_deadline,
       resolve_authority_expected_provider_identity_digest AS expected_provider_identity_digest,
       resolve_authority_activation_evidence_jcs AS activation_evidence_jcs,
       resolve_authority_activation_evidence_digest AS activation_evidence_digest,
       resolve_authority_effect_resolution_jcs AS effect_resolution_jcs,
       resolve_authority_effect_resolution_digest AS effect_resolution_digest
FROM nodecontrol.node_security_incidents
WHERE resolution_authority_operation_id = $1 FOR UPDATE;

-- name: LockResourceEnvelopeActivationOutcome :one
SELECT activation_authority_effect_commitment_jcs AS commitment_jcs,
       activation_authority_effect_commitment_digest AS commitment_digest,
       activation_authority_provider_head_jcs AS provider_head_jcs,
       activation_authority_provider_head_digest AS provider_head_digest,
       activation_authority_checkpoint_anchor_jcs AS checkpoint_anchor_jcs,
       activation_authority_checkpoint_anchor_digest AS checkpoint_anchor_digest,
       activation_authority_effect_reason AS effect_reason,
       activation_authority_attestation_expires_at AS attestation_expires_at,
       activation_authority_activation_deadline AS activation_deadline,
       activation_authority_expected_provider_identity_digest AS expected_provider_identity_digest,
       activation_authority_activation_evidence_jcs AS activation_evidence_jcs,
       activation_authority_activation_evidence_digest AS activation_evidence_digest,
       activation_authority_effect_resolution_jcs AS effect_resolution_jcs,
       activation_authority_effect_resolution_digest AS effect_resolution_digest
FROM nodecontrol.node_resource_envelopes
WHERE authority_operation_id = $1 FOR UPDATE;

-- name: LockStateSigningIntentActivationOutcome :one
SELECT activation_authority_effect_commitment_jcs AS commitment_jcs,
       activation_authority_effect_commitment_digest AS commitment_digest,
       activation_authority_provider_head_jcs AS provider_head_jcs,
       activation_authority_provider_head_digest AS provider_head_digest,
       activation_authority_checkpoint_anchor_jcs AS checkpoint_anchor_jcs,
       activation_authority_checkpoint_anchor_digest AS checkpoint_anchor_digest,
       activation_authority_effect_reason AS effect_reason,
       activation_authority_attestation_expires_at AS attestation_expires_at,
       activation_authority_activation_deadline AS activation_deadline,
       activation_authority_expected_provider_identity_digest AS expected_provider_identity_digest,
       activation_authority_activation_evidence_jcs AS activation_evidence_jcs,
       activation_authority_activation_evidence_digest AS activation_evidence_digest,
       activation_authority_effect_resolution_jcs AS effect_resolution_jcs,
       activation_authority_effect_resolution_digest AS effect_resolution_digest
FROM nodecontrol.node_state_signing_intents
WHERE authority_operation_id = $1 FOR UPDATE;

-- name: LockRootMetadataPublishIntentActivationOutcome :one
SELECT activation_authority_effect_commitment_jcs AS commitment_jcs,
       activation_authority_effect_commitment_digest AS commitment_digest,
       activation_authority_provider_head_jcs AS provider_head_jcs,
       activation_authority_provider_head_digest AS provider_head_digest,
       activation_authority_checkpoint_anchor_jcs AS checkpoint_anchor_jcs,
       activation_authority_checkpoint_anchor_digest AS checkpoint_anchor_digest,
       activation_authority_effect_reason AS effect_reason,
       activation_authority_attestation_expires_at AS attestation_expires_at,
       activation_authority_activation_deadline AS activation_deadline,
       activation_authority_expected_provider_identity_digest AS expected_provider_identity_digest,
       activation_authority_activation_evidence_jcs AS activation_evidence_jcs,
       activation_authority_activation_evidence_digest AS activation_evidence_digest,
       activation_authority_effect_resolution_jcs AS effect_resolution_jcs,
       activation_authority_effect_resolution_digest AS effect_resolution_digest
FROM nodecontrol.node_root_metadata_publish_intents
WHERE authority_operation_id = $1 FOR UPDATE;

-- name: GetNodeControlDatabaseIdentity :one
SELECT (CASE WHEN system_identifier < 0
               THEN system_identifier::numeric + 18446744073709551616::numeric
               ELSE system_identifier::numeric
        END)::numeric(20,0) AS system_id,
       (CASE WHEN timeline_id < 0
               THEN timeline_id::bigint + 4294967296::bigint
               ELSE timeline_id::bigint
        END)::bigint AS timeline,
       pg_current_wal_insert_lsn()::text AS required_lsn
FROM pg_control_system(), pg_control_checkpoint();

-- name: AcquireAuthorityV7SourceFreezeForSeal :exec
SELECT nodecontrol.v7_acquire_source_freeze_for_seal();
