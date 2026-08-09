-- name: ConsumeEnrollmentGrant :one
UPDATE deviceauth.enrollment_grants
SET state = 'consumed', consumed_at = $3, consumed_device_id = $2
WHERE token_hash = $1 AND state = 'unused' AND expires_at >= $3
RETURNING *;

-- name: GetGrantForChallenge :one
SELECT id, principal_id, account_session_id, token_hash, policy_marker, provisional_until, expires_at
FROM deviceauth.enrollment_grants
WHERE token_hash = $1 AND state = 'unused' AND expires_at >= $2;

-- name: CreateDevice :exec
INSERT INTO deviceauth.devices
  (id, principal_id, display_name_ciphertext, display_name_key_version, signing_public_key, hpke_public_key, key_version, state, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,'active',$8,$8);

-- name: GetAuthorizationForUpdate :one
SELECT * FROM deviceauth.device_authorizations WHERE id = $1 FOR UPDATE;

-- name: FindDeviceAccessToken :one
SELECT f.id AS family_id, f.authorization_id, f.state AS family_state,
       f.state_version AS family_state_version, f.access_expires_at,
       f.idle_expires_at, f.absolute_expires_at,
       a.principal_id, a.device_id, a.state AS authorization_state,
       a.state_version AS authorization_state_version,
       ac.state AS account_state
FROM deviceauth.device_token_families f
JOIN deviceauth.device_authorizations a ON a.id = f.authorization_id
JOIN identity.accounts ac ON ac.id = a.principal_id
WHERE f.access_token_hash = $1;

-- name: GetDeviceRefreshForUpdate :one
SELECT r.token_hash, r.family_id, r.previous_token_hash, r.state AS refresh_state,
       r.issued_at, r.used_at, r.revoked_at,
       f.authorization_id, f.state AS family_state, f.state_version AS family_state_version,
       f.idle_expires_at, f.absolute_expires_at
FROM deviceauth.device_refresh_tokens r
JOIN deviceauth.device_token_families f ON f.id = r.family_id
WHERE r.token_hash = $1 FOR UPDATE OF r, f;

-- name: CompromiseDeviceTokenFamily :one
UPDATE deviceauth.device_token_families
SET state = 'compromised', state_version = state_version + 1, updated_at = $2
WHERE id = $1 AND state = 'active'
RETURNING *;

-- name: CreateEnrollmentGrant :exec
INSERT INTO deviceauth.enrollment_grants
  (id, principal_id, account_session_id, token_hash, policy_marker, provisional_until, state, expires_at, created_at)
VALUES ($1,$2,$3,$4,$5,$6,'unused',$7,$8);

-- name: CreateDeviceAuthorization :exec
INSERT INTO deviceauth.device_authorizations
  (id, principal_id, device_id, state, state_version, provisional_until, created_at, updated_at)
VALUES ($1,$2,$3,$4,1,$5,$6,$6);

-- name: CreateDeviceTokenFamily :exec
INSERT INTO deviceauth.device_token_families
  (id, authorization_id, state, state_version, access_token_hash, access_expires_at,
   idle_expires_at, absolute_expires_at, created_at, updated_at)
VALUES ($1,$2,'active',1,$3,$4,$5,$6,$7,$7);

-- name: InsertDeviceRefreshToken :exec
INSERT INTO deviceauth.device_refresh_tokens
  (token_hash, family_id, previous_token_hash, state, issued_at)
VALUES ($1,$2,$3,'active',$4);

-- name: CreateDevicePolicySnapshot :exec
INSERT INTO deviceauth.device_policy_snapshots
  (authorization_id, schema_version, policy, created_at)
VALUES ($1,'device-policy-v1',$2,$3);

-- name: BindAccountSessionToAuthorization :execrows
UPDATE identity.account_sessions
SET device_authorization_id=$2, state_version=state_version+1, updated_at=$3
WHERE id=$1 AND state='active' AND device_authorization_id IS NULL;

-- name: RevokeDeviceBoundAccountSessions :execrows
UPDATE identity.account_sessions
SET state='revoked', state_version=state_version+1, updated_at=$2
WHERE device_authorization_id=$1 AND state IN ('active','review_required');

-- name: MarkDeviceRefreshUsed :one
UPDATE deviceauth.device_refresh_tokens
SET state='used', used_at=$2
WHERE token_hash=$1 AND state='active'
RETURNING *;

-- name: RotateDeviceFamilyAccess :one
UPDATE deviceauth.device_token_families
SET access_token_hash=$2, access_expires_at=$3, idle_expires_at=$4,
    state_version=state_version+1, updated_at=$5
WHERE id=$1 AND state='active' AND idle_expires_at >= $5 AND absolute_expires_at >= $5
RETURNING *;

-- name: RevokeDeviceRefreshTokens :execrows
UPDATE deviceauth.device_refresh_tokens
SET state='revoked', revoked_at=$2
WHERE family_id=$1 AND state='active';

-- name: ActivateProvisionalAuthorization :execrows
UPDATE deviceauth.device_authorizations
SET state='active', state_version=state_version+1, provisional_until=NULL, updated_at=$2
WHERE principal_id=$1 AND state='provisional' AND provisional_until >= $2;

-- name: RevokeDeviceAuthorization :execrows
UPDATE deviceauth.device_authorizations
SET state='revoked', state_version=state_version+1, updated_at=$2
WHERE id=$1 AND state IN ('provisional','active','suspended');

-- name: RevokeDeviceRecord :execrows
UPDATE deviceauth.devices
SET state='revoked', updated_at=$2
WHERE id=$1 AND state IN ('active','suspended');

-- name: GetActiveBundleAuthority :one
SELECT a.id AS authorization_id, a.principal_id, a.device_id,
       a.state AS authorization_state, a.state_version,
       d.hpke_public_key, d.key_version, p.schema_version, p.policy,
       ac.state AS account_state
FROM deviceauth.device_authorizations a
JOIN deviceauth.devices d ON d.id=a.device_id
JOIN deviceauth.device_policy_snapshots p ON p.authorization_id=a.id
JOIN identity.accounts ac ON ac.id=a.principal_id
WHERE a.id=$1 AND d.state='active'
  AND (a.state='active' OR (a.state='provisional' AND a.provisional_until >= $2));
