package streamer

import (
	"sync"
)

// RingBuffer stores a fixed number of recent data points
type RingBuffer struct {
	data  []BroadcastMsg
	size  int
	head  int
	count int
	mu    sync.RWMutex
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		data: make([]BroadcastMsg, size),
		size: size,
	}
}

func (r *RingBuffer) Add(msg BroadcastMsg) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.data[r.head] = msg
	r.head = (r.head + 1) % r.size
	if r.count < r.size {
		r.count++
	}
}

func (r *RingBuffer) GetRecent(machineID, chamberID string) []BroadcastMsg {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]BroadcastMsg, 0, r.count)

	// Start from oldest available
	start := (r.head - r.count + r.size) % r.size

	for i := 0; i < r.count; i++ {
		idx := (start + i) % r.size
		msg := r.data[idx]

		matchMachine := machineID == "" || msg.MachineID == machineID
		matchChamber := chamberID == "" || msg.ChamberID == chamberID

		if matchMachine && matchChamber {
			result = append(result, msg)
		}
	}

	return result
}
