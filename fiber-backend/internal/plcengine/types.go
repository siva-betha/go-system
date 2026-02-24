package plcengine

import (
	"time"
)

// PLCType defines supported PLC data types
type PLCType int

const (
	TypeBool PLCType = iota
	TypeInt8
	TypeInt16
	TypeInt32
	TypeReal
	TypeString
)

// SymbolInfo contains metadata about a PLC symbol
type SymbolInfo struct {
	Name       string
	Type       PLCType
	Size       int
	IsWritable bool
	MinValue   float64
	MaxValue   float64
	Unit       string
	Comment    string
}

// PLCValue represents a data point read from a PLC
type PLCValue struct {
	Symbol    string      `json:"symbol"`
	Value     interface{} `json:"value"`
	Type      PLCType     `json:"type"`
	Quality   int         `json:"quality"` // 0-100
	Timestamp time.Time   `json:"timestamp"`
	Source    string      `json:"source"` // Machine ID
}

// WriteRequest defines a request to change a PLC field
type WriteRequest struct {
	ID           string             `json:"id"`
	MachineID    string             `json:"machine_id"`
	Symbol       string             `json:"symbol"`
	Value        interface{}        `json:"value"`
	Priority     int                `json:"priority"` // 0=low, 10=high
	RequireAck   bool               `json:"require_ack"`
	Timeout      time.Duration      `json:"timeout"`
	ResponseChan chan WriteResponse `json:"-"`
}

// WriteResponse confirms the result of a write operation
type WriteResponse struct {
	ID        string    `json:"id"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// ConnectionStatus represents the health of a PLC connection
type ConnectionStatus struct {
	MachineID      string    `json:"machine_id"`
	Connected      bool      `json:"connected"`
	LastSeen       time.Time `json:"last_seen"`
	ErrorCount     int       `json:"error_count"`
	ReconnectCount int       `json:"reconnect_count"`
}
