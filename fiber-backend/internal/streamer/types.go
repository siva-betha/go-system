package streamer

import (
	"time"
)

// Message types for WebSocket communication
type MessageType string

const (
	MsgTypeData      MessageType = "data"
	MsgTypeSubscribe MessageType = "subscribe"
	MsgTypeUnsub     MessageType = "unsubscribe"
	MsgTypeHistory   MessageType = "history"
	MsgTypeError     MessageType = "error"
)

// BroadcastMsg is the JSON packet sent to browser clients
type BroadcastMsg struct {
	Type      MessageType            `json:"type"`
	MachineID string                 `json:"machine_id,omitempty"`
	ChamberID string                 `json:"chamber_id,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Error     string                 `json:"error,omitempty"`
}

// ClientMessage is the JSON packet received from browser clients
type ClientMessage struct {
	Type      MessageType `json:"type"`
	MachineID string      `json:"machine_id,omitempty"`
	ChamberID string      `json:"chamber_id,omitempty"`
	Symbols   []string    `json:"symbols,omitempty"`
	Duration  string      `json:"duration,omitempty"` // For history requests
}

// Subscription tracks what a client is interested in
type Subscription struct {
	MachineID string
	ChamberID string
	Symbols   []string
}

// StreamStats tracks performance of the streamer
type StreamStats struct {
	ActiveClients int    `json:"active_clients"`
	MessagesSent  uint64 `json:"messages_sent"`
	Uptime        string `json:"uptime"`
}
