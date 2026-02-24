package middleware

import (
	"strings"

	"fiber-backend/internal/auth"

	"github.com/gofiber/fiber/v3"
)

// JWT validates the Authorization header and stores parsed claims in Locals.
func JWT() fiber.Handler {
	return func(c fiber.Ctx) error {
		h := c.Get("Authorization")
		if h == "" {
			return fiber.ErrUnauthorized
		}

		tokenStr := h
		if strings.HasPrefix(h, "Bearer ") {
			tokenStr = strings.TrimPrefix(h, "Bearer ")
		}

		claims, err := auth.ParseToken(tokenStr)
		if err != nil {
			return fiber.ErrUnauthorized
		}

		// Store claims in context for downstream handlers
		c.Locals("user", claims)
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("roles", claims.Roles)
		c.Locals("permissions", claims.Permissions)

		return c.Next()
	}
}
