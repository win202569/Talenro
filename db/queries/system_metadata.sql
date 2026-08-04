-- name: PutSystemMetadata :exec
INSERT INTO system_metadata (key, value)
VALUES ($1, $2)
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value, updated_at = now();

-- name: GetSystemMetadata :one
SELECT value
FROM system_metadata
WHERE key = $1;
