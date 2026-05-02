package url

import (
	"log"

	"github.com/hadygust/url-shortener/internal/model"
	"github.com/jmoiron/sqlx"
)

type Repository interface {
	CreateUrl(model.Url) (model.Url, error)
	GetAllUserUrl(string) ([]model.Url, error)
	GetUrlbyShortCode(shortCode string) (model.Url, error)
	DeleteUrl(string, string) (model.Url, error)
}

func (repo *urlRepository) CreateUrl(url model.Url) (model.Url, error) {
	query := `
		INSERT INTO urls (id, user_id, short_code, original_url, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING *
	`
	var resUrl model.Url
	err := repo.db.Get(&resUrl, query, url.ID, url.UserId, url.ShortCode, url.OriginalUrl, url.ExpiresAt)
	if err != nil {
		return model.Url{}, err
	}

	return resUrl, nil
}

func (repo *urlRepository) GetAllUserUrl(userId string) ([]model.Url, error) {
	query := `
		SELECT *
		FROM urls
		WHERE user_id = $1
	`
	urls := []model.Url{}

	err := repo.db.Select(&urls, query, userId)
	if err != nil {
		return []model.Url{}, err
	}

	return urls, nil
}

func (repo *urlRepository) GetUrlbyShortCode(shortCode string) (model.Url, error) {

	query := `
		SELECT * 
		FROM urls
		WHERE short_code = $1
	`

	url := model.Url{}
	err := repo.db.Get(&url, query, shortCode)
	if err != nil {
		return model.Url{}, err
	}

	return url, nil
}

func (repo *urlRepository) DeleteUrl(shortCode string, userId string) (model.Url, error) {
	query := `
		DELETE FROM urls
		WHERE short_code = $1 AND user_id = $2
		RETURNING *;
	`

	deletedUrl := model.Url{}

	log.Println(shortCode)
	err := repo.db.Get(&deletedUrl, query, shortCode, userId)
	if err != nil {

		return model.Url{}, err
	}

	return deletedUrl, nil
}

type urlRepository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *urlRepository {
	return &urlRepository{
		db: db,
	}
}

/*
INSERT INTO redirect_logs (id, url_id, ip_address, user_agent, accessed_at)
SELECT
    gen_random_uuid(),
    '8fd023fc-efe3-4093-a647-38e0e0de22dc',

    -- repeatable IP patterns (not fully random)
    '192.168.' || (gs % 50) || '.' || (gs % 255),

    -- weighted user agent distribution (realistic dominance)
    CASE
        WHEN random() < 0.4 THEN 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)'
        WHEN random() < 0.7 THEN 'Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)'
        WHEN random() < 0.9 THEN 'Mozilla/5.0 (Linux; Android 10)'
        ELSE 'curl/7.68.0'
    END,

    -- skewed toward recent (more realistic traffic pattern)
    NOW() - (random()^2 * interval '14 days')

FROM generate_series(1, 1000000) AS gs;
*/
