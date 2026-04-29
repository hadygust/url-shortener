package cache

import (
	"time"

	"github.com/hadygust/url-shortener/internal/dto"
	"github.com/hadygust/url-shortener/internal/model"
)

type Cache interface {
	Get(string) (any, error)
	Set(string, any, time.Duration) error
	Delete(string) error
	GetUrl(shortCode string) (*dto.UrlCache, error)
	SetUrl(url model.Url) error
}
