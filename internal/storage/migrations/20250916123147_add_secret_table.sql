-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE TABLE IF NOT EXISTS secrets
(
    id          UUID                  DEFAULT uuid_generate_v4() NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    user_id     UUID         NOT NULL,
    type_secret VARCHAR(255) NOT NULL,
    name        VARCHAR(255) NOT NULL,
    content     BYTEA        NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT foreign_key_user FOREIGN KEY (user_id) REFERENCES users (id)
);
CREATE INDEX IF NOT EXISTS idx_secrets_user_id ON secrets (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS secrets;
-- +goose StatementEnd
