package collector

import (
	"log"
	"sync"
	"time"

	"github.com/robinson/gouads"
)

type OESArrayReader struct {
	conn           *gouads.Connection
	symbolManager  *SymbolManager
	arraySymbol    string   // "MAIN_OES.FlatArray"
	readyFlag      string   // "MAIN_OES.FullFrameReady"
	metadataFields []string // Fields to capture with each spectrum
	arraySize      int      // 2048

	// Channels
	spectrumChan chan *Spectrum
	controlChan  chan struct{}

	// State
	lastSequence uint32
	missedFrames uint64
	mu           sync.Mutex
}

type Spectrum struct {
	Timestamp      time.Time
	Wavelengths    []float64 // Calculated from pixel mapping
	Intensities    []uint16  // Raw ADC values
	Metadata       map[string]interface{}
	SequenceNum    uint32
	AcquisitionCtr uint16
	SegmentCtr     uint16
	ExposureTime   int32
	PixelSize      uint16
	BinFactor      uint16
}

func NewOESArrayReader(conn *gouads.Connection, sm *SymbolManager) *OESArrayReader {
	return &OESArrayReader{
		conn:          conn,
		symbolManager: sm,
		arraySymbol:   "MAIN_OES.FlatArray",
		readyFlag:     "MAIN_OES.FullFrameReady",
		arraySize:     2048,
		spectrumChan:  make(chan *Spectrum, 100),
		controlChan:   make(chan struct{}),
	}
}

func (r *OESArrayReader) Start() {
	go r.readLoop()
}

func (r *OESArrayReader) readLoop() {
	ticker := time.NewTicker(1 * time.Millisecond) // High-frequency check
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check if new spectrum is ready
			ready, err := r.conn.ReadBool(r.readyFlag)
			if err != nil {
				continue
			}

			if ready {
				if err := r.readSpectrum(); err != nil {
					log.Printf("Error reading spectrum: %v", err)
				}
				// Clear the flag (write back false)
				r.conn.WriteBool(r.readyFlag, false)
			}

		case <-r.controlChan:
			return
		}
	}
}

func (r *OESArrayReader) readSpectrum() error {
	// Read the entire array in one ADS call
	rawIntensities, err := r.conn.ReadUint16Array(r.arraySymbol, r.arraySize)
	if err != nil {
		r.missedFrames++
		return err
	}

	// Read metadata fields
	metadata := make(map[string]interface{})
	for _, field := range r.metadataFields {
		val, err := r.readField(field)
		if err == nil {
			metadata[field] = val
		}
	}

	// Get sequence info
	acqCtr, _ := r.conn.ReadUint16("MAIN_OES.Acquisition_counter")
	segCtr, _ := r.conn.ReadUint16("MAIN_OES.Segment_counter")

	// Create spectrum
	spectrum := &Spectrum{
		Timestamp:      time.Now(),
		Intensities:    rawIntensities,
		Metadata:       metadata,
		SequenceNum:    r.lastSequence + 1,
		AcquisitionCtr: acqCtr,
		SegmentCtr:     segCtr,
	}

	// Send to channel (non-blocking)
	select {
	case r.spectrumChan <- spectrum:
	default:
		log.Printf("Spectrum channel full, dropping spectrum %d", spectrum.SequenceNum)
	}

	r.lastSequence++
	return nil
}

func (r *OESArrayReader) readField(name string) (interface{}, error) {
	// Simple field reader for metadata
	return r.conn.ReadValue(name)
}
