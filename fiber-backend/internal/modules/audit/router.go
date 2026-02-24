package audit

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	Repo Repository
}

func (h Handler) List(c fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "100"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logs, err := h.Repo.List(ctx, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(logs)
}

func Routes(router fiber.Router, repo Repository) {
	h := Handler{Repo: repo}
	router.Get("/list", h.List)
}
