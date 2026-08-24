-- name: InsertAuthorityFencePending :execrows
INSERT INTO nodecontrol.control_plane_authority_fences
  (operation_id, effect_kind, scope_kind, authority_epoch, authority_sequence, scope_digest,
   provider_reservation_digest, provider_status, visibility_state, reserved_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, 'reserved', 'fence_pending', $8)
ON CONFLICT (operation_id) DO NOTHING;

-- name: GetAuthorityFenceForUpdate :one
SELECT * FROM nodecontrol.control_plane_authority_fences
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
