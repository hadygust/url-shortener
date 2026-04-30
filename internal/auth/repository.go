package auth

import (
	"github.com/hadygust/url-shortener/internal/dto"
	"github.com/hadygust/url-shortener/internal/model"
	"github.com/jmoiron/sqlx"
)

type Repository interface {
	RegisterUser(model.User) (model.User, error)
	LoginUser(dto.LoginRequest) (model.User, error)
	GetUserByID(string) (model.User, error)
	// blacklistToken(string, time.Time) error
	// checkBlacklistToken(string) bool
}

func (repo *userRepository) RegisterUser(newUser model.User) (model.User, error) {

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

func (repo *userRepository) LoginUser(login dto.LoginRequest) (model.User, error) {
	user := model.User{}
	err := repo.db.Get(&user, "SELECT * FROM users WHERE email=$1", login.Email)

	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

func (repo *userRepository) GetUserByID(id string) (model.User, error) {
	user := model.User{}
	err := repo.db.Get(&user, "SELECT * FROM users WHERE id=$1", id)

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
