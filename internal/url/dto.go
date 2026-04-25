package url

import "time"

type UrlResponse struct {
	ID          string     `json:"id"`
	ShortCode   string     `json:"short_code"`
	OriginalUrl string     `json:"original_url"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

type CreateUrlRequest struct {
	OriginalUrl string     `json:"original_url" binding:"required,url"`
	ShortCode   string     `json:"short_code,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}
