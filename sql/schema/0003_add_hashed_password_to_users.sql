-- +goose Up
ALTER TABLE users
ADD COLUMN hashed_password TEXT NOT NULL DEFAULT 'unset';

-- +goose Down
ALTER TABLE users
ALTER COLUMN hashed_password DROP DEFAULT,
DROP COLUMN hashed_password;
