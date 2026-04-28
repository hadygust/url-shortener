-- +goose Up
CREATE TABLE IF NOT EXISTS urls (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    short_code text NOT NULL,
		original_url text NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		expires_at TIMESTAMPTZ,
		CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT short_code_unique UNIQUE(short_code)
);

-- +goose Down
DROP TABLE urls;
