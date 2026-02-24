package middleware

import (
	"strings"

	"fiber-backend/internal/auth"

	"github.com/gofiber/fiber/v3"
)

// JWT validates the Authorization header and stores parsed claims in Locals.
// Downstream handlers can access: c.Locals("userID"), c.Locals("email"), c.Locals("role").
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
		c.Locals("userID", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}
