-- +goose Up
CREATE TABLE system_metadata (
  key text PRIMARY KEY,
  value jsonb NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT system_metadata_key_format CHECK (key ~ '^[a-z][a-z0-9_.-]{0,127}$')
);

-- +goose Down
DROP TABLE system_metadata;
