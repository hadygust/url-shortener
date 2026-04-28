-- +goose Up
CREATE TABLE IF NOT EXISTS redirect_logs (
	id UUID PRIMARY KEY,
	url_id UUID NOT NULL,
	ip_address text NOT NULL,
	user_agent text NOT NULL,
	accessed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	CONSTRAINT fk_url FOREIGN KEY (url_id) REFERENCES urls(id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- +goose Down
DROP TABLE redirect_logs;
