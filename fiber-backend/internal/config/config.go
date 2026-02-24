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

	db := os.Getenv("DB_URL")
	if db == "" {
		db = "postgres://" +
			os.Getenv("POSTGRES_USER") + ":" +
			os.Getenv("POSTGRES_PASSWORD") + "@" +
			os.Getenv("POSTGRES_HOST") + ":" +
			os.Getenv("POSTGRES_PORT") + "/" +
			os.Getenv("POSTGRES_DB") +
			"?sslmode=disable"
	}

	return Config{
		AppPort: os.Getenv("APP_PORT"),
		DBUrl:   db,
	}
}
