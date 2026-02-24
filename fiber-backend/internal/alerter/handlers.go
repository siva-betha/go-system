package alerter

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v3"
)

func (m *StorageMonitor) RegisterRoutes(router fiber.Router) {
	group := router.Group("/storage")
	group.Get("/status", m.HandleStatus)
	group.Get("/alerts", m.HandleAlerts)
	group.Post("/alerts/:id/acknowledge", m.HandleAcknowledge)
}

func (m *StorageMonitor) HandleStatus(c fiber.Ctx) error {
	// In a real system, we'd store the latest stats in a map protected by a mutex
	// For now, we'll perform an on-demand check for simplicity in this demo
	stats := make(map[string]DiskStats)
	for component, path := range m.config.Paths {
		s, _ := m.getDiskUsage(path)
		s.Component = component
		stats[component] = s
	}

	return c.JSON(fiber.Map{
		"status": "healthy",
		"disks":  stats,
	})
}

func (m *StorageMonitor) HandleAlerts(c fiber.Ctx) error {
	if m.db == nil {
		return c.Status(http.StatusServiceUnavailable).SendString("DB not connected")
	}

	query := `SELECT id, level, component, path, used_percent, free_bytes, total_bytes, acknowledged, created_at 
	          FROM storage_alerts ORDER BY created_at DESC LIMIT 50`

	rows, err := m.db.Query(context.Background(), query)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var alerts []map[string]interface{}
	for rows.Next() {
		var id, level, component, path string
		var usedPercent float64
		var freeBytes, totalBytes uint64
		var acknowledged bool
		var createdAt interface{}

		err := rows.Scan(&id, &level, &component, &path, &usedPercent, &freeBytes, &totalBytes, &acknowledged, &createdAt)
		if err != nil {
			continue
		}

		alerts = append(alerts, map[string]interface{}{
			"id":           id,
			"level":        level,
			"component":    component,
			"path":         path,
			"used_percent": usedPercent,
			"free_bytes":   freeBytes,
			"total_bytes":  totalBytes,
			"acknowledged": acknowledged,
			"created_at":   createdAt,
		})
	}

	return c.JSON(alerts)
}

func (m *StorageMonitor) HandleAcknowledge(c fiber.Ctx) error {
	if m.db == nil {
		return c.Status(http.StatusServiceUnavailable).SendString("DB not connected")
	}

	id := c.Params("id")
	user := "admin" // In a real system, get from JWT context

	query := `UPDATE storage_alerts SET acknowledged = true, acknowledged_by = $1, acknowledged_at = NOW() WHERE id = $2`
	_, err := m.db.Exec(context.Background(), query, user, id)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "acknowledged"})
}
