package auth_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/hadygust/url-shortener/internal/auth"
	"github.com/hadygust/url-shortener/internal/cache"
	"github.com/hadygust/url-shortener/internal/dto"
	"github.com/hadygust/url-shortener/internal/model"
	"golang.org/x/crypto/bcrypt"
)

// Integration tests for auth endpoints

type MockAuthRepository struct {
	registerUserFunc func(model.User) (model.User, error)
	loginUserFunc    func(dto.LoginRequest) (model.User, error)
	getUserByIDFunc  func(string) (model.User, error)
}

func (m *MockAuthRepository) RegisterUser(user model.User) (model.User, error) {
	if m.registerUserFunc != nil {
		return m.registerUserFunc(user)
	}
	return user, nil
}

func (m *MockAuthRepository) LoginUser(login dto.LoginRequest) (model.User, error) {
	if m.loginUserFunc != nil {
		return m.loginUserFunc(login)
	}
	return model.User{}, nil
}

func (m *MockAuthRepository) GetUserByID(id string) (model.User, error) {
	if m.getUserByIDFunc != nil {
		return m.getUserByIDFunc(id)
	}
	return model.User{}, nil
}

// MockCacheForAuth for auth service testing
type MockCacheForAuth struct {
	GetFunc func(string) (any, error)
	SetFunc func(string, any, time.Duration) error
}

// GetIPRateLimit implements [cache.Cache].
func (m *MockCacheForAuth) GetIPRateLimit(ip string) (*cache.RequestCounter, error) {
	panic("unimplemented")
}

// GetRedirectLimit implements [cache.Cache].
func (m *MockCacheForAuth) GetRedirectLimit(urlID string) (*cache.RequestCounter, error) {
	panic("unimplemented")
}

// SetIPRateLimit implements [cache.Cache].
func (m *MockCacheForAuth) SetIPRateLimit(ip string, counter *cache.RequestCounter, ttl time.Duration) error {
	panic("unimplemented")
}

// SetRedirectLimit implements [cache.Cache].
func (m *MockCacheForAuth) SetRedirectLimit(urlID string, counter *cache.RequestCounter, ttl time.Duration) error {
	panic("unimplemented")
}

func (m *MockCacheForAuth) Get(key string) (any, error) {
	if m.GetFunc != nil {
		return m.GetFunc(key)
	}
	return nil, nil
}

func (m *MockCacheForAuth) Set(key string, value any, ttl time.Duration) error {
	if m.SetFunc != nil {
		return m.SetFunc(key, value, ttl)
	}
	return nil
}

func (m *MockCacheForAuth) Delete(key string) error {
	return nil
}

func (m *MockCacheForAuth) GetUrl(shortCode string) (*dto.UrlCache, error) {
	return nil, nil
}

func (m *MockCacheForAuth) SetUrl(model model.Url) error {
	return nil
}

func (m *MockCacheForAuth) GetUrlPostLimit(userID string) (*cache.RequestCounter, error) {
	return nil, nil
}

func (m *MockCacheForAuth) SetUrlPostLimit(userID string, counter *cache.RequestCounter, ttl time.Duration) error {
	return nil
}

func (m *MockCacheForAuth) GetUrlGetLimit(key string) (*cache.RequestCounter, error) {
	return nil, nil
}

func (m *MockCacheForAuth) SetUrlGetLimit(key string, counter *cache.RequestCounter, ttl time.Duration) error {
	return nil
}

func TestRegisterUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userId := uuid.New()
	mockRepo := &MockAuthRepository{
		registerUserFunc: func(user model.User) (model.User, error) {
			user.ID = userId
			return user, nil
		},
	}

	mockCache := &MockCacheForAuth{}

	service := auth.NewUserService(mockRepo, mockCache, "test-secret")
	handler := auth.NewUserHandler(service)

	router := gin.New()
	router.POST("/auth/register", handler.RegisterUser)

	reqBody := dto.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, request)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response dto.UserResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Email != reqBody.Email {
		t.Errorf("expected email %s, got %s", reqBody.Email, response.Email)
	}
}

