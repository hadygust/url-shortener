package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/hadygust/url-shortener/internal/env"
	"github.com/hadygust/url-shortener/internal/model"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	registerUser(RegisterRequest) (UserResponse, error)
	loginUser(LoginRequest) (UserResponse, string, error)
	blacklistToken(string, time.Time) error
	getUserByID(string) (UserResponse, error)
	checkBlacklistToken(string) bool
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

func (s *userService) loginUser(login LoginRequest) (UserResponse, string, error) {
	user, err := s.repo.loginUser(login)
	if err != nil {
		return UserResponse{}, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password))
	if err != nil {
		return UserResponse{}, "", err
	}

	key, err := env.LoadEnv("JWT_SECRET")
	if err != nil {
		return UserResponse{}, "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  user.ID,
		"exp": time.Now().Add(time.Minute * 15).Unix(),
		"jti": uuid.New(),
	})

	tokenString, err := token.SignedString([]byte(key))
	if err != nil {
		return UserResponse{}, "", err
	}

	return *NewUserResponse(user), tokenString, nil
}

func (s *userService) blacklistToken(jti string, exp time.Time) error {
	return s.repo.blacklistToken(jti, exp)
}

func (s *userService) getUserByID(id string) (UserResponse, error) {
	user, err := s.repo.getUserByID(id)
	if err != nil {
		return UserResponse{}, err
	}

	return *NewUserResponse(user), nil
}

func (s *userService) checkBlacklistToken(jti string) bool {
	return s.repo.checkBlacklistToken(jti)
}

type userService struct {
	repo Repository
}

func NewUserService(repo Repository) Service {
	return &userService{
		repo: repo,
	}
}
