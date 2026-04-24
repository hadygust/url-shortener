package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *application) mount() *gin.Engine {
	r := gin.Default()

	r.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, "Helo Url")
	})
	return r
}

func (app *application) run(router *gin.Engine) error {

	return router.Run(app.cfg.addr)
}

type Config struct {
	addr string
	db   dbConfig
}

type dbConfig struct {
	dsn string
}

type application struct {
	cfg Config
}
