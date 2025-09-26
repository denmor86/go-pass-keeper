-- +goose Up
-- +goose StatementBegin
ALTER TABLE secrets
ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE secrets
DROP COLUMN IF EXISTS updated_at;
-- +goose StatementEnd
