package url

import (
	"github.com/hadygust/url-shortener/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Repository interface {
	CreateUrl(model.Url) (model.Url, error)
}

func (repo *urlRepository) CreateUrl(url model.Url) (model.Url, error) {
	query := `
		INSERT INTO urls (id, user_id, short_code, original_url, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING *
	`
	var resUrl model.Url
	err := repo.db.Get(&resUrl, query, url.ID, url.UserId, url.ShortCode, url.OriginalUrl, url.ExpiresAt)
	if err != nil {
		return model.Url{}, err
	}

	return resUrl, nil
}

type urlRepository struct {
	db    *sqlx.DB
	redis *redis.Client
}

func NewRepository(db *sqlx.DB, redis *redis.Client) *urlRepository {
	return &urlRepository{
		db:    db,
		redis: redis,
	}
}
