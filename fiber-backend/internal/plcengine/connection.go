package plcengine

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ADSClient defines the low-level interface for ADS communication
type ADSClient interface {
	ReadSymbol(name string) (interface{}, error)
	WriteSymbol(name string, value interface{}) error
	ReadSymbols(names []string) (map[string]interface{}, error)
	Close() error
}

// ConnectionState represents the current state of a PLC connection
type ConnectionState int

const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateConnected
	StateError
)

var (
	ErrNotConnected = errors.New("PLC connection not established")
)

// PLCConnection handles a single physical connection to a TwinCAT PLC.
// It multiplexes requests from multiple goroutines to respect the ADS
// "one connection per IP" limitation.
type PLCConnection struct {
	MachineID string
	IP        string
	Port      int
	AmsNetID  string

	client ADSClient
	state  ConnectionState
	mu     sync.RWMutex

	// Internal queues for multiplexing
	requestChan chan *internalRequest
	stopChan    chan struct{}

	stats ConnectionStatus
}

type internalRequest struct {
	op       string // "read", "write", "batch_read"
	symbol   string
	symbols  []string
	value    interface{}
	respChan chan *internalResponse
}

type internalResponse struct {
	value  interface{}
	values map[string]interface{}
	err    error
}

func NewPLCConnection(machineID, ip string, amsID string, port int) *PLCConnection {
	return &PLCConnection{
		MachineID:   machineID,
		IP:          ip,
		Port:        port,
		AmsNetID:    amsID,
		requestChan: make(chan *internalRequest, 100),
		stopChan:    make(chan struct{}),
		stats: ConnectionStatus{
			MachineID: machineID,
		},
	}
}

func (c *PLCConnection) Start(ctx context.Context, clientFactory func() (ADSClient, error)) {
	go c.handler(clientFactory)
}

func (c *PLCConnection) Stop() {
	close(c.stopChan)
}

func (c *PLCConnection) handler(clientFactory func() (ADSClient, error)) {
	// Initial connection attempt
	c.checkConnection(clientFactory)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			if c.client != nil {
				c.client.Close()
			}
			return

		case req := <-c.requestChan:
			c.processRequest(req)

		case <-ticker.C:
			c.checkConnection(clientFactory)
		}
	}
}

func (c *PLCConnection) checkConnection(factory func() (ADSClient, error)) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state == StateConnected {
		return
	}

	c.state = StateConnecting
	client, err := factory()
	if err != nil {
		c.state = StateError
		c.stats.ErrorCount++
		c.stats.Connected = false
		return
	}

	c.client = client
	c.state = StateConnected
	c.stats.Connected = true
	c.stats.ReconnectCount++
	c.stats.LastSeen = time.Now()
}

func (c *PLCConnection) processRequest(req *internalRequest) {
	c.mu.RLock()
	client := c.client
	state := c.state
	c.mu.RUnlock()

	if state != StateConnected || client == nil {
		req.respChan <- &internalResponse{err: ErrNotConnected}
		return
	}

	var resp internalResponse
	switch req.op {
	case "read":
		resp.value, resp.err = client.ReadSymbol(req.symbol)
	case "write":
		resp.err = client.WriteSymbol(req.symbol, req.value)
	case "batch_read":
		resp.values, resp.err = client.ReadSymbols(req.symbols)
	}

	if resp.err == nil {
		c.mu.Lock()
		c.stats.LastSeen = time.Now()
		c.mu.Unlock()
	}

	req.respChan <- &resp
}

// Public API for the connection (Thread-safe via channel)

func (c *PLCConnection) ReadSymbol(symbol string) (interface{}, error) {
	respChan := make(chan *internalResponse, 1)
	c.requestChan <- &internalRequest{op: "read", symbol: symbol, respChan: respChan}

	resp := <-respChan
	return resp.value, resp.err
}

func (c *PLCConnection) WriteSymbol(symbol string, value interface{}) error {
	respChan := make(chan *internalResponse, 1)
	c.requestChan <- &internalRequest{op: "write", symbol: symbol, value: value, respChan: respChan}

	resp := <-respChan
	return resp.err
}

func (c *PLCConnection) ReadSymbols(symbols []string) (map[string]interface{}, error) {
	respChan := make(chan *internalResponse, 1)
	c.requestChan <- &internalRequest{op: "batch_read", symbols: symbols, respChan: respChan}

	resp := <-respChan
	return resp.values, resp.err
}
