package auth

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

type AuthMiddleware struct {
	jwtSecret   []byte
	tokenExpiry time.Duration
}

func NewAuthMiddleware(secret string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret:   []byte(secret),
		tokenExpiry: 24 * time.Hour,
	}
}

// Authenticate validates the JWT token in the Authorization header
func (m *AuthMiddleware) Authenticate() fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{"error": "missing authorization header"})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(401).JSON(fiber.Map{"error": "invalid authorization format"})
		}

		token, err := jwt.ParseWithClaims(parts[1], &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return m.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			return c.Status(401).JSON(fiber.Map{"error": "invalid or expired token"})
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "invalid token claims"})
		}

		// Store user context
		c.Locals("user", claims)
		c.Locals("user_id", claims.UserID)
		c.Locals("roles", claims.Roles)

		return c.Next()
	}
}

// RequirePermission checks if the user has a specific permission for a resource
func (m *AuthMiddleware) RequirePermission(resource, action string) fiber.Handler {
	return func(c fiber.Ctx) error {
		user, ok := c.Locals("user").(*Claims)
		if !ok {
			return c.Status(500).JSON(fiber.Map{"error": "user context missing"})
		}

		if !m.hasPermission(user, resource, action) {
			return c.Status(403).JSON(fiber.Map{
				"error": "insufficient permissions",
				"required": fiber.Map{
					"resource": resource,
					"action":   action,
				},
			})
		}

		return c.Next()
	}
}

func (m *AuthMiddleware) hasPermission(user *Claims, resource, action string) bool {
	// Admin has all permissions
	for _, role := range user.Roles {
		if role == "admin" {
			return true
		}
	}

	permissions, exists := user.Permissions[resource]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == action || p == "*" {
			return true
		}
	}

	return false
}

// RequireRole checks if the user has one of the specified roles
func (m *AuthMiddleware) RequireRole(roles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		userRoles, ok := c.Locals("roles").([]string)
		if !ok {
			return c.Status(403).JSON(fiber.Map{"error": "insufficient role"})
		}

		for _, required := range roles {
			for _, userRole := range userRoles {
				if required == userRole {
					return c.Next()
				}
			}
		}

		return c.Status(403).JSON(fiber.Map{
			"error":          "insufficient role",
			"required_roles": roles,
		})
	}
}
