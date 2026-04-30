package main

import (
	"fmt"
	"log"

	"github.com/hadygust/url-shortener/internal/cache"
	"github.com/hadygust/url-shortener/internal/env"
	"github.com/hadygust/url-shortener/internal/ratelimit"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	pgHost := env.LoadEnvFallback("POSTGRES_HOST", "localhost")
	pgPort := env.LoadEnvFallback("POSTGRES_PORT", "5432")
	pgUser := env.LoadEnvFallback("POSTGRES_USER", "postgres")
	dbname := env.LoadEnvFallback("POSTGRES_DB", "url_short")
	pgPassword := env.LoadEnvFallback("POSTGRES_PW", "postgres")
	dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable", pgHost, pgPort, dbname, pgUser, pgPassword)

	cfg := Config{
		addr: ":8080",
		db: dbConfig{
			dsn: dsn,
		},
	}

	db, err := sqlx.Connect("postgres", cfg.db.dsn)
	if err != nil {
		log.Fatalln(err)
	}

	redisHost := env.LoadEnvFallback("REDIS_HOST", "localhost")
	redisPort := env.LoadEnvFallback("REDIS_HOST", "6379")
	cache := cache.NewRedisCache(redisHost + ":" + redisPort)

	// Create rate limiter with cache dependency
	rateLimiter := ratelimit.NewRateLimiter(cache)

	app := application{
		cfg:         cfg,
		db:          db,
		cache:       cache,
		rateLimiter: rateLimiter,
	}

	err = app.run(app.mount())
	if err != nil {
		log.Fatal(err.Error())
	}
}
