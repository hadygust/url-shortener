package auth

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func (m *AuthMiddleware) RequireAuth(c *gin.Context) {
	tokenString, err := c.Cookie("Authentication")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, err.Error())
		return
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		secret := m.svc.JwtSecret()
		if err != nil {
			return nil, err
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil {
		log.Println("Parse failed")
		c.SetCookie("Authentication", "", -1, "", "", false, true)
		c.AbortWithStatusJSON(http.StatusUnauthorized, err.Error())
		return
	}

	claims := token.Claims.(jwt.MapClaims)

	id := claims["id"].(string)
	jti := claims["jti"].(string)

	ok := m.svc.CheckBlacklistToken(jti)
	if !ok {
		c.SetCookie("Authentication", "", -1, "", "", false, true)
		c.AbortWithStatusJSON(http.StatusForbidden, "token invalid")
		return
	}

	user, err := m.svc.GetUserByID(id)

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
