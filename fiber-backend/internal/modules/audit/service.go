package audit

import (
	"context"

	"github.com/gofiber/fiber/v3"
)

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) Log(c fiber.Ctx, action, resource string, resourceID *string, oldVal, newVal any) {
	userIDStr, _ := c.Locals("user_id").(string)
	var userID *string
	if userIDStr != "" {
		userID = &userIDStr
	}

	log := &AuditLog{
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		OldValues:  oldVal,
		NewValues:  newVal,
		IPAddress:  c.IP(),
		UserAgent:  c.Get("User-Agent"),
	}

	// We don't want to block the request for audit logging, but we should handle errors.
	// For production, consider using a background queue/worker.
	go func() {
		ctx := context.Background()
		_ = s.Repo.Create(ctx, log)
	}()
}
