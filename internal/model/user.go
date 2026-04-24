package model

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// users
// ├── id (UUID)
// ├── email (unique)
// ├── password_hash
// └── created_at
type User struct {
	ID         uuid.UUID
	Email      string
	Password   string
	Created_at pgtype.Timestamptz
}
