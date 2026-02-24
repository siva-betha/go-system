package machine_config

import (
	"fiber-backend/internal/modules/approval"
	"fiber-backend/internal/modules/audit"

	"github.com/gofiber/fiber/v3"
)

func Register(router fiber.Router, repo Repository, approvalSvc *approval.Service, auditSvc *audit.Service) {
	handler := &Handler{
		Repo:            repo,
		ApprovalService: approvalSvc,
	}
	group := router.Group("/config")

	group.Get("/machines", handler.GetMachines)
	group.Post("/machines", handler.SaveMachines)
}
