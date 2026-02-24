package plcengine

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Engine is the high-level interface for the PLC Read/Write system
type Engine interface {
	Start(configs []MachineConfig) error
	Stop() error

	ReadSymbol(machineID, symbol string) (*PLCValue, error)
	ReadSymbols(machineID string, symbols []string) (map[string]*PLCValue, error)

	WriteSymbol(machineID, symbol string, value interface{}) error
	WriteAsync(req WriteRequest) <-chan WriteResponse

	GetStatus() map[string]ConnectionStatus
}

// MachineConfig is a subset of the configuration needed for connection
type MachineConfig struct {
	ID       string
	IP       string
	AmsNetID string
	Port     int
}

type PLCReadWriteEngine struct {
	connections map[string]*PLCConnection
	writer      *PrioritizedWriter
	mu          sync.RWMutex

	dataChan     chan PLCValue
	writeConfirm chan WriteResponse
	stopChan     chan struct{}
	wg           sync.WaitGroup

	// Dependency injection for client creation
	ClientFactory func(ip, amsID string, port int) (ADSClient, error)
}

func NewEngine(dataChan chan PLCValue) *PLCReadWriteEngine {
	e := &PLCReadWriteEngine{
		connections:  make(map[string]*PLCConnection),
		dataChan:     dataChan,
		writeConfirm: make(chan WriteResponse, 100),
		stopChan:     make(chan struct{}),
	}
	e.writer = NewPrioritizedWriter(e)
	return e
}

func (e *PLCReadWriteEngine) Start(configs []MachineConfig) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.writer.Start()

	for _, cfg := range configs {
		conn := NewPLCConnection(cfg.ID, cfg.IP, cfg.AmsNetID, cfg.Port)
		e.connections[cfg.ID] = conn

		conn.Start(context.Background(), func() (ADSClient, error) {
			if e.ClientFactory != nil {
				return e.ClientFactory(cfg.IP, cfg.AmsNetID, cfg.Port)
			}
			return NewMockADSClient(cfg.ID), nil // Default to mock for dev
		})
	}

	return nil
}

func (e *PLCReadWriteEngine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.writer.Stop()

	for _, conn := range e.connections {
		conn.Stop()
	}

	return nil
}

func (e *PLCReadWriteEngine) getConnection(machineID string) (*PLCConnection, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	conn, ok := e.connections[machineID]
	if !ok {
		return nil, fmt.Errorf("machine %s not found", machineID)
	}
	return conn, nil
}

func (e *PLCReadWriteEngine) ReadSymbol(machineID, symbol string) (*PLCValue, error) {
	conn, err := e.getConnection(machineID)
	if err != nil {
		return nil, err
	}

	val, err := conn.ReadSymbol(symbol)
	if err != nil {
		return nil, err
	}

	return &PLCValue{
		Symbol:    symbol,
		Value:     val,
		Timestamp: time.Now(),
		Source:    machineID,
	}, nil
}

func (e *PLCReadWriteEngine) ReadSymbols(machineID string, symbols []string) (map[string]*PLCValue, error) {
	conn, err := e.getConnection(machineID)
	if err != nil {
		return nil, err
	}

	rawValues, err := conn.ReadSymbols(symbols)
	if err != nil {
		return nil, err
	}

	results := make(map[string]*PLCValue)
	now := time.Now()
	for sym, val := range rawValues {
		results[sym] = &PLCValue{
			Symbol:    sym,
			Value:     val,
			Timestamp: now,
			Source:    machineID,
		}
	}
	return results, nil
}

func (e *PLCReadWriteEngine) WriteSymbol(machineID, symbol string, value interface{}) error {
	conn, err := e.getConnection(machineID)
	if err != nil {
		return err
	}

	return conn.WriteSymbol(symbol, value)
}

func (e *PLCReadWriteEngine) GetStatus() map[string]ConnectionStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()

	status := make(map[string]ConnectionStatus)
	for id, conn := range e.connections {
		conn.mu.RLock()
		status[id] = conn.stats
		conn.mu.RUnlock()
	}
	return status
}

func (e *PLCReadWriteEngine) WriteAsync(req WriteRequest) <-chan WriteResponse {
	respChan := make(chan WriteResponse, 1)
	req.ResponseChan = respChan

	if err := e.writer.Submit(req); err != nil {
		respChan <- WriteResponse{
			ID:      req.ID,
			Success: false,
			Error:   err.Error(),
		}
	}

	return respChan
}

// Mock implementation for development

type MockADSClient struct {
	MachineID string
}

func NewMockADSClient(id string) *MockADSClient {
	return &MockADSClient{MachineID: id}
}

func (m *MockADSClient) ReadSymbol(name string) (interface{}, error) {
	return 42.0, nil // Simulated value
}

func (m *MockADSClient) WriteSymbol(name string, value interface{}) error {
	return nil
}

func (m *MockADSClient) ReadSymbols(names []string) (map[string]interface{}, error) {
	res := make(map[string]interface{})
	for _, n := range names {
		res[n] = 42.0
	}
	return res, nil
}

func (m *MockADSClient) Close() error { return nil }
