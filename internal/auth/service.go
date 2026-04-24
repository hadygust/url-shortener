package auth

import (
	"errors"

	"github.com/google/uuid"
	"github.com/hadygust/url-shortener/internal/model"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	registerUser(RegisterRequest) (UserResponse, error)
}

var (
	ErrEmailUsed = errors.New("Email already registered")
)

func (s *userService) registerUser(register RegisterRequest) (UserResponse, error) {
	// Generate Password
	password, err := bcrypt.GenerateFromPassword([]byte(register.Password), bcrypt.DefaultCost)
	if err != nil {
		return UserResponse{}, err
	}

	// Create user model
	newUser := model.User{
		ID:       uuid.New(),
		Email:    register.Email,
		Password: string(password),
	}

	// Register new user
	registeredUser, err := s.repo.registerUser(newUser)
	if err != nil {
		return UserResponse{}, err
	}

	userResp := NewUserResponse(registeredUser)

	return *userResp, nil
}

type userService struct {
	repo Repository
}

func NewUserService(repo Repository) Service {
	return &userService{
		repo: repo,
	}
}
