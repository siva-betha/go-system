package influx

import "github.com/gofiber/fiber/v3"

func Routes(r fiber.Router, h Handler) {
	r.Get("/health", h.Health)
	r.Get("/range", h.QueryRange)
}
