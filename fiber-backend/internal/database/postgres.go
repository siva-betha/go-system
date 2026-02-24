package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}

	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour

	return pgxpool.NewWithConfig(context.Background(), cfg)
}
