package auth_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/hadygust/url-shortener/internal/auth"
	"github.com/hadygust/url-shortener/internal/dto"
)

type MockService struct {
	CheckBlacklistTokenFunc func(string) bool
	GetUserByIDFunc         func(string) (dto.UserResponse, error)
	jwtSecret               string
}

// JwtSecret implements [auth.Service].
func (m *MockService) JwtSecret() string {
	return m.jwtSecret
}

func (m *MockService) RegisterUser(req dto.RegisterRequest) (dto.UserResponse, error) {
	return dto.UserResponse{}, nil
}

func (m *MockService) LoginUser(req dto.LoginRequest) (dto.UserResponse, string, error) {
	return dto.UserResponse{}, "", nil
}

func (m *MockService) BlacklistToken(jti string, exp time.Time) error {
	return nil
}

func (m *MockService) CheckBlacklistToken(jti string) bool {
	if m.CheckBlacklistTokenFunc != nil {
		return m.CheckBlacklistTokenFunc(jti)
	}
	return true
}

func (m *MockService) GetUserByID(id string) (dto.UserResponse, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(id)
	}
	return dto.UserResponse{
		ID:    uuid.New(),
		Email: "test@example.com",
	}, nil
}

func TestAuthMiddleware_NoCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	middleware := auth.NewMiddleware(&MockService{})

	router.GET("/", middleware.RequireAuth, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}

}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	middleware := auth.NewMiddleware(&MockService{})

	router.GET("/protected", middleware.RequireAuth, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest("GET", "/protected", nil)
	request.AddCookie(&http.Cookie{
		Name:  "Authentication",
		Value: "invalid-token",
	})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}

}

func TestAuthMiddleware_BlacklistToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	middleware := auth.NewMiddleware(&MockService{jwtSecret: "test-secret"})

	router.GET("/blacklist", middleware.RequireAuth, func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  "test-userId",
		"exp": time.Now().Add(time.Minute * 15).Unix(),
		"jti": "test-jti",
	})
	tokenString, _ := token.SignedString([]byte("test-secret"))

	fmt.Printf("Token String: %s\n", tokenString)

	request := httptest.NewRequest("GET", "/blacklist", nil)
	request.AddCookie(&http.Cookie{
		Name:  "Authentication",
		Value: tokenString,
	})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockService := &MockService{
		jwtSecret: "test-secret",
		CheckBlacklistTokenFunc: func(jti string) bool {
			return true // Token is valid (not blacklisted)
		},
		GetUserByIDFunc: func(id string) (dto.UserResponse, error) {
			return dto.UserResponse{
				ID:    uuid.New(),
				Email: "user@example.com",
			}, nil
		},
	}

	middleware := auth.NewMiddleware(mockService)

	router.GET("/protected", middleware.RequireAuth, func(ctx *gin.Context) {
		user, _ := ctx.Get("user")
		if user != nil {
			ctx.JSON(http.StatusOK, gin.H{"message": "success"})
		}
	})

	// Generate valid JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  "test-userId",
		"exp": time.Now().Add(time.Minute * 15).Unix(),
		"jti": "valid-jti",
	})
	tokenString, _ := token.SignedString([]byte("test-secret"))

	request := httptest.NewRequest("GET", "/protected", nil)
	request.AddCookie(&http.Cookie{
		Name:  "Authentication",
		Value: tokenString,
	})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAuthMiddleware_RequireNonAuth_WithCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	middleware := auth.NewMiddleware(&MockService{})

	router.GET("/register", middleware.RequireNonAuth, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest("GET", "/register", nil)
	request.AddCookie(&http.Cookie{
		Name:  "Authentication",
		Value: "some-token",
	})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_RequireNonAuth_NoCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	middleware := auth.NewMiddleware(&MockService{})

	router.GET("/register", middleware.RequireNonAuth, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest("GET", "/register", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	middleware := auth.NewMiddleware(&MockService{jwtSecret: "test-secret"})

	router.GET("/protected", middleware.RequireAuth, func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})

	// Generate expired JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  "test-userId",
		"exp": time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
		"jti": "expired-jti",
	})
	tokenString, _ := token.SignedString([]byte("test-secret"))

	request := httptest.NewRequest("GET", "/protected", nil)
	request.AddCookie(&http.Cookie{
		Name:  "Authentication",
		Value: tokenString,
	})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, w.Code)
	}
}
