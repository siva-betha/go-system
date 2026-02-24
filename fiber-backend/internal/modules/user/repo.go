package user

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the contract for user data access.
// Concrete implementation: PgRepo. Tests can supply a mock.
type Repository interface {
	Create(ctx context.Context, u *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id int64) (*User, error)
	List(ctx context.Context) ([]User, error)
	Update(ctx context.Context, id int64, u *User) error
	Delete(ctx context.Context, id int64) error
}

// PgRepo is the PostgreSQL-backed implementation of Repository.
type PgRepo struct {
	DB *pgxpool.Pool
}

// Compile-time check: PgRepo must satisfy Repository.
var _ Repository = (*PgRepo)(nil)

// Create inserts a new user and populates the generated fields.
func (r PgRepo) Create(ctx context.Context, u *User) error {
	return r.DB.QueryRow(ctx,
		`INSERT INTO users(name, email, password_hash, role)
		 VALUES($1, $2, $3, $4)
		 RETURNING id, created_at, updated_at`,
		u.Name, u.Email, u.PasswordHash, u.Role,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

// GetByEmail returns a single user by email address.
func (r PgRepo) GetByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := r.DB.QueryRow(ctx,
		`SELECT id, name, email, password_hash, role, created_at, updated_at
		 FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetByID returns a single user by primary key.
func (r PgRepo) GetByID(ctx context.Context, id int64) (*User, error) {
	var u User
	err := r.DB.QueryRow(ctx,
		`SELECT id, name, email, password_hash, role, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// List returns all users (without password hashes).
func (r PgRepo) List(ctx context.Context) ([]User, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT id, name, email, role, created_at, updated_at FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, nil
}

// Update modifies name and email for the given user.
func (r PgRepo) Update(ctx context.Context, id int64, u *User) error {
	ct, err := r.DB.Exec(ctx,
		`UPDATE users SET name=$1, email=$2, updated_at=now() WHERE id=$3`,
		u.Name, u.Email, id,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// Delete removes a user by ID.
func (r PgRepo) Delete(ctx context.Context, id int64) error {
	ct, err := r.DB.Exec(ctx,
		`DELETE FROM users WHERE id=$1`, id,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
