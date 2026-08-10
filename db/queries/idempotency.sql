-- name: TryBeginIdempotency :execrows
INSERT INTO idempotency_records
  (principal_scope, operation, idempotency_key_hash, request_digest, state, created_at, expires_at)
VALUES ($1,$2,$3,$4,'in_progress',$5,$6)
ON CONFLICT (principal_scope, operation, idempotency_key_hash) DO NOTHING;

-- name: GetIdempotencyForUpdate :one
SELECT * FROM idempotency_records
WHERE principal_scope=$1 AND operation=$2 AND idempotency_key_hash=$3
FOR UPDATE;

-- name: CompleteIdempotency :one
UPDATE idempotency_records
SET state = 'completed', response_status = $5, response_ciphertext = $6, response_key_version=$7
WHERE principal_scope = $1 AND operation = $2 AND idempotency_key_hash = $3
  AND request_digest=$4 AND state = 'in_progress'
RETURNING *;

-- name: PruneIdempotencyRecords :execrows
DELETE FROM idempotency_records WHERE expires_at < $1 AND state IN ('completed','failed');
