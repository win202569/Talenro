-- name: CreateAccount :exec
INSERT INTO identity.accounts (id, state, state_version, locale, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetAccountForUpdate :one
SELECT * FROM identity.accounts WHERE id = $1 FOR UPDATE;

-- name: FindIdentityByLookupDigest :one
SELECT * FROM identity.email_identities WHERE lookup_digest = $1 FOR UPDATE;

-- name: FindIdentityByLookupDigestRead :one
SELECT * FROM identity.email_identities WHERE lookup_digest = $1;

-- name: LockEmailLookupDigest :exec
SELECT pg_advisory_xact_lock(
  hashtextextended(encode(sqlc.arg(lookup_digest)::bytea, 'hex'), 0)
);

-- name: ActivateVerifiedAccount :one
UPDATE identity.accounts
SET state = 'active', state_version = state_version + 1, updated_at = $2
WHERE id = $1 AND state = 'pending_email'
RETURNING *;

-- name: ConsumeEmailVerification :one
UPDATE identity.email_identities
SET verification_consumed_at = $3, verified_at = $3, updated_at = $3
WHERE principal_id = $1
  AND verification_token_hash = $2
  AND verification_consumed_at IS NULL
  AND verification_expires_at >= $3
RETURNING *;

-- name: GetEmailVerificationForUpdate :one
SELECT * FROM identity.email_identities
WHERE verification_token_hash = $1
FOR UPDATE;

-- name: GetPasswordCredential :one
SELECT * FROM identity.password_credentials WHERE principal_id = $1 FOR UPDATE;

-- name: GetPasswordResetForUpdate :one
SELECT * FROM identity.password_credentials WHERE reset_token_hash = $1 FOR UPDATE;

-- name: GetAccountSessionForUpdate :one
SELECT * FROM identity.account_sessions
WHERE id = $1 AND principal_id = $2
FOR UPDATE;

-- name: LockPrincipalAccountSessions :many
SELECT * FROM identity.account_sessions
WHERE principal_id = $1
ORDER BY id
FOR UPDATE;

-- name: CreateAccountSession :exec
INSERT INTO identity.account_sessions
  (id, principal_id, state, state_version, client_signing_public_key, access_token_hash,
   access_expires_at, absolute_expires_at, created_at, updated_at)
VALUES ($1, $2, 'active', 1, $3, $4, $5, $6, $7, $7);

-- name: DiscoverRefreshToken :one
SELECT r.token_hash, r.session_id, s.principal_id
FROM identity.account_refresh_tokens r
JOIN identity.account_sessions s ON s.id=r.session_id
WHERE r.token_hash=$1;

-- name: LockSessionRefreshTokens :many
SELECT r.*
FROM identity.account_refresh_tokens r
WHERE r.session_id=$1
ORDER BY r.token_hash
FOR UPDATE OF r;

-- name: LockPrincipalRefreshTokens :many
SELECT r.*
FROM identity.account_refresh_tokens r
JOIN identity.account_sessions s ON s.id=r.session_id
WHERE s.principal_id=$1
ORDER BY r.token_hash
FOR UPDATE OF r;

-- name: ListSessionRefreshTokens :many
SELECT r.*
FROM identity.account_refresh_tokens r
WHERE r.session_id=$1
ORDER BY r.token_hash;

-- name: ListPrincipalRefreshTokens :many
SELECT r.*
FROM identity.account_refresh_tokens r
JOIN identity.account_sessions s ON s.id=r.session_id
WHERE s.principal_id=$1
ORDER BY r.token_hash;

-- name: MarkAccountRefreshUsed :one
UPDATE identity.account_refresh_tokens
SET state = 'used', used_at = $2
WHERE token_hash = $1 AND state = 'active'
RETURNING *;

-- name: RevokeAccountSession :execrows
UPDATE identity.account_sessions
SET state = 'revoked', state_version = state_version + 1, updated_at = $2
WHERE id = $1 AND state IN ('active','review_required');

-- name: RevokePrincipalAccountSessions :many
WITH revoked_refresh AS (
  UPDATE identity.account_refresh_tokens r
  SET state='revoked', revoked_at=sqlc.arg(revoked_at)
  WHERE r.token_hash=ANY(sqlc.arg(token_hashes)::bytea[])
    AND r.state='active'
  RETURNING r.session_id
)
UPDATE identity.account_sessions s
SET state='revoked', state_version=state_version+1, updated_at=sqlc.arg(revoked_at)
WHERE s.id=ANY(sqlc.arg(session_ids)::uuid[])
  AND s.state IN ('active','review_required')
  AND (SELECT count(*) FROM revoked_refresh) >= 0
RETURNING s.id;

-- name: AcceptTOTPStep :execrows
UPDATE identity.totp_credentials
SET last_accepted_step = $2
WHERE principal_id = $1 AND state = 'active'
  AND (last_accepted_step IS NULL OR last_accepted_step < $2);

-- name: CreateEmailIdentity :exec
INSERT INTO identity.email_identities
  (id, principal_id, lookup_key_version, lookup_digest, ciphertext, encryption_key_version,
   verification_token_hash, verification_expires_at, verification_delivery_id,
   verification_delivery_ciphertext, verification_delivery_key_version, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$12);

-- name: ResetEmailVerification :one
UPDATE identity.email_identities
SET verification_token_hash = $2, verification_expires_at = $3,
    verification_consumed_at = NULL, verification_delivery_id=$4,
    verification_delivery_ciphertext=$5, verification_delivery_key_version=$6, updated_at=$7
WHERE principal_id = $1 AND verified_at IS NULL
RETURNING *;

-- name: GetPendingEmailDelivery :one
SELECT verification_delivery_id, verification_delivery_ciphertext,
       verification_delivery_key_version, verification_expires_at
FROM identity.email_identities
WHERE verification_delivery_id=$1 AND verification_consumed_at IS NULL;

-- name: ClearPendingEmailDelivery :execrows
UPDATE identity.email_identities
SET verification_delivery_id=NULL, verification_delivery_ciphertext=NULL,
    verification_delivery_key_version=NULL, updated_at=$2
WHERE verification_delivery_id=$1;

-- name: CreatePasswordCredential :exec
INSERT INTO identity.password_credentials
  (principal_id, policy_version, memory_kib, time_cost, parallelism, salt, password_hash, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8);

-- name: UpdatePasswordCredential :execrows
UPDATE identity.password_credentials
SET policy_version=$2, memory_kib=$3, time_cost=$4, parallelism=$5,
    salt=$6, password_hash=$7, updated_at=$8
WHERE principal_id=$1;

-- name: SetPasswordReset :one
UPDATE identity.password_credentials
SET reset_token_hash=$2, reset_expires_at=$3, reset_consumed_at=NULL,
    reset_delivery_id=$4, reset_delivery_ciphertext=$5,
    reset_delivery_key_version=$6, updated_at=$7
WHERE principal_id=$1
RETURNING *;

-- name: GetPendingPasswordResetDelivery :one
SELECT reset_delivery_id, reset_delivery_ciphertext, reset_delivery_key_version, reset_expires_at
FROM identity.password_credentials
WHERE reset_delivery_id=$1 AND reset_consumed_at IS NULL;

-- name: ClearPendingPasswordResetDelivery :execrows
UPDATE identity.password_credentials
SET reset_delivery_id=NULL, reset_delivery_ciphertext=NULL,
    reset_delivery_key_version=NULL, updated_at=$2
WHERE reset_delivery_id=$1;

-- name: ConsumePasswordReset :one
UPDATE identity.password_credentials
SET policy_version=$3, memory_kib=$4, time_cost=$5, parallelism=$6,
    salt=$7, password_hash=$8, reset_consumed_at=$9,
    reset_delivery_id=NULL, reset_delivery_ciphertext=NULL,
    reset_delivery_key_version=NULL, updated_at=$9
WHERE principal_id=$1 AND reset_token_hash=$2
  AND reset_consumed_at IS NULL AND reset_expires_at >= $9
RETURNING *;

-- name: FindAccountAccessToken :one
SELECT s.id, s.principal_id, s.state, s.state_version, s.access_expires_at,
       s.absolute_expires_at, a.state AS account_state
FROM identity.account_sessions s
JOIN identity.accounts a ON a.id=s.principal_id
WHERE s.access_token_hash=$1;

-- name: InsertAccountRefreshToken :exec
INSERT INTO identity.account_refresh_tokens
  (token_hash, session_id, previous_token_hash, state, issued_at, idle_expires_at, absolute_expires_at)
VALUES ($1,$2,$3,'active',$4,$5,$6);

-- name: RotateAccountSessionAccess :one
UPDATE identity.account_sessions
SET access_token_hash=$2, access_expires_at=$3, updated_at=$4,
    state_version=state_version+1
WHERE id=$1 AND state='active' AND absolute_expires_at >= $4
RETURNING *;

-- name: RevokeAccountRefreshTokens :execrows
UPDATE identity.account_refresh_tokens
SET state='revoked', revoked_at=sqlc.arg(revoked_at)
WHERE token_hash=ANY(sqlc.arg(token_hashes)::bytea[]) AND state='active';

-- name: MarkAccountSessionCompromised :execrows
UPDATE identity.account_sessions
SET state='compromised', state_version=state_version+1, updated_at=$2
WHERE id=$1 AND state IN ('active','review_required');

-- name: MarkPrincipalSessionsReviewRequired :execrows
UPDATE identity.account_sessions
SET state='review_required', state_version=state_version+1, updated_at=sqlc.arg(updated_at)
WHERE id=ANY(sqlc.arg(session_ids)::uuid[]) AND state='active';

-- name: CreatePasskeyCredential :exec
INSERT INTO identity.passkey_credentials
  (credential_id, principal_id, public_key, attestation_format, transports,
   protocol_flags, sign_count, state, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,'active',$8,$8);

-- name: ListActivePasskeys :many
SELECT * FROM identity.passkey_credentials
WHERE principal_id=$1 AND state='active'
ORDER BY created_at, credential_id LIMIT 10
FOR UPDATE;

-- name: UpdatePasskeyCounter :execrows
UPDATE identity.passkey_credentials
SET sign_count=$3, protocol_flags=$4, updated_at=$5
WHERE credential_id=$1 AND principal_id=$2 AND state='active'
  AND (sign_count < $3 OR (sign_count = 0 AND $3 = 0));

-- name: RevokePasskey :execrows
UPDATE identity.passkey_credentials
SET state='revoked', revoked_at=$3, updated_at=$3
WHERE credential_id=$1 AND principal_id=$2 AND state='active';

-- name: CreateTOTPEnrollment :execrows
INSERT INTO identity.totp_credentials AS totp_factors
  (principal_id, ciphertext, encryption_key_version, state, created_at)
VALUES (
  sqlc.arg(principal_id),
  sqlc.arg(encrypted_secret),
  sqlc.arg(key_version),
  'pending',
  sqlc.arg(enrolled_at)
)
ON CONFLICT (principal_id) DO UPDATE
SET ciphertext = EXCLUDED.ciphertext,
    encryption_key_version = EXCLUDED.encryption_key_version,
    state = 'pending',
    created_at = EXCLUDED.created_at,
    verified_at = NULL,
    last_accepted_step = NULL,
    revoked_at = NULL
WHERE totp_factors.state = 'revoked';

-- name: GetTOTPForUpdate :one
SELECT * FROM identity.totp_credentials WHERE principal_id=$1 FOR UPDATE;

-- name: ActivateTOTP :execrows
UPDATE identity.totp_credentials
SET state='active', verified_at=sqlc.arg(verified_at)
WHERE principal_id=sqlc.arg(principal_id) AND state='pending';

-- name: RevokeTOTP :execrows
UPDATE identity.totp_credentials
SET state='revoked', revoked_at=$2
WHERE principal_id=$1 AND state IN ('pending','active');

-- name: CreateRecoveryCodeSet :exec
INSERT INTO identity.recovery_code_sets
  (id, principal_id, generation, code_hashes, state, created_at, updated_at)
VALUES ($1,$2,$3,$4,'active',$5,$5);

-- name: GetNextRecoveryCodeGeneration :one
SELECT (COALESCE(MAX(generation), 0) + 1)::integer
FROM identity.recovery_code_sets
WHERE principal_id=$1;

-- name: GetActiveRecoveryCodeSetForUpdate :one
SELECT * FROM identity.recovery_code_sets
WHERE principal_id=$1 AND state='active' FOR UPDATE;

-- name: ConsumeRecoveryCode :one
UPDATE identity.recovery_code_sets
SET code_hashes=array_remove(code_hashes, $2::bytea),
    state=CASE WHEN cardinality(array_remove(code_hashes, $2::bytea))=0 THEN 'exhausted' ELSE state END,
    updated_at=$3
WHERE id=$1 AND state='active' AND $2::bytea=ANY(code_hashes)
RETURNING *;

-- name: RevokeRecoveryCodeSets :execrows
UPDATE identity.recovery_code_sets
SET state=$2, updated_at=$3
WHERE principal_id=$1 AND state='active' AND $2 IN ('superseded','revoked');

-- name: InsertSecurityEvent :exec
INSERT INTO identity.security_events
  (id, principal_id, category, fingerprint, aggregate_version, occurred_at)
VALUES ($1,$2,$3,$4,$5,$6);
