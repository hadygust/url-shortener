package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hadygust/url-shortener/internal/auth"
	"github.com/hadygust/url-shortener/internal/cache"
	"github.com/hadygust/url-shortener/internal/env"
	redirectlog "github.com/hadygust/url-shortener/internal/redirect_log"
	"github.com/hadygust/url-shortener/internal/url"
	"github.com/jmoiron/sqlx"
)

func (app *application) mount() *gin.Engine {
	r := gin.Default()

	jwtSecret, err := env.LoadEnv("JWT_SECRET")
	if err != nil {
		panic("JWT KEY NOT FOUND")
	}

	authRepo := auth.NewRepository(app.db)
	authService := auth.NewUserService(authRepo, app.cache, jwtSecret)
	authHandler := auth.NewUserHandler(authService)

	authMiddleware := auth.NewMiddleware(authService)

	auth := r.Group("/auth")
	auth.POST("/register", authMiddleware.RequireNonAuth, authHandler.RegisterUser)
	auth.POST("/login", authMiddleware.RequireNonAuth, authHandler.LoginUser)
	auth.POST("/logout", authMiddleware.RequireAuth, authHandler.Logout)

	redirectLogRepo := redirectlog.NewRepository(app.db)
	redirectLogService := redirectlog.NewService(redirectLogRepo)

	urlRepo := url.NewRepository(app.db)
	urlService := url.NewService(urlRepo, redirectLogService, app.cache)
	urlHandler := url.NewHandler(urlService)

	url := r.Group("/urls")
	url.POST("/", authMiddleware.RequireAuth, urlHandler.CreateUrl)
	url.GET("/", authMiddleware.RequireAuth, urlHandler.GetAllUserUrl)
	url.DELETE("/:shortCode", authMiddleware.RequireAuth, urlHandler.DeleteUrl)

	r.GET("/:shortCode", urlHandler.GetOrigin)

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
	cache cache.Cache
}
