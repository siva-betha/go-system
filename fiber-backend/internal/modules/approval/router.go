package approval

import (
	"github.com/gofiber/fiber/v3"
)

func Routes(router fiber.Router, h *Handler) {
	router.Get("/pending", h.ListPending)
	router.Post("/review/:id", h.Review)
}
