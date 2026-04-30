package cache

import (
	"time"

	"github.com/hadygust/url-shortener/internal/dto"
	"github.com/hadygust/url-shortener/internal/model"
)

// RequestCounter stores request timestamps for rate limiting
type RequestCounter struct {
	Timestamps []int64 `json:"timestamps"`
}

type Cache interface {
	// Generic cache operations
	Get(string) (any, error)
	Set(string, any, time.Duration) error
	Delete(string) error

	// URL caching operations
	GetUrl(shortCode string) (*dto.UrlCache, error)
	SetUrl(url model.Url) error

	// Rate limit operations - URL POST (creation)
	GetUrlPostLimit(userID string) (*RequestCounter, error)
	SetUrlPostLimit(userID string, counter *RequestCounter, ttl time.Duration) error

	// Rate limit operations - URL GET (access)
	GetUrlGetLimit(key string) (*RequestCounter, error)
	SetUrlGetLimit(key string, counter *RequestCounter, ttl time.Duration) error

	// Rate limit operations - IP-based limits
	GetIPRateLimit(ip string) (*RequestCounter, error)
	SetIPRateLimit(ip string, counter *RequestCounter, ttl time.Duration) error

	// Rate limit operations - Redirect tracking
	GetRedirectLimit(urlID string) (*RequestCounter, error)
	SetRedirectLimit(urlID string, counter *RequestCounter, ttl time.Duration) error
}
