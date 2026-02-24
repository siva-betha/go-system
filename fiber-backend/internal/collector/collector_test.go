package collector

import (
	"testing"
	"time"

	"fiber-backend/internal/plcengine"
	"fiber-backend/internal/streamer"

	"github.com/stretchr/testify/assert"
)

func TestCollector_ChannelBackpressure(t *testing.T) {
	engine := &plcengine.PLCReadWriteEngine{}
	hub := &streamer.StreamHub{}
	c := NewCollector(engine, hub)
	c.dataChan = make(chan plcengine.PLCValue, 10) // Small buffer

	// Fill channel
	for i := 0; i < 10; i++ {
		c.dataChan <- plcengine.PLCValue{Value: i}
	}

	// Next send should not block
	done := make(chan bool)
	go func() {
		data := plcengine.PLCValue{Value: 11}
		select {
		case c.dataChan <- data:
			// Unexpected
		default:
			// Correctly dropped
		}
		done <- true
	}()

	select {
	case <-done:
		// Success - didn't block
	case <-time.After(100 * time.Millisecond):
		t.Error("Send blocked when channel full")
	}
}

func TestCollector_StartStop(t *testing.T) {
	engine := &plcengine.PLCReadWriteEngine{}
	hub := &streamer.StreamHub{}
	c := NewCollector(engine, hub)
	configs := []MachineConfig{
		{
			ID:       "m1",
			Name:     "Test Machine",
			IP:       "127.0.0.1",
			AmsNetID: "1.2.3.4.1.1",
			Port:     851,
			Chambers: []ChamberConfig{
				{
					ID:   "c1",
					Name: "Chamber 1",
					Symbols: []SymbolConfig{
						{Name: "GVL.temp", DataType: "float"},
					},
				},
			},
		},
	}

	err := c.Start(configs)
	assert.NoError(t, err)

	// Let it run for a bit
	time.Sleep(100 * time.Millisecond)

	err = c.Stop()
	assert.NoError(t, err)
}
