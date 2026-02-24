package user

import "time"

// User is the database model.
type User struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"         validate:"required,min=2"`
	Email        string    `json:"email"        validate:"required,email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ---------- Request DTOs ----------

// RegisterRequest is the payload for POST /api/auth/register.
type RegisterRequest struct {
	Name     string `json:"name"     validate:"required,min=2"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginRequest is the payload for POST /api/auth/login.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
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
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// ToResponse converts a User model to its public response form.
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}
