package model

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Url struct {
	ID          uuid.UUID          `json:"id"`
	UserId      uuid.UUID          `json:"user_id"`
	ShortCode   string             `json:"short_code"`
	OriginalUrl string             `json:"original_url"`
	CreatedAt   pgtype.Timestamptz `json:"created_at"`
	ExpiresAt   pgtype.Timestamptz `json:"expires_at"`
}
