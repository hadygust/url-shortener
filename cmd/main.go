package main

import (
	"fmt"
	"log"

	"github.com/hadygust/url-shortener/internal/env"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	host := env.LoadEnv("POSTGRES_HOST", "localhost")
	port := env.LoadEnv("POSTGRES_PORT", "5432")
	user := env.LoadEnv("POSTGRES_USER", "postgres")
	dbname := env.LoadEnv("POSTGRES_DB", "url_short")
	password := env.LoadEnv("POSTGRES_PW", "postgres")
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

	app := application{
		cfg: cfg,
		db:  db,
	}

	err = app.run(app.mount())
	if err != nil {
		log.Fatal(err.Error())
	}
}
