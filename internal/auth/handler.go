package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hadygust/url-shortener/internal/dto"
)

func (h *handler) RegisterUser(c *gin.Context) {
	var register dto.RegisterRequest
	err := c.BindJSON(&register)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.svc.RegisterUser(register)
	if err != nil {
		if errors.Is(err, ErrEmailUsed) {
			c.AbortWithStatusJSON(http.StatusConflict, err.Error())
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *handler) LoginUser(c *gin.Context) {
	var login dto.LoginRequest
	err := c.BindJSON(&login)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	user, tokenString, err := h.svc.LoginUser(login)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.SetCookie("Authentication", tokenString, 60*15, "", "", false, true)

	c.JSON(http.StatusOK, user)
}

func (h *handler) Logout(c *gin.Context) {
	tokenString, err := c.Cookie("Authentication")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, err.Error())
		return
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if err != nil {
			return nil, err
		}
		return []byte(h.svc.JwtSecret()), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, err.Error())
		return
	}

	claims := token.Claims.(jwt.MapClaims)

	jti := claims["jti"].(string)
	exp := int64(claims["exp"].(float64))

	h.svc.BlacklistToken(jti, time.Unix(exp, 0))

	c.SetCookie("Authentication", "", -1, "", "", false, true)
	c.JSON(http.StatusOK, "Success")
}

type handler struct {
	svc Service
}

func NewUserHandler(svc Service) *handler {
	return &handler{
		svc: svc,
	}
}
