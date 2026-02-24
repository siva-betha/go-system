package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

func Security(app *fiber.App) {
	app.Use(helmet.New())

	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 60,
	}))
}
