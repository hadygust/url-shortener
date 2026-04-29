package auth

import (
	"errors"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/hadygust/url-shortener/internal/cache"
	"github.com/hadygust/url-shortener/internal/dto"
	"github.com/hadygust/url-shortener/internal/model"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	RegisterUser(dto.RegisterRequest) (dto.UserResponse, error)
	LoginUser(dto.LoginRequest) (dto.UserResponse, string, error)
	BlacklistToken(string, time.Time) error
	GetUserByID(string) (dto.UserResponse, error)
	CheckBlacklistToken(string) bool
	JwtSecret() string
}

var (
	ErrEmailUsed = errors.New("Email already registered")
)

func (s *userService) RegisterUser(register dto.RegisterRequest) (dto.UserResponse, error) {
	// Generate Password
	password, err := bcrypt.GenerateFromPassword([]byte(register.Password), bcrypt.DefaultCost)
	if err != nil {
		return dto.UserResponse{}, err
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
		return dto.UserResponse{}, err
	}

	userResp := dto.NewUserResponse(registeredUser)

	return *userResp, nil
}

func (s *userService) LoginUser(login dto.LoginRequest) (dto.UserResponse, string, error) {
	user, err := s.repo.loginUser(login)
	if err != nil {
		return dto.UserResponse{}, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password))
	if err != nil {
		return dto.UserResponse{}, "", err
	}

	key := s.jwtSecret

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  user.ID,
		"exp": time.Now().Add(time.Minute * 15).Unix(),
		"jti": uuid.New(),
	})

	tokenString, err := token.SignedString([]byte(key))
	if err != nil {
		return dto.UserResponse{}, "", err
	}

	return *dto.NewUserResponse(user), tokenString, nil
}

func (s *userService) BlacklistToken(jti string, exp time.Time) error {
	return s.cache.Set("blacklist:"+jti, 1, time.Until(exp))
}

func (s *userService) CheckBlacklistToken(jti string) bool {
	test, err := s.cache.Get("blacklist:" + jti)

	log.Printf("Black list: %#v err: %#v", test, err)

	if test == nil {
		// Not found -> safe
		return true
	}

	// Found -> not safe
	return false
}

func (s *userService) GetUserByID(id string) (dto.UserResponse, error) {
	user, err := s.repo.getUserByID(id)
	if err != nil {
		return dto.UserResponse{}, err
	}

	return *dto.NewUserResponse(user), nil
}

func (s *userService) JwtSecret() string {
	return s.jwtSecret
}

type userService struct {
	repo      Repository
	cache     cache.Cache
	jwtSecret string
}

func NewUserService(repo Repository, cache cache.Cache, jwtSecret string) Service {
	return &userService{
		repo:      repo,
		cache:     cache,
		jwtSecret: jwtSecret,
	}
}
