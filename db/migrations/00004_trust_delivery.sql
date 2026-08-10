-- +goose Up
CREATE SCHEMA trust;

CREATE TABLE trust.trust_root_metadata (
  version bigint PRIMARY KEY CHECK (version > 0),
  canonical_payload bytea NOT NULL CHECK (octet_length(canonical_payload) BETWEEN 2 AND 65536),
  signature bytea NOT NULL CHECK (octet_length(signature) = 64),
  valid_from timestamptz NOT NULL,
  valid_until timestamptz NOT NULL,
  created_at timestamptz NOT NULL,
  CHECK (valid_until > valid_from)
);

CREATE TABLE trust.signing_key_metadata (
  key_id text PRIMARY KEY CHECK (key_id ~ '^[A-Za-z0-9_-]{16,64}$'),
  root_metadata_version bigint NOT NULL REFERENCES trust.trust_root_metadata(version),
  algorithm text NOT NULL CHECK (algorithm = 'Ed25519'),
  public_key bytea NOT NULL CHECK (octet_length(public_key) = 32),
  state text NOT NULL CHECK (state IN ('future','active','retiring','revoked')),
  not_before timestamptz NOT NULL,
  not_after timestamptz NOT NULL,
  CHECK (not_after > not_before)
);

CREATE TABLE trust.highest_bundle_versions (
  authorization_id uuid PRIMARY KEY REFERENCES deviceauth.device_authorizations(id),
  highest_issued bigint NOT NULL CHECK (highest_issued > 0),
  updated_at timestamptz NOT NULL
);

CREATE TABLE trust.bundle_issuances (
  id uuid PRIMARY KEY,
  authorization_id uuid NOT NULL REFERENCES deviceauth.device_authorizations(id),
  bundle_version bigint NOT NULL CHECK (bundle_version > 0),
  locator_hash bytea NOT NULL UNIQUE CHECK (octet_length(locator_hash) = 32),
  locator_ciphertext bytea NOT NULL CHECK (octet_length(locator_ciphertext) BETWEEN 29 AND 512),
  locator_key_version integer NOT NULL CHECK (locator_key_version > 0),
  envelope bytea NOT NULL CHECK (octet_length(envelope) BETWEEN 64 AND 1048576),
  envelope_sha256 bytea NOT NULL CHECK (octet_length(envelope_sha256) = 32),
  signer_key_id text NOT NULL REFERENCES trust.signing_key_metadata(key_id),
  issued_at timestamptz NOT NULL,
  not_before timestamptz NOT NULL,
  expires_at timestamptz NOT NULL,
  UNIQUE (authorization_id, bundle_version),
  CHECK (expires_at > not_before)
);

CREATE TABLE trust.bundle_acknowledgements (
  bundle_id uuid NOT NULL REFERENCES trust.bundle_issuances(id),
  authorization_id uuid NOT NULL REFERENCES deviceauth.device_authorizations(id),
  bundle_version bigint NOT NULL CHECK (bundle_version > 0),
  acknowledged_at timestamptz NOT NULL,
  PRIMARY KEY (bundle_id, authorization_id)
);

CREATE TABLE idempotency_records (
  principal_scope text NOT NULL CHECK (principal_scope ~ '^[A-Za-z0-9:_-]{1,128}$'),
  operation text NOT NULL CHECK (operation ~ '^[a-z][a-z0-9_]{0,63}$'),
  idempotency_key_hash bytea NOT NULL CHECK (octet_length(idempotency_key_hash) = 32),
  request_digest bytea NOT NULL CHECK (octet_length(request_digest) = 32),
  state text NOT NULL CHECK (state IN ('in_progress','completed','failed')),
  response_status integer,
  response_ciphertext bytea CHECK (response_ciphertext IS NULL OR octet_length(response_ciphertext) <= 1048608),
  response_key_version integer CHECK (response_key_version IS NULL OR response_key_version > 0),
  created_at timestamptz NOT NULL,
  expires_at timestamptz NOT NULL,
  CHECK ((response_ciphertext IS NULL) = (response_key_version IS NULL)),
  PRIMARY KEY (principal_scope, operation, idempotency_key_hash)
);

CREATE TABLE transactional_outbox (
  event_id uuid PRIMARY KEY,
  event_type text NOT NULL CHECK (event_type ~ '^talenro[.][a-z0-9_.-]{1,127}$'),
  aggregate_type text NOT NULL CHECK (aggregate_type ~ '^[a-z][a-z0-9_]{0,63}$'),
  aggregate_id uuid NOT NULL,
  aggregate_version bigint NOT NULL CHECK (aggregate_version > 0),
  idempotency_key text NOT NULL CHECK (idempotency_key ~ '^[A-Za-z0-9:_-]{1,128}$'),
  payload bytea NOT NULL CHECK (octet_length(payload) BETWEEN 1 AND 262144),
  occurred_at timestamptz NOT NULL,
  available_at timestamptz NOT NULL,
  claimed_until timestamptz,
  attempts integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
  published_at timestamptz
);
CREATE INDEX outbox_available_idx ON transactional_outbox(available_at, occurred_at) WHERE published_at IS NULL;

CREATE TABLE consumed_event_ids (
  consumer text NOT NULL CHECK (consumer ~ '^[a-z][a-z0-9_.-]{0,127}$'),
  event_id uuid NOT NULL,
  consumed_at timestamptz NOT NULL,
  expires_at timestamptz NOT NULL,
  PRIMARY KEY (consumer, event_id)
);

-- +goose Down
DROP TABLE consumed_event_ids;
DROP TABLE transactional_outbox;
DROP TABLE idempotency_records;
DROP SCHEMA trust CASCADE;
