-- name: NextBundleVersion :one
INSERT INTO trust.highest_bundle_versions (authorization_id, highest_issued, updated_at)
VALUES ($1, 1, $2)
ON CONFLICT (authorization_id) DO UPDATE
SET highest_issued = trust.highest_bundle_versions.highest_issued + 1,
    updated_at = EXCLUDED.updated_at
RETURNING highest_issued;

-- name: GetHighestBundleVersion :one
SELECT highest_issued FROM trust.highest_bundle_versions WHERE authorization_id=$1;

-- name: GetBundleByLocatorHash :one
SELECT * FROM trust.bundle_issuances
WHERE locator_hash = $1 AND expires_at >= $2;

-- name: InsertBundleAcknowledgement :exec
INSERT INTO trust.bundle_acknowledgements
  (bundle_id, authorization_id, bundle_version, acknowledged_at)
VALUES ($1,$2,$3,$4)
ON CONFLICT (bundle_id, authorization_id) DO NOTHING;

-- name: InsertTrustRootMetadata :exec
INSERT INTO trust.trust_root_metadata
  (version, canonical_payload, signature, valid_from, valid_until, created_at)
VALUES ($1,$2,$3,$4,$5,$6);

-- name: ListTrustRootMetadata :many
SELECT * FROM trust.trust_root_metadata ORDER BY version;

-- name: InsertSigningKeyMetadata :exec
INSERT INTO trust.signing_key_metadata
  (key_id, root_metadata_version, algorithm, public_key, state, not_before, not_after)
VALUES ($1,$2,'Ed25519',$3,$4,$5,$6);

-- name: ListSigningKeysForMetadata :many
SELECT * FROM trust.signing_key_metadata
WHERE root_metadata_version=$1 ORDER BY key_id;

-- name: InsertBundleIssuance :exec
INSERT INTO trust.bundle_issuances
  (id, authorization_id, bundle_version, locator_hash, locator_ciphertext,
   locator_key_version, envelope, envelope_sha256, signer_key_id,
   issued_at, not_before, expires_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12);

-- name: GetBundleIssuance :one
SELECT * FROM trust.bundle_issuances WHERE id=$1;
