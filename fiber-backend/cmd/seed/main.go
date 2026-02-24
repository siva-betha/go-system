package main

import (
	"context"
	"log"
	"time"

	"fiber-backend/internal/auth"
	"fiber-backend/internal/config"
	"fiber-backend/internal/database"
	"fiber-backend/internal/modules/user"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DBUrl)
	if err != nil {
		log.Fatal("Could not connect to DB:", err)
	}
	defer db.Close()

	userRepo := user.PgRepo{DB: db}

	type seedUser struct {
		Name     string
		Email    string
		Password string
		Role     string
	}

	seedUsers := []seedUser{
		{Name: "Admin User", Email: "admin@example.com", Password: "Admin@123456", Role: "admin"},
		{Name: "John Doe", Email: "john.doe@example.com", Password: "John@123456", Role: "user"},
		{Name: "Jane Smith", Email: "jane.smith@example.com", Password: "Jane@123456", Role: "user"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("Seeding users...")
	for _, su := range seedUsers {
		hash, err := auth.HashPassword(su.Password)
		if err != nil {
			log.Printf("Failed to hash password for %s: %v\n", su.Name, err)
			continue
		}

		u := user.User{
			Name:         su.Name,
			Email:        su.Email,
			PasswordHash: hash,
			Role:         su.Role,
		}

		err = userRepo.Create(ctx, &u)
		if err != nil {
			log.Printf("Failed to seed user %s: %v\n", su.Name, err)
			continue
		}
		log.Printf("Seeded user: %s (ID: %d, Role: %s)\n", su.Name, u.ID, su.Role)
	}

	log.Println("Seeding completed successfully!")
}
