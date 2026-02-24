package machine_config

import (
	"context"
	"fmt"
	"time"

	"fiber-backend/internal/modules/approval"
	"fiber-backend/internal/validator"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	Repo            Repository
	ApprovalService *approval.Service
}

// GetMachines godoc
// @Summary     Get all configured PLC machines
// @Description Returns a list of all PLC machines with their chambers and symbols
// @Tags        config
// @Produce     json
// @Success     200 {array}  Machine
// @Failure     500 {object} map[string]interface{} "Internal server error"
// @Router      /config/machines [get]
func (h Handler) GetMachines(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	machines, err := h.Repo.GetMachines(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(machines)
}

// SaveMachines godoc
// @Summary     Save PLC machine configurations
// @Description Overwrites the current machine configurations with the provided list
// @Tags        config
// @Accept      json
// @Produce     json
// @Param       machines body     []Machine true "List of machines to configure"
// @Success     200      {object} map[string]interface{} "Success message"
// @Failure     400      {object} map[string]interface{} "Validation error"
// @Failure     500      {object} map[string]interface{} "Internal server error"
// @Router      /config/machines [post]
func (h Handler) SaveMachines(c fiber.Ctx) error {
	var machines []Machine
	if err := c.Bind().Body(&machines); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	// Basic validation for each machine
	for _, m := range machines {
		if err := validator.V.Struct(m); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": fmt.Sprintf("Machine %s validation failed: %v", m.Name, err)})
		}
	}

	// If user is not admin, or we want forced dual-auth for this sensitive action:
	// For this implementation, we'll route it to approvals if it's a significant change.
	err := h.ApprovalService.Request(c, "UPDATE", "machine_config", nil, machines)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to queue approval request: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Configuration change requested. Awaiting administrator approval.",
		"status":  "pending_approval",
	})
}
