package alerter

import (
	"time"
)

type EmailConfig struct {
	SMTPHost  string   `json:"smtp_host"`
	SMTPPort  int      `json:"smtp_port"`
	Username  string   `json:"username"`
	Password  string   `json:"password"`
	From      string   `json:"from"`
	To        []string `json:"to"`
	EnableSSL bool     `json:"enable_ssl"`
	AuthType  string   `json:"auth_type"` // "plain", "login", "none"
}

type StorageAlert struct {
	ID          string    `json:"id"`
	Level       string    `json:"level"`     // "warning", "critical", "emergency"
	Component   string    `json:"component"` // "influxdb", "postgresql", "kafka", "system"
	Path        string    `json:"path"`
	TotalBytes  uint64    `json:"total_bytes"`
	UsedBytes   uint64    `json:"used_bytes"`
	FreeBytes   uint64    `json:"free_bytes"`
	UsedPercent float64   `json:"used_percent"`
	Timestamp   time.Time `json:"timestamp"`
	Hostname    string    `json:"hostname"`
	Action      string    `json:"action"` // "notify", "cleanup", "shutdown"
}

type DiskStats struct {
	Component   string    `json:"component"`
	Path        string    `json:"path"`
	TotalBytes  uint64    `json:"total_bytes"`
	UsedBytes   uint64    `json:"used_bytes"`
	FreeBytes   uint64    `json:"free_bytes"`
	UsedPercent float64   `json:"used_percent"`
	InodesFree  uint64    `json:"inodes_free"` // Primarily for Linux, but kept for compatibility
	Timestamp   time.Time `json:"timestamp"`
}

type EmailMessage struct {
	To          []string `json:"to"`
	Subject     string   `json:"subject"`
	Body        string   `json:"body"`
	Priority    int      `json:"priority"` // 1=high, 5=low
	RetryCount  int      `json:"retry_count"`
	Attachments []string `json:"attachments"`
}

type ActionRequest struct {
	Alert StorageAlert `json:"alert"`
}

type AlerterConfig struct {
	CheckInterval    time.Duration     `json:"check_interval"`
	WarningPercent   float64           `json:"warning_percent"`
	CriticalPercent  float64           `json:"critical_percent"`
	EmergencyPercent float64           `json:"emergency_percent"`
	Email            EmailConfig       `json:"email"`
	Paths            map[string]string `json:"paths"`
	AutoCleanup      bool              `json:"auto_cleanup"`
}