func TestRegisterUser_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockAuthRepository{}
	mockCache := &MockCacheForAuth{}

	service := auth.NewUserService(mockRepo, mockCache, "test-secret")
	handler := auth.NewUserHandler(service)

	router := gin.New()
	router.POST("/auth/register", handler.RegisterUser)

	request := httptest.NewRequest("POST", "/auth/register", bytes.NewReader([]byte("invalid json")))
	request.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, request)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestRegisterUser_RepositoryError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockAuthRepository{
		registerUserFunc: func(user model.User) (model.User, error) {
			return model.User{}, errors.New("database error")
		},
	}

	mockCache := &MockCacheForAuth{}

	service := auth.NewUserService(mockRepo, mockCache, "test-secret")
	handler := auth.NewUserHandler(service)

	router := gin.New()
	router.POST("/auth/register", handler.RegisterUser)

	reqBody := dto.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, request)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestLoginUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userId := uuid.New()
	email := "test@example.com"
	password := "password123"

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	mockRepo := &MockAuthRepository{
		loginUserFunc: func(login dto.LoginRequest) (model.User, error) {
			return model.User{
				ID:       userId,
				Email:    email,
				Password: string(hashedPassword),
			}, nil
		},
	}

	mockCache := &MockCacheForAuth{}

	service := auth.NewUserService(mockRepo, mockCache, "test-secret")
	handler := auth.NewUserHandler(service)

	router := gin.New()
	router.POST("/auth/login", handler.LoginUser)

	reqBody := dto.LoginRequest{
		Email:    email,
		Password: password,
	}

	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, request)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	// Check if cookie is set
	var cookies []*http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == "Authentication" {
			cookies = append(cookies, c)
		}
	}

	if len(cookies) == 0 {
		t.Errorf("expected Authentication cookie to be set")
	}
}

func TestLoginUser_InvalidPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userId := uuid.New()
	email := "test@example.com"
	password := "password123"

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	mockRepo := &MockAuthRepository{
		loginUserFunc: func(login dto.LoginRequest) (model.User, error) {
			return model.User{
				ID:       userId,
				Email:    email,
				Password: string(hashedPassword),
			}, nil
		},
	}

	mockCache := &MockCacheForAuth{}

	service := auth.NewUserService(mockRepo, mockCache, "test-secret")
	handler := auth.NewUserHandler(service)

	router := gin.New()
	router.POST("/auth/login", handler.LoginUser)

	reqBody := dto.LoginRequest{
		Email:    email,
		Password: "wrongpassword",
	}

	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, request)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestLogout_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockAuthRepository{}
	mockCache := &MockCacheForAuth{
		SetFunc: func(key string, value any, ttl time.Duration) error {
			return nil
		},
	}

	service := auth.NewUserService(mockRepo, mockCache, "test-secret")
	handler := auth.NewUserHandler(service)

	router := gin.New()
	router.POST("/auth/logout", handler.Logout)

	// Create a valid JWT token
	tokenString, _ := createValidToken("test-user-id", "test-secret")

	request := httptest.NewRequest("POST", "/auth/logout", nil)
	request.AddCookie(&http.Cookie{
		Name:  "Authentication",
		Value: tokenString,
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, request)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	// Check if cookie is cleared
	var clearedCookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == "Authentication" {
			clearedCookie = c
			break
		}
	}

	if clearedCookie == nil || clearedCookie.Value != "" {
		t.Errorf("expected Authentication cookie to be cleared")
	}
}

func TestLogout_NoCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockAuthRepository{}
	mockCache := &MockCacheForAuth{}

	service := auth.NewUserService(mockRepo, mockCache, "test-secret")
	handler := auth.NewUserHandler(service)

	router := gin.New()
	router.POST("/auth/logout", handler.Logout)

	request := httptest.NewRequest("POST", "/auth/logout", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, request)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// Helper function to create valid JWT token
func createValidToken(userId string, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  userId,
		"exp": time.Now().Add(time.Minute * 15).Unix(),
		"jti": uuid.New().String(),
	})
	return token.SignedString([]byte(secret))
}
