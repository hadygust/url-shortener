package auth

import (
	"github.com/hadygust/url-shortener/internal/model"
	"github.com/jmoiron/sqlx"
)

type Repository interface {
	registerUser(model.User) (model.User, error)
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

type userRepository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &userRepository{
		db: db,
	}
}
