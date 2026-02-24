package apikey

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, k *APIKey) error
	GetByHash(ctx context.Context, hash string) (*APIKey, error)
	ListByUser(ctx context.Context, userID string) ([]APIKey, error)
	Delete(ctx context.Context, id string) error
	UpdateLastUsed(ctx context.Context, id string) error
}

type PgRepo struct {
	DB *pgxpool.Pool
}

func (r PgRepo) Create(ctx context.Context, k *APIKey) error {
	_, err := r.DB.Exec(ctx,
		`INSERT INTO api_keys (user_id, name, key_hash, prefix, scopes, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		k.UserID, k.Name, k.KeyHash, k.Prefix, k.Scopes, k.ExpiresAt,
	)
	return err
}

func (r PgRepo) GetByHash(ctx context.Context, hash string) (*APIKey, error) {
	var k APIKey
	err := r.DB.QueryRow(ctx,
		`SELECT id, user_id, name, prefix, scopes, expires_at, last_used_at, created_at
		 FROM api_keys WHERE key_hash = $1`, hash,
	).Scan(&k.ID, &k.UserID, &k.Name, &k.Prefix, &k.Scopes, &k.ExpiresAt, &k.LastUsedAt, &k.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func (r PgRepo) ListByUser(ctx context.Context, userID string) ([]APIKey, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT id, user_id, name, prefix, scopes, expires_at, last_used_at, created_at
		 FROM api_keys WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []APIKey
	for rows.Next() {
		var k APIKey
		err := rows.Scan(&k.ID, &k.UserID, &k.Name, &k.Prefix, &k.Scopes, &k.ExpiresAt, &k.LastUsedAt, &k.CreatedAt)
		if err != nil {
			return nil, err
		}
		results = append(results, k)
	}
	return results, nil
}

func (r PgRepo) Delete(ctx context.Context, id string) error {
	_, err := r.DB.Exec(ctx, `DELETE FROM api_keys WHERE id = $1`, id)
	return err
}

func (r PgRepo) UpdateLastUsed(ctx context.Context, id string) error {
	_, err := r.DB.Exec(ctx, `UPDATE api_keys SET last_used_at = now() WHERE id = $1`, id)
	return err
}
