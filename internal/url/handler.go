package url

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hadygust/url-shortener/internal/auth"
)

func (h *handler) CreateUrl(c *gin.Context) {
	var urlReq CreateUrlRequest
	if err := c.BindJSON(&urlReq); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	log.Printf("Url Req: %v", urlReq)

	user, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, errors.New("You must be logged in to do that"))
		return
	}
	log.Printf("User: %v", user.(auth.UserResponse))

	urlResp, err := h.svc.CreateUrl(urlReq, user.(auth.UserResponse).ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, urlResp)
}

type handler struct {
	svc Service
}

func NewHandler(svc Service) *handler {
	return &handler{
		svc: svc,
	}
}
