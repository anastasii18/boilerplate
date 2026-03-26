-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
                       user_uuid UUID PRIMARY KEY,
                       login VARCHAR(255) NOT NULL UNIQUE,
                       email VARCHAR(255) NOT NULL UNIQUE,
                       hashed_password VARCHAR(255) NOT NULL,
                       notification_methods JSONB NOT NULL DEFAULT '{}'::jsonb
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user;
-- +goose StatementEnd
