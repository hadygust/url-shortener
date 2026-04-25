package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hadygust/url-shortener/internal/auth"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

func (app *application) mount() *gin.Engine {
	r := gin.Default()

	authRepo := auth.NewRepository(app.db, app.redis)
	authService := auth.NewUserService(authRepo)
	authHandler := auth.NewUserHandler(authService)

	authMiddleware := auth.NewMiddleware(authService)

	auth := r.Group("/auth")
	auth.POST("/register", authMiddleware.RequireNonAuth, authHandler.RegisterUser)
	auth.POST("/login", authMiddleware.RequireNonAuth, authHandler.LoginUser)
	auth.POST("/logout", authMiddleware.RequireAuth, authHandler.Logout)

	r.GET("/", authMiddleware.RequireAuth, func(ctx *gin.Context) {
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
	cfg   Config
	db    *sqlx.DB
	redis *redis.Client
}
