-- +goose Up
CREATE SCHEMA identity;

CREATE TABLE identity.accounts (
  id uuid PRIMARY KEY,
  state text NOT NULL CHECK (state IN ('pending_email','active','suspended','deletion_pending','deleted')),
  state_version bigint NOT NULL CHECK (state_version > 0),
  locale text NOT NULL CHECK (locale ~ '^[A-Za-z]{2,3}(-[A-Za-z0-9]{2,8})*$'),
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CHECK (updated_at >= created_at)
);

CREATE TABLE identity.email_identities (
  id uuid PRIMARY KEY,
  principal_id uuid NOT NULL UNIQUE REFERENCES identity.accounts(id),
  lookup_key_version integer NOT NULL CHECK (lookup_key_version > 0),
  lookup_digest bytea NOT NULL UNIQUE CHECK (octet_length(lookup_digest) = 32),
  ciphertext bytea NOT NULL CHECK (octet_length(ciphertext) BETWEEN 29 AND 2048),
  encryption_key_version integer NOT NULL CHECK (encryption_key_version > 0),
  verification_token_hash bytea CHECK (verification_token_hash IS NULL OR octet_length(verification_token_hash) = 32),
  verification_expires_at timestamptz,
  verification_consumed_at timestamptz,
  verification_delivery_id uuid UNIQUE,
  verification_delivery_ciphertext bytea CHECK (verification_delivery_ciphertext IS NULL OR octet_length(verification_delivery_ciphertext) BETWEEN 29 AND 4096),
  verification_delivery_key_version integer CHECK (verification_delivery_key_version IS NULL OR verification_delivery_key_version > 0),
  verified_at timestamptz,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CHECK ((verification_delivery_id IS NULL) = (verification_delivery_ciphertext IS NULL)),
  CHECK ((verification_delivery_id IS NULL) = (verification_delivery_key_version IS NULL))
);

CREATE TABLE identity.password_credentials (
  principal_id uuid PRIMARY KEY REFERENCES identity.accounts(id),
  policy_version integer NOT NULL CHECK (policy_version > 0),
  memory_kib integer NOT NULL CHECK (memory_kib >= 65536),
  time_cost integer NOT NULL CHECK (time_cost >= 3),
  parallelism integer NOT NULL CHECK (parallelism >= 1 AND parallelism <= 16),
  salt bytea NOT NULL CHECK (octet_length(salt) = 16),
  password_hash bytea NOT NULL CHECK (octet_length(password_hash) = 32),
  reset_token_hash bytea CHECK (reset_token_hash IS NULL OR octet_length(reset_token_hash) = 32),
  reset_expires_at timestamptz,
  reset_consumed_at timestamptz,
  reset_delivery_id uuid UNIQUE,
  reset_delivery_ciphertext bytea CHECK (reset_delivery_ciphertext IS NULL OR octet_length(reset_delivery_ciphertext) BETWEEN 29 AND 4096),
  reset_delivery_key_version integer CHECK (reset_delivery_key_version IS NULL OR reset_delivery_key_version > 0),
  updated_at timestamptz NOT NULL,
  CHECK ((reset_delivery_id IS NULL) = (reset_delivery_ciphertext IS NULL)),
  CHECK ((reset_delivery_id IS NULL) = (reset_delivery_key_version IS NULL))
);

CREATE TABLE identity.account_sessions (
  id uuid PRIMARY KEY,
  principal_id uuid NOT NULL REFERENCES identity.accounts(id),
  state text NOT NULL CHECK (state IN ('active','review_required','revoked','compromised')),
  state_version bigint NOT NULL CHECK (state_version > 0),
  client_signing_public_key bytea NOT NULL CHECK (octet_length(client_signing_public_key) = 32),
  access_token_hash bytea NOT NULL UNIQUE CHECK (octet_length(access_token_hash) = 32),
  access_expires_at timestamptz NOT NULL,
  absolute_expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE TABLE identity.account_refresh_tokens (
  token_hash bytea PRIMARY KEY CHECK (octet_length(token_hash) = 32),
  session_id uuid NOT NULL REFERENCES identity.account_sessions(id),
  previous_token_hash bytea CHECK (previous_token_hash IS NULL OR octet_length(previous_token_hash) = 32),
  state text NOT NULL CHECK (state IN ('active','used','revoked')),
  issued_at timestamptz NOT NULL,
  idle_expires_at timestamptz NOT NULL,
  absolute_expires_at timestamptz NOT NULL,
  used_at timestamptz,
  revoked_at timestamptz
);
CREATE INDEX account_refresh_tokens_session_idx ON identity.account_refresh_tokens(session_id);

CREATE TABLE identity.passkey_credentials (
  credential_id bytea PRIMARY KEY CHECK (octet_length(credential_id) BETWEEN 16 AND 1024),
  principal_id uuid NOT NULL REFERENCES identity.accounts(id),
  public_key bytea NOT NULL CHECK (octet_length(public_key) BETWEEN 32 AND 4096),
  attestation_format text NOT NULL CHECK (length(attestation_format) BETWEEN 1 AND 64),
  transports text[] NOT NULL DEFAULT '{}',
  protocol_flags smallint NOT NULL CHECK (protocol_flags BETWEEN 0 AND 255),
  sign_count bigint NOT NULL CHECK (sign_count >= 0),
  state text NOT NULL CHECK (state IN ('active','revoked')),
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  revoked_at timestamptz
);
CREATE INDEX passkey_credentials_principal_idx ON identity.passkey_credentials(principal_id);

CREATE TABLE identity.totp_credentials (
  principal_id uuid PRIMARY KEY REFERENCES identity.accounts(id),
  ciphertext bytea NOT NULL CHECK (octet_length(ciphertext) BETWEEN 29 AND 1024),
  encryption_key_version integer NOT NULL CHECK (encryption_key_version > 0),
  state text NOT NULL CHECK (state IN ('pending','active','revoked')),
  last_accepted_step bigint,
  created_at timestamptz NOT NULL,
  verified_at timestamptz,
  revoked_at timestamptz
);

CREATE TABLE identity.recovery_code_sets (
  id uuid PRIMARY KEY,
  principal_id uuid NOT NULL REFERENCES identity.accounts(id),
  generation integer NOT NULL CHECK (generation > 0),
  code_hashes bytea[] NOT NULL CHECK (cardinality(code_hashes) BETWEEN 1 AND 10),
  state text NOT NULL CHECK (state IN ('active','superseded','exhausted','revoked')),
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  UNIQUE (principal_id, generation)
);

CREATE TABLE identity.security_events (
  id uuid PRIMARY KEY,
  principal_id uuid REFERENCES identity.accounts(id),
  category text NOT NULL CHECK (category ~ '^[a-z][a-z0-9_]{0,63}$'),
  fingerprint text NOT NULL CHECK (fingerprint ~ '^[a-z][a-z0-9_.-]{0,127}$'),
  aggregate_version bigint NOT NULL CHECK (aggregate_version > 0),
  occurred_at timestamptz NOT NULL
);
CREATE INDEX security_events_principal_time_idx ON identity.security_events(principal_id, occurred_at DESC);

-- +goose Down
DROP SCHEMA identity CASCADE;
