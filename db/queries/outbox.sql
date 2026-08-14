-- name: ClaimOutboxBatch :many
WITH candidates AS (
  SELECT candidate.event_id FROM transactional_outbox AS candidate
  WHERE candidate.published_at IS NULL AND candidate.available_at <= $1
    AND (candidate.claimed_until IS NULL OR candidate.claimed_until < $1)
  ORDER BY candidate.occurred_at
  LIMIT $2
  FOR UPDATE OF candidate SKIP LOCKED
)
UPDATE transactional_outbox o
SET claimed_until = $3, attempts = o.attempts + 1
FROM candidates c
WHERE o.event_id = c.event_id
RETURNING o.*;

-- name: MarkOutboxPublished :execrows
UPDATE transactional_outbox SET published_at = $2, claimed_until = NULL
WHERE event_id = $1 AND published_at IS NULL;

-- name: RecordConsumedEvent :execrows
INSERT INTO consumed_event_ids (consumer, event_id, consumed_at, expires_at)
VALUES ($1,$2,$3,$4)
ON CONFLICT DO NOTHING;

-- name: HasConsumedEvent :one
SELECT EXISTS (
  SELECT 1 FROM consumed_event_ids
  WHERE consumer=$1 AND event_id=$2
);

-- name: InsertOutboxEvent :exec
INSERT INTO transactional_outbox
  (event_id,event_type,aggregate_type,aggregate_id,aggregate_version,
   idempotency_key,payload,occurred_at,available_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9);

-- name: ReleaseOutboxClaim :execrows
UPDATE transactional_outbox
SET claimed_until=NULL, available_at=$2
WHERE event_id=$1 AND published_at IS NULL;

-- name: GetOutboxHealth :one
SELECT count(*)::bigint AS backlog,
       COALESCE(EXTRACT(EPOCH FROM ($1-min(occurred_at))),0)::double precision AS oldest_age_seconds
FROM transactional_outbox WHERE published_at IS NULL;

-- name: PruneConsumedEvents :execrows
DELETE FROM consumed_event_ids WHERE expires_at < $1;
