package exporter

import (
	"time"
)

type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type ExportRequest struct {
	ID           string              `json:"id"`
	TimeRange    TimeRange           `json:"time_range"`
	Machines     []string            `json:"machines"`
	Chambers     []string            `json:"chambers"`
	Symbols      []string            `json:"symbols"`
	Format       string              `json:"format"` // "binary", "csv", "json"
	Compression  bool                `json:"compression"`
	SplitSize    int64               `json:"split_size"`
	Destination  string              `json:"destination"`
	ResponseChan chan<- ExportResult `json:"-"`
}

type ExportProgress struct {
	RequestID       string        `json:"request_id"`
	TotalPoints     int64         `json:"total_points"`
	ExportedPoints  int64         `json:"exported_points"`
	CurrentFile     string        `json:"current_file"`
	PercentComplete float64       `json:"percent_complete"`
	Speed           float64       `json:"speed"` // points/sec
	EstimatedTime   time.Duration `json:"estimated_time"`
	Error           error         `json:"error,omitempty"`
}

type ExportResult struct {
	RequestID      string        `json:"request_id"`
	Success        bool          `json:"success"`
	Files          []string      `json:"files"`
	TotalPoints    int64         `json:"total_points"`
	TotalSize      int64         `json:"total_size"`
	CompressedSize int64         `json:"compressed_size"`
	Ratio          float64       `json:"ratio"`
	Duration       time.Duration `json:"duration"`
	Error          string        `json:"error,omitempty"`
}

type ImportRequest struct {
	ID         string `json:"id"`
	SourceFile string `json:"source_file"`
	Mode       string `json:"mode"` // "append", "replace", "merge"
}

type ImportProgress struct {
	RequestID       string  `json:"request_id"`
	TotalBlocks     int     `json:"total_blocks"`
	ImportedBlocks  int     `json:"imported_blocks"`
	TotalPoints     int64   `json:"total_points"`
	ImportedPoints  int64   `json:"imported_points"`
	PercentComplete float64 `json:"percent_complete"`
}

type ImportResult struct {
	RequestID   string        `json:"request_id"`
	Success     bool          `json:"success"`
	TotalPoints int           `json:"total_points"`
	Duration    time.Duration `json:"duration"`
	Error       string        `json:"error,omitempty"`
}

type Point struct {
	Timestamp time.Time   `json:"t"`
	Machine   string      `json:"m"`
	Chamber   string      `json:"c"`
	Symbol    string      `json:"s"`
	Value     interface{} `json:"v"`
}

type BlockIndex struct {
	Offset     int64     `json:"offset"`
	Length     int64     `json:"length"`
	PointCount int       `json:"point_count"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
}

type Metadata struct {
	Version     string    `json:"version"`
	ExportTime  time.Time `json:"export_time"`
	TimeRange   TimeRange `json:"time_range"`
	TotalPoints int64     `json:"total_points"`
}
