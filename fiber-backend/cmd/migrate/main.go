package main

import (
	"log"
	"os"

	"strconv"

	"fiber-backend/internal/config"
	"fiber-backend/internal/database"

	"github.com/golang-migrate/migrate/v4"
)

func main() {
	cfg := config.Load()

	if len(os.Args) < 2 {
		log.Fatal("expected 'up' or 'down' subcommands")
	}

	switch os.Args[1] {
	case "up":
		if err := database.RunMigrations(cfg.DBUrl, "up", 0); err != nil {
			if err == migrate.ErrNoChange {
				log.Println("No migrations to apply")
				return
			}
			log.Fatal("migration up failed: ", err)
		}
		log.Println("Migrations applied successfully")
	case "down":
		if err := database.RunMigrations(cfg.DBUrl, "down", 0); err != nil {
			if err == migrate.ErrNoChange {
				log.Println("No migrations to revert")
				return
			}
			log.Fatal("migration down failed: ", err)
		}
		log.Println("Migrations reverted successfully")
	case "drop":
		if err := database.RunMigrations(cfg.DBUrl, "drop", 0); err != nil {
			log.Fatal("migration drop failed: ", err)
		}
		log.Println("Database dropped successfully")
	case "force":
		if len(os.Args) < 3 {
			log.Fatal("expected version number for 'force' command")
		}
		v, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatal("invalid version number: ", err)
		}
		if err := database.RunMigrations(cfg.DBUrl, "force", v); err != nil {
			log.Fatal("migration force failed: ", err)
		}
		log.Println("Migration version forced successfully")
	default:
		log.Fatal("expected 'up', 'down', 'drop', or 'force' subcommands")
	}
}
