package database

import (
	"context"
	"log"
	"time"

	"fiber-backend/internal/auth"
	"fiber-backend/internal/modules/machine_config"
	"fiber-backend/internal/modules/user"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Seed(db *pgxpool.Pool) {
	seedUsers(db)
	seedMachines(db)
	log.Println("Global seeding phase completed!")
}

func seedUsers(db *pgxpool.Pool) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	userRepo := user.PgRepo{DB: db}
	users, err := userRepo.List(ctx)
	if err != nil {
		log.Printf("User seeding check failed: %v\n", err)
		return
	}

	if len(users) > 0 {
		log.Println("Database already has users, skipping user seeding.")
		return
	}

	log.Println("Seeding initial user data...")
	seedUsersList := []struct {
		Name     string
		Email    string
		Password string
		Role     string
	}{
		{Name: "Admin User", Email: "admin@example.com", Password: "Admin@123456", Role: "admin"},
		{Name: "John Doe", Email: "john.doe@example.com", Password: "John@123456", Role: "user"},
		{Name: "Jane Smith", Email: "jane.smith@example.com", Password: "Jane@123456", Role: "user"},
	}

	for _, su := range seedUsersList {
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
}

func seedMachines(db *pgxpool.Pool) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	repo := machine_config.PgRepo{DB: db}
	existing, err := repo.GetMachines(ctx)
	if err == nil && len(existing) > 0 {
		log.Println("Database already has machine configurations, skipping machine seeding.")
		return
	}

	machines := []machine_config.Machine{
		{
			ID:       uuid.New().String(),
			Name:     "PLC-Alpha-01",
			IP:       "192.168.1.10",
			AmsNetID: "5.67.89.10.1.1",
			Port:     851,
			Chambers: []machine_config.Chamber{
				{
					ID:   uuid.New().String(),
					Name: "Coating Chamber A",
					Symbols: []machine_config.Symbol{
						{ID: uuid.New().String(), Name: "GVL.temp_setpoint", DataType: "float", Unit: "°C"},
						{ID: uuid.New().String(), Name: "GVL.pressure_actual", DataType: "float", Unit: "mbar"},
						{ID: uuid.New().String(), Name: "GVL.status_ready", DataType: "bool"},
					},
				},
				{
					ID:   uuid.New().String(),
					Name: "Loading Zone",
					Symbols: []machine_config.Symbol{
						{ID: uuid.New().String(), Name: "GVL.conveyor_speed", DataType: "int", Unit: "m/s"},
					},
				},
			},
		},
		{
			ID:       uuid.New().String(),
			Name:     "PLC-Beta-02",
			IP:       "192.168.1.11",
			AmsNetID: "5.67.89.11.1.1",
			Port:     851,
			Chambers: []machine_config.Chamber{
				{
					ID:   uuid.New().String(),
					Name: "Curing Oven",
					Symbols: []machine_config.Symbol{
						{ID: uuid.New().String(), Name: "GVL.oven_temp", DataType: "float", Unit: "°C"},
						{ID: uuid.New().String(), Name: "GVL.timer_remaining", DataType: "int", Unit: "sec"},
					},
				},
			},
		},
	}

	if err := repo.SaveMachines(ctx, machines); err != nil {
		log.Printf("Failed to seed machines: %v\n", err)
	} else {
		log.Println("Seeded 2 dummy PLC machine configurations.")
	}
}
