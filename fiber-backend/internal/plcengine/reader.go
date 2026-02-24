package plcengine

import (
	"log"
	"time"
)

// BatchReader manages continuous data collection from a set of PLC symbols
type BatchReader struct {
	conn     *PLCConnection
	symbols  []string
	interval time.Duration
	dataChan chan<- PLCValue
	stopChan chan struct{}
}

func NewBatchReader(conn *PLCConnection, symbols []string, interval time.Duration, dataChan chan<- PLCValue) *BatchReader {
	return &BatchReader{
		conn:     conn,
		symbols:  symbols,
		interval: interval,
		dataChan: dataChan,
		stopChan: make(chan struct{}),
	}
}

func (r *BatchReader) Start() {
	go r.run()
}

func (r *BatchReader) Stop() {
	close(r.stopChan)
}

func (r *BatchReader) run() {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-r.stopChan:
			return
		case <-ticker.C:
			// Execute efficient batch read via connection multiplexer
			values, err := r.conn.ReadSymbols(r.symbols)
			if err != nil {
				log.Printf("Batch read error on machine %s: %v", r.conn.MachineID, err)
				continue
			}

			now := time.Now()
			for sym, val := range values {
				r.dataChan <- PLCValue{
					Symbol:    sym,
					Value:     val,
					Timestamp: now,
					Source:    r.conn.MachineID,
					Quality:   100,
				}
			}
		}
	}
}

// Subscription handles ADS notification based updates (Push instead of Pull)
type Subscription struct {
	conn     *PLCConnection
	symbols  []string
	dataChan chan<- PLCValue
	stopChan chan struct{}
}

// In a real implementation, this would use ADS DLL/Library hooks for callbacks.
// For now, we simulate the interface needed for notification handling.
func NewSubscription(conn *PLCConnection, symbols []string, dataChan chan<- PLCValue) *Subscription {
	return &Subscription{
		conn:     conn,
		symbols:  symbols,
		dataChan: dataChan,
		stopChan: make(chan struct{}),
	}
}

func (s *Subscription) Start() {
	// In real ADS, this would call AdsAddDeviceNotification
	log.Printf("Subscription started for %d symbols on machine %s", len(s.symbols), s.conn.MachineID)
}

func (s *Subscription) Stop() {
	close(s.stopChan)
}
