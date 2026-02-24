package plcengine

import (
	"fmt"
	"time"
)

// PrioritizedWriter handles asynchronous write requests with priority levels
type PrioritizedWriter struct {
	engine *PLCReadWriteEngine

	// Channels for each priority level (0=High, 1=Medium, 2=Low)
	queues [3]chan WriteRequest

	stopChan chan struct{}
}

func NewPrioritizedWriter(engine *PLCReadWriteEngine) *PrioritizedWriter {
	return &PrioritizedWriter{
		engine: engine,
		queues: [3]chan WriteRequest{
			make(chan WriteRequest, 100),  // High
			make(chan WriteRequest, 500),  // Medium
			make(chan WriteRequest, 1000), // Low
		},
		stopChan: make(chan struct{}),
	}
}

func (w *PrioritizedWriter) Start() {
	go w.processor()
}

func (w *PrioritizedWriter) Stop() {
	close(w.stopChan)
}

func (w *PrioritizedWriter) Submit(req WriteRequest) error {
	var qIdx int
	if req.Priority >= 8 {
		qIdx = 0 // High
	} else if req.Priority >= 4 {
		qIdx = 1 // Medium
	} else {
		qIdx = 2 // Low
	}

	select {
	case w.queues[qIdx] <- req:
		return nil
	default:
		return fmt.Errorf("priority queue %d is full", qIdx)
	}
}

func (w *PrioritizedWriter) processor() {
	for {
		select {
		case <-w.stopChan:
			return
		default:
			// Check priorities in strict order: High > Medium > Low
			select {
			case req := <-w.queues[0]:
				w.execute(req)
			case req := <-w.queues[1]:
				w.execute(req)
			case req := <-w.queues[2]:
				w.execute(req)
			case <-time.After(10 * time.Millisecond):
				// No requests, idle loop
			}
		}
	}
}

func (w *PrioritizedWriter) execute(req WriteRequest) {
	resp := WriteResponse{
		ID:        req.ID,
		Timestamp: time.Now(),
	}

	// 1. Execute Write
	err := w.engine.WriteSymbol(req.MachineID, req.Symbol, req.Value)
	if err != nil {
		resp.Success = false
		resp.Error = err.Error()
	} else {
		// 2. Read-after-write verification (if enabled)
		if req.RequireAck {
			time.Sleep(20 * time.Millisecond) // Wait for PLC cycle
			val, rErr := w.engine.ReadSymbol(req.MachineID, req.Symbol)
			if rErr != nil {
				resp.Success = false
				resp.Error = "verification read failed: " + rErr.Error()
			} else {
				// Simple equality check (could be improved with type-specific comparison)
				if fmt.Sprintf("%v", val.Value) == fmt.Sprintf("%v", req.Value) {
					resp.Success = true
				} else {
					resp.Success = false
					resp.Error = fmt.Sprintf("verification failed: expected %v, got %v", req.Value, val.Value)
				}
			}
		} else {
			resp.Success = true
		}
	}

	// 3. Notify caller
	if req.ResponseChan != nil {
		select {
		case req.ResponseChan <- resp:
		case <-time.After(500 * time.Millisecond):
			// Drop if response chan blocked to avoid stalling the engine
		}
	}

	// Also send to global engine confirm channel (if defined)
	select {
	case w.engine.writeConfirm <- resp:
	default:
	}
}
