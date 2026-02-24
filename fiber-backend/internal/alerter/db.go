package alerter

import (
	"context"
	"log"
	"time"
)

func (m *StorageMonitor) recordAlert(alert StorageAlert) error {
	if m.db == nil {
		return nil
	}

	query := `
		INSERT INTO storage_alerts (
			level, component, path, used_percent, free_bytes, total_bytes, hostname, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var id string
	err := m.db.QueryRow(context.Background(), query,
		alert.Level,
		alert.Component,
		alert.Path,
		alert.UsedPercent,
		alert.FreeBytes,
		alert.TotalBytes,
		alert.Hostname,
		time.Now(),
	).Scan(&id)

	if err != nil {
		log.Printf("Failed to record alert in DB: %v", err)
		return err
	}

	return nil
}

func (m *StorageMonitor) logCleanupAction(alertID string, actionType string, success bool, errStr string, freedBytes uint64) {
	if m.db == nil {
		return
	}

	query := `
		INSERT INTO cleanup_actions (
			alert_id, action_type, success, error_message, freed_bytes, created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := m.db.Exec(context.Background(), query,
		alertID,
		actionType,
		success,
		errStr,
		freedBytes,
		time.Now(),
	)

	if err != nil {
		log.Printf("Failed to log cleanup action in DB: %v", err)
	}
}
