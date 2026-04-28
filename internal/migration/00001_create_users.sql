-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email text NOT NULL,
    password text NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT unique_email UNIQUE(email)
);

-- +goose Down
DROP TABLE users;
