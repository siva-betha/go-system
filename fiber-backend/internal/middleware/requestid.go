package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func RequestID() fiber.Handler {
	return func(c fiber.Ctx) error {
		id := uuid.NewString()
		c.Set("X-Request-ID", id)
		c.Locals("reqid", id)
		return c.Next()
	}
}
