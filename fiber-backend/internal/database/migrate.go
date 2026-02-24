package database

import (
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func RunMigrations(dbURL string, command string, version int) error {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return err
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return err
	}

	switch command {
	case "up":
		err = m.Up()
	case "down":
		err = m.Down()
	case "drop":
		err = m.Drop()
	case "force":
		err = m.Force(version)
	}

	if err == migrate.ErrNoChange {
		return nil
	}

	return err
}
