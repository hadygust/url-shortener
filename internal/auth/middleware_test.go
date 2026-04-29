package auth_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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
	// if m.CheckBlacklistTokenFunc != nil {
	// 	return m.CheckBlacklistTokenFunc(jti)
	// }
	return false
}

func (m *MockService) GetUserByID(id string) (dto.UserResponse, error) {
	// if m.GetUserByIDFunc != nil {
	// 	return m.GetUserByIDFunc(id)
	// }
	return dto.UserResponse{}, nil
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

	// fmt.Printf("Key: %s\n", key)
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
