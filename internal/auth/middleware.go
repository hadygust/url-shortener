package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hadygust/url-shortener/internal/env"
)

func (m *AuthMiddleware) RequireAuth(c *gin.Context) {
	tokenString, err := c.Cookie("Authentication")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, err.Error())
		return
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		secret, err := env.LoadEnv("JWT_SECRET")
		if err != nil {
			return nil, err
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, err.Error())
		return
	}

	claims := token.Claims.(jwt.MapClaims)

	id := claims["id"].(string)

	user, err := m.svc.getUserByID(id)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, err.Error())
		return
	}

	c.Set("user", user)

	c.Next()
}

func (m *AuthMiddleware) RequireNonAuth(c *gin.Context) {
	_, err := c.Cookie("Authentication")

	if err == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, "You must be logged out to do that")
		return
	}

	c.Next()
}

type AuthMiddleware struct {
	svc Service
}

func NewMiddleware(svc Service) *AuthMiddleware {
	return &AuthMiddleware{
		svc: svc,
	}
}
