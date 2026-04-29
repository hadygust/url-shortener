package dto

import (
	"time"

	"github.com/hadygust/url-shortener/internal/model"
)

type UrlResponse struct {
	ID          string     `json:"id"`
	ShortCode   string     `json:"short_code"`
	OriginalUrl string     `json:"original_url"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

type CreateUrlRequest struct {
	OriginalUrl string     `json:"original_url" binding:"required,url"`
	ShortCode   string     `json:"short_code"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

type UrlCache struct {
	ID          string `json:"id"`
	OriginalUrl string `json:"original_url"`
}

func NewUrlResponse(url model.Url) *UrlResponse {
	return &UrlResponse{
		ID:          url.ID.String(),
		ShortCode:   url.ShortCode,
		OriginalUrl: url.OriginalUrl,
		CreatedAt:   url.CreatedAt.Time,
		ExpiresAt:   &url.ExpiresAt.Time,
	}
}
