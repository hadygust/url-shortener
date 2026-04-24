package migration

var schema = `
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email text NOT NULL,
    password text NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE urls (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    short_code text NOT NULL,
		original_url text NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		expires_at TIMESTAMPTZ,
		CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE redirect_logs (
	id UUID PRIMARY KEY,
	url_id UUID NOT NULL,
	ip_address text NOT NULL,
	user_agent text NOT NULL,
	accessed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	CONSTRAINT fk_url FOREIGN KEY (url_id) REFERENCES urls(id) ON UPDATE CASCADE ON DELETE CASCADE
);`

func GetSchema() string {
	return schema
}
