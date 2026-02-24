package approval

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
)

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) Request(c fiber.Ctx, action, resource string, resourceID *string, data any) error {
	requestedBy, _ := c.Locals("user_id").(string)

	expiresAt := time.Now().Add(24 * time.Hour) // 24h default expiry

	a := &PendingApproval{
		RequestedBy: requestedBy,
		Action:      action,
		Resource:    resource,
		ResourceID:  resourceID,
		Data:        data,
		Status:      StatusPending,
		ExpiresAt:   &expiresAt,
	}

	return s.Repo.Create(context.Background(), a)
}

func (s *Service) Approve(ctx context.Context, id string, reviewerID string, notes *string) error {
	return s.Repo.UpdateStatus(ctx, id, StatusApproved, reviewerID, notes)
}

func (s *Service) Reject(ctx context.Context, id string, reviewerID string, notes *string) error {
	return s.Repo.UpdateStatus(ctx, id, StatusRejected, reviewerID, notes)
}
