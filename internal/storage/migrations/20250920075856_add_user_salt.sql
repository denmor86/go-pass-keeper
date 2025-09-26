-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
ADD salt TEXT DEFAULT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
DROP COLUMN salt;
-- +goose StatementEnd
