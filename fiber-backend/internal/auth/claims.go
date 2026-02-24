package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

// Claims defines the structure of JWT claims used for RBAC
type Claims struct {
	UserID       string              `json:"user_id"`
	Username     string              `json:"username"`
	Roles        []string            `json:"roles"`
	Permissions  map[string][]string `json:"permissions"`
	ChamberScope []string            `json:"chamber_scope"`
	jwt.RegisteredClaims
}

// User represents a user in the RBAC system
type User struct {
	ID          string              `json:"id"`
	Username    string              `json:"username"`
	Email       string              `json:"email"`
	FullName    string              `json:"full_name"`
	Department  string              `json:"department"`
	Title       string              `json:"title"`
	EmployeeID  string              `json:"employee_id"`
	IsActive    bool                `json:"is_active"`
	Roles       []string            `json:"roles"`
	Permissions map[string][]string `json:"permissions"`
	LockedUntil *string             `json:"locked_until,omitempty"`
}

// PublicUser represents a user with sensitive info stripped
type PublicUser struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	FullName   string `json:"full_name"`
	Department string `json:"department"`
}
