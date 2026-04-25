package auth

import (
	"context"
	"time"

	"github.com/hadygust/url-shortener/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Repository interface {
	registerUser(model.User) (model.User, error)
	loginUser(LoginRequest) (model.User, error)
	blacklistToken(string, time.Time) error
	getUserByID(string) (model.User, error)
	checkBlacklistToken(string) bool
}

func (repo *userRepository) registerUser(newUser model.User) (model.User, error) {

	query := `
		INSERT INTO users (id, email, password)
		VALUES ($1, $2, $3)
		RETURNING *
	`
	user := model.User{}
	err := repo.db.Get(&user, query, newUser.ID, newUser.Email, newUser.Password)

	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

func (repo *userRepository) loginUser(login LoginRequest) (model.User, error) {
	user := model.User{}
	err := repo.db.Get(&user, "SELECT * FROM users WHERE email=$1", login.Email)

	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

func (repo *userRepository) blacklistToken(jti string, exp time.Time) error {
	repo.redis.Set(context.Background(), "blacklist:"+jti, "1", time.Until(exp))
	return nil
}

func (repo *userRepository) checkBlacklistToken(jti string) bool {
	_, err := repo.redis.Get(context.Background(), "blacklist:"+jti).Result()

	if err == nil {
		return false
	}

	return true
}

func (repo *userRepository) getUserByID(id string) (model.User, error) {
	user := model.User{}
	err := repo.db.Get(&user, "SELECT * FROM users WHERE id=$1", id)

	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

type userRepository struct {
	db    *sqlx.DB
	redis *redis.Client
}

func NewRepository(db *sqlx.DB, redis *redis.Client) Repository {
	return &userRepository{
		db:    db,
		redis: redis,
	}
}
