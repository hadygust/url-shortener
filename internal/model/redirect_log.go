package model

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type RedirectLog struct {
	ID         uuid.UUID          `json:"id" db:"id"`
	UrlId      uuid.UUID          `json:"url_id" db:"url_id"`
	IpAddress  string             `json:"ip_address" db:"ip_address"`
	UserAgent  string             `json:"user_agent" db:"user_agent"`
	AccessedAt pgtype.Timestamptz `json:"accessed_at" db:"accessed_at"`
}
