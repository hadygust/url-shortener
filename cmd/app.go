package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hadygust/url-shortener/internal/auth"
	"github.com/jmoiron/sqlx"
)

func (app *application) mount() *gin.Engine {
	r := gin.Default()

	r.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, "Helo Url")
	})

	authRepo := auth.NewRepository(app.db)
	authService := auth.NewUserService(authRepo)
	authHandler := auth.NewUserHandler(authService)

	auth := r.Group("/auth")
	auth.POST("/register", authHandler.RegisterUser)

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
	db  *sqlx.DB
}
