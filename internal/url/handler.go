package url

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hadygust/url-shortener/internal/dto"
)

func (h *handler) CreateUrl(c *gin.Context) {
	var urlReq dto.CreateUrlRequest
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
	log.Printf("User: %v", user.(dto.UserResponse))

	urlResp, err := h.svc.CreateUrl(urlReq, user.(dto.UserResponse).ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, urlResp)
}

func (h *handler) GetAllUserUrl(c *gin.Context) {
	user, ok := c.Get("user")

	if !ok {
		c.JSON(http.StatusUnauthorized, "user not found")
		return
	}

	urls, err := h.svc.GetAllUserUrl(user.(dto.UserResponse).ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, urls)
}

func (h *handler) GetOrigin(c *gin.Context) {
	shortCode := c.Param("shortCode")
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	log.Printf("ip: %s, agent: %s", ipAddress, userAgent)

	origin, err := h.svc.GetOrigin(shortCode, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.Redirect(http.StatusMovedPermanently, origin)
}

func (h *handler) DeleteUrl(c *gin.Context) {
	shortCode := c.Param("shortCode")
	userId, ok := c.Get("user")

	if !ok {
		c.JSON(http.StatusUnauthorized, errors.New("you must be logged in to do that"))
		return
	}

	url, err := h.svc.DeleteUrl(shortCode, userId.(dto.UserResponse).ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, url)
}

type handler struct {
	svc Service
}

func NewHandler(svc Service) *handler {
	return &handler{
		svc: svc,
	}
}
