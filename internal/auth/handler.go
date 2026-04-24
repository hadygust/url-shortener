package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handler) RegisterUser(c *gin.Context) {
	var register RegisterRequest
	err := c.BindJSON(&register)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.svc.registerUser(register)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, user)
}

type handler struct {
	svc Service
}

func NewUserHandler(svc Service) *handler {
	return &handler{
		svc: svc,
	}
}
