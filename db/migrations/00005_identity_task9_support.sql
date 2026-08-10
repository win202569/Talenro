-- +goose Up
CREATE UNIQUE INDEX identity_email_verification_token_hash_unique
  ON identity.email_identities (verification_token_hash)
  WHERE verification_token_hash IS NOT NULL;

CREATE UNIQUE INDEX identity_password_reset_token_hash_unique
  ON identity.password_credentials (reset_token_hash)
  WHERE reset_token_hash IS NOT NULL;

-- +goose Down
DROP INDEX identity.identity_password_reset_token_hash_unique;
DROP INDEX identity.identity_email_verification_token_hash_unique;
