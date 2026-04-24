package cmd


func app *Application mount()



type Service interface {
}

type Config struct {
	db dbConfig
}

type dbConfig struct {
	dsn string
}

type Application struct {
	svc Service
	cfg Config
}
