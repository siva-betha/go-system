package collector

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/robinson/gouads"
)

type SymbolManager struct {
	conn            *gouads.Connection
	symbols         map[string]*SymbolInfo
	mappings        map[string]string // source -> target
	reverseMappings map[string]string // target -> source
	mu              sync.RWMutex
}

type SymbolInfo struct {
	Name         string
	TwinCATType  string
	GoType       reflect.Kind
	Size         int
	IsArray      bool
	ArraySize    int
	IsIdentifier bool
	StorageType  string // "influx", "postgres", "both", "array"
	TargetName   string // mapped target name
}

type DataCollectionConfig struct {
	ScalarFields  []FieldConfig
	FieldMappings map[string]string
}

type FieldConfig struct {
	Name    string
	StoreIn string
}

type ReadGroup struct {
	Name     string
	Symbols  []string
	Interval time.Duration
}

func NewSymbolManager(conn *gouads.Connection) *SymbolManager {
	return &SymbolManager{
		conn:            conn,
		symbols:         make(map[string]*SymbolInfo),
		mappings:        make(map[string]string),
		reverseMappings: make(map[string]string),
	}
}

func (sm *SymbolManager) LoadSymbols(config *DataCollectionConfig) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Discover all symbols from PLC
	plcSymbols, err := sm.conn.ReadSymbolsInfo()
	if err != nil {
		return fmt.Errorf("failed to read symbols: %w", err)
	}

	// Apply field mappings
	for source, target := range config.FieldMappings {
		sm.mappings[source] = target
		sm.reverseMappings[target] = source
	}

	// Build symbol info
	for _, sym := range plcSymbols {
		info := &SymbolInfo{
			Name:        sym.Name,
			TwinCATType: sym.Type,
			Size:        sym.Size,
			IsArray:     sym.IsArray,
		}

		// Determine Go type
		info.GoType = sm.mapTwinCATType(sym.Type)

		// Check if this is our OES array
		if sym.Name == "MAIN_OES.FlatArray" {
			info.ArraySize = 2048
			info.StorageType = "array"
		}

		// Check mappings
		if target, ok := sm.mappings[sym.Name]; ok {
			info.TargetName = target
		}

		// Check config for storage preferences
		for _, field := range config.ScalarFields {
			if field.Name == sym.Name {
				info.StorageType = field.StoreIn
				break
			}
		}

		sm.symbols[sym.Name] = info
	}

	return nil
}

func (sm *SymbolManager) mapTwinCATType(tcType string) reflect.Kind {
	switch tcType {
	case "BOOL":
		return reflect.Bool
	case "SINT", "INT", "DINT":
		return reflect.Int32
	case "USINT", "UINT", "UDINT":
		return reflect.Uint32
	case "REAL":
		return reflect.Float32
	case "LREAL":
		return reflect.Float64
	case "STRING":
		return reflect.String
	default:
		return reflect.Interface
	}
}

func (sm *SymbolManager) GetReadGroups() []*ReadGroup {
	// Group symbols for efficient reading
	// This is a simplified implementation for the demonstration
	groups := []*ReadGroup{
		{
			Name:     "scalars_10ms",
			Symbols:  sm.getSymbolsByInterval(10 * time.Millisecond),
			Interval: 10 * time.Millisecond,
		},
		{
			Name:     "recipe_info",
			Symbols:  sm.getRecipeSymbols(),
			Interval: 1 * time.Second,
		},
	}
	return groups
}

func (sm *SymbolManager) getSymbolsByInterval(interval time.Duration) []string {
	// Mock implementation
	return []string{"MAIN_OES.Continuos_Aquis"}
}

func (sm *SymbolManager) getRecipeSymbols() []string {
	// Mock implementation
	return []string{"Recipe.recipe_exe.Process_Job", "Recipe.recipe_exe.Substrate_ID"}
}
