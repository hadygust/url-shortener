package auth

import (
	"github.com/google/uuid"
	"github.com/hadygust/url-shortener/internal/model"
)

type LoginResponse struct {
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

func NewUserResponse(user model.User) *UserResponse {
	return &UserResponse{
		ID:    user.ID,
		Email: user.Email,
	}
}
