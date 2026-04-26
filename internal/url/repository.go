package url

import (
	"log"

	"github.com/hadygust/url-shortener/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Repository interface {
	CreateUrl(model.Url) (model.Url, error)
	GetAllUserUrl(string) ([]model.Url, error)
	GetOrigin(string) (string, error)
	DeleteUrl(string) (model.Url, error)
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

func (repo *urlRepository) GetOrigin(shortCode string) (string, error) {

	query := `
		SELECT * 
		FROM urls
		WHERE short_code = $1
	`

	url := model.Url{}
	err := repo.db.Get(&url, query, shortCode)
	if err != nil {
		return "", err
	}

	return url.OriginalUrl, nil
}

func (repo *urlRepository) DeleteUrl(shortCode string) (model.Url, error) {
	query := `
		DELETE FROM urls
		WHERE short_code = $1
		RETURNING *;
	`

	deletedUrl := model.Url{}

	log.Println(shortCode)
	err := repo.db.Get(&deletedUrl, query, shortCode)
	if err != nil {
		// if errors.Is(err, sql.ErrNoRows) {
		// 	err = errors.New("url didnt exists")
		// }
		return model.Url{}, err
	}

	return deletedUrl, nil
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
