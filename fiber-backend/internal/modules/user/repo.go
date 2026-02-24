package user

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the contract for user data access.
type Repository interface {
	Create(ctx context.Context, u *User) error
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	List(ctx context.Context) ([]User, error)
	Update(ctx context.Context, id string, u *User) error
	Delete(ctx context.Context, id string) error
	GetRolesAndPermissions(ctx context.Context, userID string) ([]string, map[string][]string, error)
}

// PgRepo is the PostgreSQL-backed implementation of Repository.
type PgRepo struct {
	DB *pgxpool.Pool
}

var _ Repository = (*PgRepo)(nil)

func (r PgRepo) Create(ctx context.Context, u *User) error {
	return r.DB.QueryRow(ctx,
		`INSERT INTO users(username, email, password_hash, full_name, department, title, employee_id)
		 VALUES($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, created_at, updated_at`,
		u.Username, u.Email, u.PasswordHash, u.FullName, u.Department, u.Title, u.EmployeeID,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r PgRepo) GetByUsername(ctx context.Context, username string) (*User, error) {
	var u User
	err := r.DB.QueryRow(ctx,
		`SELECT id, username, email, password_hash, full_name, department, is_active, created_at, updated_at
		 FROM users WHERE username = $1`, username,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FullName, &u.Department, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}

	roles, perms, _ := r.GetRolesAndPermissions(ctx, u.ID)
	u.Roles = roles
	u.Permissions = perms

	return &u, nil
}

func (r PgRepo) GetByID(ctx context.Context, id string) (*User, error) {
	var u User
	err := r.DB.QueryRow(ctx,
		`SELECT id, username, email, password_hash, full_name, department, is_active, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FullName, &u.Department, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}

	roles, perms, _ := r.GetRolesAndPermissions(ctx, u.ID)
	u.Roles = roles
	u.Permissions = perms

	return &u, nil
}

func (r PgRepo) GetRolesAndPermissions(ctx context.Context, userID string) ([]string, map[string][]string, error) {
	// Fetch Roles
	roleRows, err := r.DB.Query(ctx,
		`SELECT r.name FROM roles r JOIN user_roles ur ON r.id = ur.role_id WHERE ur.user_id = $1`, userID)
	if err != nil {
		return nil, nil, err
	}
	defer roleRows.Close()
	var roles []string
	for roleRows.Next() {
		var r string
		roleRows.Scan(&r)
		roles = append(roles, r)
	}

	// Fetch Permissions
	permRows, err := r.DB.Query(ctx,
		`SELECT p.resource, p.action FROM permissions p 
		 JOIN role_permissions rp ON p.id = rp.permission_id 
		 JOIN user_roles ur ON rp.role_id = ur.role_id 
		 WHERE ur.user_id = $1`, userID)
	if err != nil {
		return roles, nil, err
	}
	defer permRows.Close()

	perms := make(map[string][]string)
	for permRows.Next() {
		var res, act string
		permRows.Scan(&res, &act)
		perms[res] = append(perms[res], act)
	}

	return roles, perms, nil
}

func (r PgRepo) List(ctx context.Context) ([]User, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT id, username, email, full_name, department, created_at, updated_at FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.FullName, &u.Department, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, nil
}

func (r PgRepo) Update(ctx context.Context, id string, u *User) error {
	ct, err := r.DB.Exec(ctx,
		`UPDATE users SET username=$1, email=$2, full_name=$3, department=$4, updated_at=now() WHERE id=$5`,
		u.Username, u.Email, u.FullName, u.Department, id,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r PgRepo) Delete(ctx context.Context, id string) error {
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
