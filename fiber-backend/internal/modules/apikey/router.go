package apikey

import (
	"github.com/gofiber/fiber/v3"
)

func Routes(router fiber.Router, h *Handler) {
	router.Get("/", h.List)
	router.Post("/", h.Create)
	router.Delete("/:id", h.Delete)
}
