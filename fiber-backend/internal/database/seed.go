package database

import (
	"context"
	"log"
	"time"

	"fiber-backend/internal/auth"
	"fiber-backend/internal/modules/user"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Seed populates the database with initial data. It is idempotent —
// if users already exist the function returns immediately.
func Seed(db *pgxpool.Pool) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userRepo := user.PgRepo{DB: db}

	// Check if users already exist to make seeding idempotent
	users, err := userRepo.List(ctx)
	if err != nil {
		log.Printf("Seeding check failed: %v\n", err)
		return
	}

	if len(users) > 0 {
		log.Println("Database already has users, skipping seeding.")
		return
	}

	log.Println("Seeding initial data...")

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

		if err := userRepo.Create(ctx, &u); err != nil {
			log.Printf("Failed to seed user %s: %v\n", su.Name, err)
			continue
		}
		log.Printf("Seeded user: %s (%s)\n", su.Name, su.Role)
	}

	log.Println("Seeding completed!")
}
