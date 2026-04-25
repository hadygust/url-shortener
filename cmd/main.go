package main

import (
	"fmt"
	"log"

	"github.com/hadygust/url-shortener/internal/env"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	host := env.LoadEnvFallback("POSTGRES_HOST", "localhost")
	port := env.LoadEnvFallback("POSTGRES_PORT", "5432")
	user := env.LoadEnvFallback("POSTGRES_USER", "postgres")
	dbname := env.LoadEnvFallback("POSTGRES_DB", "url_short")
	password := env.LoadEnvFallback("POSTGRES_PW", "postgres")
	dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable", host, port, dbname, user, password)

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

	redis := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // No password set
		DB:       0,  // Use default DB
	})
	app := application{
		cfg:   cfg,
		db:    db,
		redis: redis,
	}

	err = app.run(app.mount())
	if err != nil {
		log.Fatal(err.Error())
	}
}
