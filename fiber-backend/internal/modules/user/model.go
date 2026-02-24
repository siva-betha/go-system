package user

import "time"

// User is the database model.
type User struct {
	ID                string              `json:"id"`
	Username          string              `json:"username"     validate:"required,min=2"`
	Email             string              `json:"email"        validate:"required,email"`
	PasswordHash      string              `json:"-"`
	FullName          string              `json:"full_name"`
	Department        string              `json:"department"`
	Title             string              `json:"title"`
	EmployeeID        string              `json:"employee_id"`
	IsActive          bool                `json:"is_active"`
	Roles             []string            `json:"roles"`
	Permissions       map[string][]string `json:"permissions"`
	LastLogin         *time.Time          `json:"last_login,omitempty"`
	LockedUntil       *time.Time          `json:"locked_until,omitempty"`
	PasswordChangedAt *time.Time          `json:"password_changed_at,omitempty"`
	CreatedAt         time.Time           `json:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at"`
}

// ---------- Request DTOs ----------

// RegisterRequest is the payload for POST /api/auth/register.
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=2"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"full_name"`
}

// LoginRequest is the payload for POST /api/auth/login.
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// ---------- Response DTOs ----------

// AuthResponse is returned on successful register / login / refresh.
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int          `json:"expires_in"` // access token lifetime in seconds
	User         UserResponse `json:"user"`
}

// RefreshRequest is the payload for POST /api/auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// UserResponse is the public representation of a user (no password hash).
type UserResponse struct {
	ID         string    `json:"id"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	FullName   string    `json:"full_name"`
	Department string    `json:"department"`
	Roles      []string  `json:"roles"`
	CreatedAt  time.Time `json:"created_at"`
}

// ToResponse converts a User model to its public response form.
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:         u.ID,
		Username:   u.Username,
		Email:      u.Email,
		FullName:   u.FullName,
		Department: u.Department,
		Roles:      u.Roles,
		CreatedAt:  u.CreatedAt,
	}
}
