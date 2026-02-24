package approval

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	Service *Service
}

func (h Handler) ListPending(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	approvals, err := h.Service.Repo.ListPending(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(approvals)
}

func (h Handler) Review(c fiber.Ctx) error {
	id := c.Params("id")

	var req struct {
		Action      string  `json:"action"` // approve, reject
		ReviewNotes *string `json:"review_notes"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	reviewerID, _ := c.Locals("user_id").(string)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var err error
	if req.Action == "approve" {
		err = h.Service.Approve(ctx, id, reviewerID, req.ReviewNotes)
	} else if req.Action == "reject" {
		err = h.Service.Reject(ctx, id, reviewerID, req.ReviewNotes)
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "invalid action"})
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}
