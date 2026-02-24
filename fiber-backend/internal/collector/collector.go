package collector

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"fiber-backend/internal/kafka"
	"fiber-backend/internal/plcengine"
	"fiber-backend/internal/streamer"
)

type Collector struct {
	engine   *plcengine.PLCReadWriteEngine
	hub      *streamer.StreamHub
	dataChan chan plcengine.PLCValue
	stopChan chan struct{}
	wg       sync.WaitGroup
	mu       sync.RWMutex
	machines map[string]*MachineCollector
}

type MachineCollector struct {
	config MachineConfig
	stop   chan struct{}
}

func NewCollector(engine *plcengine.PLCReadWriteEngine, hub *streamer.StreamHub) *Collector {
	return &Collector{
		engine:   engine,
		hub:      hub,
		machines: make(map[string]*MachineCollector),
	}
}

func (c *Collector) Start(configs []MachineConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Use engine's data channel
	c.dataChan = make(chan plcengine.PLCValue, 10000)
	c.stopChan = make(chan struct{})

	// 0. Connect engine to PLCs
	var engineConfigs []plcengine.MachineConfig
	for _, cfg := range configs {
		engineConfigs = append(engineConfigs, plcengine.MachineConfig{
			ID:       cfg.ID,
			IP:       cfg.IP,
			AmsNetID: cfg.AmsNetID,
			Port:     cfg.Port,
		})
	}
	if err := c.engine.Start(engineConfigs); err != nil {
		return err
	}

	// 1. Start Kafka worker pool (4 workers)
	for i := 0; i < 4; i++ {
		c.wg.Add(1)
		go c.kafkaWorker(i)
	}

	// 2. Start Streamer broadcaster
	c.wg.Add(1)
	go c.streamerWorker()

	// 3. Start poller goroutine per chamber (using engine)
	for _, cfg := range configs {
		for _, chamberCfg := range cfg.Chambers {
			c.wg.Add(1)
			go c.runChamberPoller(cfg.ID, chamberCfg)
		}
	}

	return nil
}

func (c *Collector) runChamberPoller(machineID string, cfg ChamberConfig) {
	defer c.wg.Done()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	var lastErrorLog time.Time

	symbols := make([]string, len(cfg.Symbols))
	for i, s := range cfg.Symbols {
		symbols[i] = s.Name
	}

	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			vals, err := c.engine.ReadSymbols(machineID, symbols)
			if err != nil {
				// Throttle error logging to avoid flooding
				if time.Since(lastErrorLog) > 10*time.Second {
					log.Printf("Poller error on machine %s, chamber %s: %v", machineID, cfg.Name, err)
					lastErrorLog = time.Now()
				}
				continue
			}

			// Broadcast each symbol update if needed, or group by chamber
			// For now, we group by chamber for the streamer
			data := streamer.BroadcastMsg{
				Type:      streamer.MsgTypeData,
				MachineID: machineID,
				ChamberID: cfg.ID,
				Data:      make(map[string]interface{}),
				Timestamp: time.Now(),
			}

			for sym, v := range vals {
				data.Data[sym] = v.Value
				// Also send individual symbols to the main dataChan for Kafka
				c.dataChan <- *v
			}

			c.hub.Broadcast(data)
		}
	}
}

func (c *Collector) streamerWorker() {
	defer c.wg.Done()
	// The StreamHub runs its own loop, we just need to manage its stop signal if needed
	// or perform additional data transformations here.
	log.Println("Streamer worker started")
	<-c.stopChan
	log.Println("Streamer worker stopped")
}

func (c *Collector) kafkaWorker(workerID int) {
	defer c.wg.Done()

	batch := make([]plcengine.PLCValue, 0, 1000)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	producer := kafka.NewProducer("localhost:9092", "plc-data")
	defer producer.Close()

	for {
		select {
		case data, ok := <-c.dataChan:
			if !ok {
				if len(batch) > 0 {
					c.flushBatch(producer, batch, workerID)
				}
				return
			}
			batch = append(batch, data)
			if len(batch) >= 1000 {
				c.flushBatch(producer, batch, workerID)
				batch = make([]plcengine.PLCValue, 0, 1000)
			}

		case <-ticker.C:
			if len(batch) > 0 {
				c.flushBatch(producer, batch, workerID)
				batch = make([]plcengine.PLCValue, 0, 1000)
			}

		case <-c.stopChan:
			if len(batch) > 0 {
				c.flushBatch(producer, batch, workerID)
			}
			return
		}
	}
}

func (c *Collector) flushBatch(producer kafka.Producer, batch []plcengine.PLCValue, workerID int) {
	messages := make([]kafka.Message, len(batch))
	for i, data := range batch {
		key := fmt.Sprintf("%s-%s", data.Source, data.Symbol)
		value, _ := json.Marshal(data)
		messages[i] = kafka.Message{
			Key:   []byte(key),
			Value: value,
		}
	}

	if err := producer.ProduceBatch(messages); err != nil {
		log.Printf("Worker %d: Kafka error: %v", workerID, err)
	}
}

func (c *Collector) Stop() error {
	log.Println("Stopping collector...")
	close(c.stopChan)

	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All goroutines stopped")
	case <-time.After(10 * time.Second):
		return fmt.Errorf("stop timeout")
	}

	return nil
}
