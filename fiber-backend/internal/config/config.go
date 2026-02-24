package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string
	DBUrl   string
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	db := "postgres://" +
		os.Getenv("DB_USER") + ":" +
		os.Getenv("DB_PASS") + "@" +
		os.Getenv("DB_HOST") + ":" +
		os.Getenv("DB_PORT") + "/" +
		os.Getenv("DB_NAME") +
		"?sslmode=" + os.Getenv("DB_SSL")

	return Config{
		AppPort: os.Getenv("APP_PORT"),
		DBUrl:   db,
	}
}
