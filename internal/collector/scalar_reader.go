package collector

import (
	"log"
	"sync"
	"time"

	"github.com/robinson/gouads"
)

type ScalarReader struct {
	conn          *gouads.Connection
	symbolManager *SymbolManager
	readGroups    []*ReadGroup

	// Channels
	dataChan    chan *DataPoint
	controlChan chan struct{}

	// Batching
	batchSize    int
	currentBatch []*DataPoint
	batchMu      sync.Mutex
}

type DataPoint struct {
	Timestamp time.Time
	Symbol    string
	Target    string // mapped name
	Value     interface{}
	Quality   int
	Source    string // PLC ID
}

func NewScalarReader(conn *gouads.Connection, sm *SymbolManager) *ScalarReader {
	return &ScalarReader{
		conn:          conn,
		symbolManager: sm,
		dataChan:      make(chan *DataPoint, 10000),
		controlChan:   make(chan struct{}),
		batchSize:     100,
	}
}

func (r *ScalarReader) Start() {
	for _, group := range r.readGroups {
		go r.readGroupLoop(group)
	}

	// Start batch flusher if needed, but here we'll just handle it in the loops
}

func (r *ScalarReader) readGroupLoop(group *ReadGroup) {
	ticker := time.NewTicker(group.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Read all symbols in group in one ADS call
			values, err := r.conn.ReadSymbols(group.Symbols)
			if err != nil {
				log.Printf("Error reading group %s: %v", group.Name, err)
				continue
			}

			timestamp := time.Now()

			// Process each value
			for _, sym := range group.Symbols {
				if val, ok := values[sym]; ok {
					// Check mapping
					target := sym
					if info, ok := r.symbolManager.symbols[sym]; ok && info.TargetName != "" {
						target = info.TargetName
					}

					point := &DataPoint{
						Timestamp: timestamp,
						Symbol:    sym,
						Target:    target,
						Value:     val,
						Quality:   100, // Good quality
					}

					// Add to batch
					r.addToBatch(point)
				}
			}

		case <-r.controlChan:
			return
		}
	}
}

func (r *ScalarReader) addToBatch(point *DataPoint) {
	r.batchMu.Lock()
	defer r.batchMu.Unlock()

	r.currentBatch = append(r.currentBatch, point)

	if len(r.currentBatch) >= r.batchSize {
		r.flushBatch()
	}
}

func (r *ScalarReader) flushBatch() {
	if len(r.currentBatch) == 0 {
		return
	}

	// Send batch to channel
	for _, point := range r.currentBatch {
		select {
		case r.dataChan <- point:
		default:
			log.Printf("Data channel full, dropping point")
		}
	}

	r.currentBatch = nil
}
