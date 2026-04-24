package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Print("Hello")

	cfg := Config{
		addr: ":8080",
		db: dbConfig{
			dsn: "",
		},
	}

	app := application{
		cfg: cfg,
	}

	err := app.run(app.mount())
	if err != nil {
		log.Fatal(err.Error())
	}
}
