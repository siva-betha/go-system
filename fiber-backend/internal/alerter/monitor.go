package alerter

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sys/windows"
)

type StorageMonitor struct {
	config AlerterConfig
	db     *pgxpool.Pool

	// Channels
	diskStatsChan chan DiskStats
	alertChan     chan StorageAlert
	emailChan     chan EmailMessage

	// Control
	stopChan chan struct{}
	wg       sync.WaitGroup

	// State
	lastAlertSent map[string]time.Time
	mu            sync.RWMutex
	hostname      string
}

func NewStorageMonitor(cfg AlerterConfig, db *pgxpool.Pool) *StorageMonitor {
	hostname, _ := os.Hostname()
	return &StorageMonitor{
		config:        cfg,
		db:            db,
		diskStatsChan: make(chan DiskStats, 100),
		alertChan:     make(chan StorageAlert, 50),
		emailChan:     make(chan EmailMessage, 50),
		stopChan:      make(chan struct{}),
		lastAlertSent: make(map[string]time.Time),
		hostname:      hostname,
	}
}

func (m *StorageMonitor) Start() {
	m.startDiskChecker()
	m.startAlertRouter()
	m.startEmailSender()
	log.Println("Storage monitoring service started")
}

func (m *StorageMonitor) Stop() {
	close(m.stopChan)
	m.wg.Wait()
	log.Println("Storage monitoring service stopped")
}

func (m *StorageMonitor) startDiskChecker() {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		ticker := time.NewTicker(m.config.CheckInterval)
		defer ticker.Stop()

		// Initial check
		m.checkAllDisks()

		for {
			select {
			case <-ticker.C:
				m.checkAllDisks()
			case <-m.stopChan:
				return
			}
		}
	}()
}

func (m *StorageMonitor) checkAllDisks() {
	for component, path := range m.config.Paths {
		stats, err := m.getDiskUsage(path)
		if err != nil {
			log.Printf("Failed to get disk usage for %s (%s): %v", component, path, err)
			continue
		}
		stats.Component = component
		m.processDiskStats(stats)
	}
}

func (m *StorageMonitor) getDiskUsage(path string) (DiskStats, error) {
	var freeBytes, totalBytes, totalFreeBytes uint64

	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return DiskStats{}, err
	}

	err = windows.GetDiskFreeSpaceEx(pathPtr, &freeBytes, &totalBytes, &totalFreeBytes)
	if err != nil {
		return DiskStats{}, err
	}

	usedBytes := totalBytes - freeBytes
	usedPercent := 0.0
	if totalBytes > 0 {
		usedPercent = float64(usedBytes) / float64(totalBytes) * 100
	}

	return DiskStats{
		Path:        path,
		TotalBytes:  totalBytes,
		UsedBytes:   usedBytes,
		FreeBytes:   freeBytes,
		UsedPercent: usedPercent,
		Timestamp:   time.Now(),
	}, nil
}

func (m *StorageMonitor) processDiskStats(stats DiskStats) {
	// Log healthy stats for observability
	// log.Printf("Disk Usage [%s]: %.1f%% used of %v", stats.Component, stats.UsedPercent, stats.TotalBytes)

	// In a real system, we might push these to InfluxDB here too

	// Check thresholds
	level := ""
	if stats.UsedPercent >= m.config.EmergencyPercent {
		level = "emergency"
	} else if stats.UsedPercent >= m.config.CriticalPercent {
		level = "critical"
	} else if stats.UsedPercent >= m.config.WarningPercent {
		level = "warning"
	}

	if level != "" {
		alert := StorageAlert{
			Level:       level,
			Component:   stats.Component,
			Path:        stats.Path,
			TotalBytes:  stats.TotalBytes,
			UsedBytes:   stats.UsedBytes,
			FreeBytes:   stats.FreeBytes,
			UsedPercent: stats.UsedPercent,
			Timestamp:   time.Now(),
			Hostname:    m.hostname,
		}

		select {
		case m.alertChan <- alert:
		default:
			log.Printf("Alert channel full, dropping alert for %s", stats.Component)
		}
	}
}

func (m *StorageMonitor) startAlertRouter() {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		for {
			select {
			case alert := <-m.alertChan:
				if m.canSendAlert(alert) {
					m.prepareEmail(alert)
				}

				if alert.Level == "emergency" && m.config.AutoCleanup {
					m.triggerEmergencyAction(alert)
				}
			case <-m.stopChan:
				return
			}
		}
	}()
}

func (m *StorageMonitor) canSendAlert(alert StorageAlert) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s_%s", alert.Component, alert.Level)
	lastSent, exists := m.lastAlertSent[key]

	if !exists {
		m.lastAlertSent[key] = time.Now()
		return true
	}

	interval := 24 * time.Hour
	switch alert.Level {
	case "warning":
		interval = 24 * time.Hour
	case "critical":
		interval = 1 * time.Hour
	case "emergency":
		interval = 15 * time.Minute
	}

	if time.Since(lastSent) >= interval {
		m.lastAlertSent[key] = time.Now()
		return true
	}

	return false
}
