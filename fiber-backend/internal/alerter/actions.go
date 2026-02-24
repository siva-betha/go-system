package alerter

import (
	"fmt"
	"log"
)

func (m *StorageMonitor) triggerEmergencyAction(alert StorageAlert) {
	log.Printf("EMERGENCY triggered for %s at %.1f%%. Attempting auto-cleanup...", alert.Component, alert.UsedPercent)

	switch alert.Component {
	case "influxdb":
		m.forceInfluxDBCleanup()
	case "kafka":
		m.triggerKafkaCleanup()
	case "postgresql":
		m.vacuumPostgreSQL()
	case "system", "logs":
		m.cleanOldLogs()
	}

	// Notify about action taken
	m.emailChan <- EmailMessage{
		To:      m.config.Email.To,
		Subject: fmt.Sprintf("[ACTION] Emergency cleanup triggered for %s", alert.Component),
		Body:    fmt.Sprintf("Component %s reached %.1f%% usage. Emergency cleanup was automatically triggered.", alert.Component, alert.UsedPercent),
	}
}

func (m *StorageMonitor) forceInfluxDBCleanup() {
	log.Println("Action: Forcing InfluxDB retention cleanup (Mock)")
	// TODO: Integrate with influx.Handler or use direct client to DROP SHARDS
}

func (m *StorageMonitor) triggerKafkaCleanup() {
	log.Println("Action: Triggering Kafka log compaction (Mock)")
	// TODO: Adjust retention.ms or retention.bytes temporarily via AdminClient
}

func (m *StorageMonitor) vacuumPostgreSQL() {
	log.Println("Action: Running PostgreSQL VACUUM (Mock)")
	// TODO: Execute "VACUUM ANALYZE" or "VACUUM FULL" if risky
}

func (m *StorageMonitor) cleanOldLogs() {
	log.Println("Action: Deleting old application logs (Mock)")
	// TODO: os.Remove old files in log directory
}
