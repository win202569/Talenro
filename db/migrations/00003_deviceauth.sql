-- +goose Up
CREATE SCHEMA deviceauth;

CREATE TABLE deviceauth.enrollment_grants (
  id uuid PRIMARY KEY,
  principal_id uuid NOT NULL REFERENCES identity.accounts(id),
  account_session_id uuid NOT NULL REFERENCES identity.account_sessions(id),
  token_hash bytea NOT NULL UNIQUE CHECK (octet_length(token_hash) = 32),
  policy_marker text NOT NULL CHECK (policy_marker IN ('standard','trial_restricted')),
  provisional_until timestamptz,
  state text NOT NULL CHECK (state IN ('unused','consumed','expired')),
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL,
  consumed_at timestamptz,
  consumed_device_id uuid,
  CHECK ((policy_marker='trial_restricted') = (provisional_until IS NOT NULL))
);

CREATE TABLE deviceauth.devices (
  id uuid PRIMARY KEY,
  principal_id uuid NOT NULL REFERENCES identity.accounts(id),
  display_name_ciphertext bytea CHECK (display_name_ciphertext IS NULL OR octet_length(display_name_ciphertext) BETWEEN 29 AND 1024),
  display_name_key_version integer,
  signing_public_key bytea NOT NULL CHECK (octet_length(signing_public_key) = 32),
  hpke_public_key bytea NOT NULL CHECK (octet_length(hpke_public_key) = 32),
  key_version integer NOT NULL CHECK (key_version > 0),
  state text NOT NULL CHECK (state IN ('active','suspended','revoked')),
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  UNIQUE (principal_id, signing_public_key),
  UNIQUE (principal_id, hpke_public_key)
);
ALTER TABLE deviceauth.enrollment_grants
  ADD CONSTRAINT enrollment_grants_consumed_device_fk
  FOREIGN KEY (consumed_device_id) REFERENCES deviceauth.devices(id);

CREATE TABLE deviceauth.device_authorizations (
  id uuid PRIMARY KEY,
  principal_id uuid NOT NULL REFERENCES identity.accounts(id),
  device_id uuid NOT NULL UNIQUE REFERENCES deviceauth.devices(id),
  state text NOT NULL CHECK (state IN ('provisional','active','suspended','revoked')),
  state_version bigint NOT NULL CHECK (state_version > 0),
  provisional_until timestamptz,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);
CREATE UNIQUE INDEX one_provisional_authorization_per_principal
  ON deviceauth.device_authorizations(principal_id) WHERE state='provisional';

ALTER TABLE identity.account_sessions
  ADD COLUMN device_authorization_id uuid REFERENCES deviceauth.device_authorizations(id);
CREATE INDEX account_sessions_device_authorization_idx
  ON identity.account_sessions(device_authorization_id);

CREATE TABLE deviceauth.device_token_families (
  id uuid PRIMARY KEY,
  authorization_id uuid NOT NULL REFERENCES deviceauth.device_authorizations(id),
  state text NOT NULL CHECK (state IN ('active','compromised','revoked')),
  state_version bigint NOT NULL CHECK (state_version > 0),
  access_token_hash bytea NOT NULL UNIQUE CHECK (octet_length(access_token_hash) = 32),
  access_expires_at timestamptz NOT NULL,
  idle_expires_at timestamptz NOT NULL,
  absolute_expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE TABLE deviceauth.device_refresh_tokens (
  token_hash bytea PRIMARY KEY CHECK (octet_length(token_hash) = 32),
  family_id uuid NOT NULL REFERENCES deviceauth.device_token_families(id),
  previous_token_hash bytea CHECK (previous_token_hash IS NULL OR octet_length(previous_token_hash) = 32),
  state text NOT NULL CHECK (state IN ('active','used','revoked')),
  issued_at timestamptz NOT NULL,
  used_at timestamptz,
  revoked_at timestamptz
);
CREATE INDEX device_refresh_tokens_family_idx ON deviceauth.device_refresh_tokens(family_id);

CREATE TABLE deviceauth.device_policy_snapshots (
  authorization_id uuid PRIMARY KEY REFERENCES deviceauth.device_authorizations(id),
  schema_version text NOT NULL CHECK (schema_version = 'device-policy-v1'),
  policy jsonb NOT NULL,
  created_at timestamptz NOT NULL,
  CHECK (octet_length(policy::text) <= 4096)
);

-- +goose Down
ALTER TABLE identity.account_sessions DROP COLUMN device_authorization_id;
DROP SCHEMA deviceauth CASCADE;
