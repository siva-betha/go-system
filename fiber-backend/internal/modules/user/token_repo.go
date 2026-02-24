package user

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TokenRepository handles refresh token persistence.
type TokenRepository interface {
	StoreRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error
	FindRefreshToken(ctx context.Context, tokenHash string) (userID int64, expiresAt time.Time, err error)
	DeleteRefreshToken(ctx context.Context, tokenHash string) error
	DeleteAllUserTokens(ctx context.Context, userID int64) error
}

// PgTokenRepo is the PostgreSQL-backed implementation of TokenRepository.
type PgTokenRepo struct {
	DB *pgxpool.Pool
}

var _ TokenRepository = (*PgTokenRepo)(nil)

func (r PgTokenRepo) StoreRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	_, err := r.DB.Exec(ctx,
		`INSERT INTO refresh_tokens(user_id, token_hash, expires_at)
		 VALUES($1, $2, $3)`,
		userID, tokenHash, expiresAt,
	)
	return err
}

func (r PgTokenRepo) FindRefreshToken(ctx context.Context, tokenHash string) (int64, time.Time, error) {
	var userID int64
	var expiresAt time.Time
	err := r.DB.QueryRow(ctx,
		`SELECT user_id, expires_at FROM refresh_tokens WHERE token_hash = $1`,
		tokenHash,
	).Scan(&userID, &expiresAt)
	return userID, expiresAt, err
}

func (r PgTokenRepo) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.DB.Exec(ctx,
		`DELETE FROM refresh_tokens WHERE token_hash = $1`,
		tokenHash,
	)
	return err
}

func (r PgTokenRepo) DeleteAllUserTokens(ctx context.Context, userID int64) error {
	_, err := r.DB.Exec(ctx,
		`DELETE FROM refresh_tokens WHERE user_id = $1`,
		userID,
	)
	return err
}
