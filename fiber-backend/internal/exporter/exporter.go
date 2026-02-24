package exporter

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type ExportSystem struct {
	// Export channels
	exportRequests chan ExportRequest
	exportProgress chan ExportProgress
	exportResults  chan ExportResult

	// Import channels
	importRequests chan ImportRequest
	importProgress chan ImportProgress
	importResults  chan ImportResult

	// Control
	stopChan chan struct{}
	wg       sync.WaitGroup

	// Dependencies
	influxClient influxdb2.Client
	org          string
	bucket       string

	// File management
	exportDir string
}

func NewExportSystem(client influxdb2.Client, org, bucket, exportDir string) *ExportSystem {
	return &ExportSystem{
		exportRequests: make(chan ExportRequest, 100),
		exportProgress: make(chan ExportProgress, 100),
		exportResults:  make(chan ExportResult, 100),
		importRequests: make(chan ImportRequest, 50),
		importProgress: make(chan ImportProgress, 100),
		importResults:  make(chan ImportResult, 50),
		stopChan:       make(chan struct{}),
		influxClient:   client,
		org:            org,
		bucket:         bucket,
		exportDir:      exportDir,
	}
}

func (e *ExportSystem) Start() {
	// Start export worker pool (2 workers)
	for i := 0; i < 2; i++ {
		e.wg.Add(1)
		go e.exportWorker(i)
	}

	// Start import worker pool (1 worker)
	for i := 0; i < 1; i++ {
		e.wg.Add(1)
		go e.importWorker(i)
	}

	log.Println("Export/Import system started")
}

func (e *ExportSystem) Stop() {
	close(e.stopChan)
	e.wg.Wait()
	log.Println("Export/Import system stopped")
}

func (e *ExportSystem) exportWorker(id int) {
	defer e.wg.Done()
	log.Printf("Export worker %d started", id)

	for {
		select {
		case req := <-e.exportRequests:
			e.executeExport(req)
		case <-e.stopChan:
			return
		}
	}
}

func (e *ExportSystem) importWorker(id int) {
	defer e.wg.Done()
	log.Printf("Import worker %d started", id)

	for {
		select {
		case req := <-e.importRequests:
			e.executeImport(req)
		case <-e.stopChan:
			return
		}
	}
}

func (e *ExportSystem) SubmitExport(req ExportRequest) {
	select {
	case e.exportRequests <- req:
	default:
		log.Println("Export request queue full")
	}
}

func (e *ExportSystem) SubmitImport(req ImportRequest) {
	select {
	case e.importRequests <- req:
	default:
		log.Println("Import request queue full")
	}
}

func (e *ExportSystem) executeExport(req ExportRequest) {
	startTime := time.Now()
	filename := e.generateFilename(req)

	file, err := os.Create(filename)
	if err != nil {
		e.sendResult(req, ExportResult{Success: false, Error: err.Error()})
		return
	}
	defer file.Close()

	writer, err := NewCompressedWriter(file, req.Compression)
	if err != nil {
		e.sendResult(req, ExportResult{Success: false, Error: err.Error()})
		return
	}

	// Query InfluxDB
	points, err := e.queryInfluxDB(req)
	if err != nil {
		e.sendResult(req, ExportResult{Success: false, Error: err.Error()})
		return
	}

	if err := writer.WriteBatch(points); err != nil {
		e.sendResult(req, ExportResult{Success: false, Error: err.Error()})
		return
	}

	if err := writer.Close(); err != nil {
		e.sendResult(req, ExportResult{Success: false, Error: err.Error()})
		return
	}

	result := ExportResult{
		RequestID:      req.ID,
		Success:        true,
		Files:          []string{filename},
		TotalPoints:    int64(len(points)),
		CompressedSize: writer.compressedSize,
		Duration:       time.Since(startTime),
	}
	e.sendResult(req, result)
}

func (e *ExportSystem) generateFilename(req ExportRequest) string {
	ts := time.Now().Format("20060102_150405")
	return fmt.Sprintf("%s/export_%s_%s.plc", e.exportDir, req.ID, ts)
}

func (e *ExportSystem) queryInfluxDB(req ExportRequest) ([]Point, error) {
	queryAPI := e.influxClient.QueryAPI(e.org)

	flux := fmt.Sprintf(`from(bucket: "%s")
		|> range(start: %s, stop: %s)
		|> filter(fn: (r) => r["_measurement"] == "plc_data")`,
		e.bucket,
		req.TimeRange.Start.Format(time.RFC3339),
		req.TimeRange.End.Format(time.RFC3339),
	)

	// Add filters if specified
	if len(req.Machines) > 0 {
		flux += ` |> filter(fn: (r) => `
		for i, m := range req.Machines {
			if i > 0 {
				flux += " or "
			}
			flux += fmt.Sprintf(`r["machine_id"] == "%s"`, m)
		}
		flux += ")"
	}

	result, err := queryAPI.Query(context.Background(), flux)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	var points []Point
	for result.Next() {
		val := result.Record().Value()
		points = append(points, Point{
			Timestamp: result.Record().Time(),
			Machine:   result.Record().ValueByKey("machine_id").(string),
			Chamber:   result.Record().ValueByKey("chamber_id").(string),
			Symbol:    result.Record().Field(),
			Value:     val,
		})
	}

	return points, result.Err()
}

func (e *ExportSystem) sendResult(req ExportRequest, res ExportResult) {
	if req.ResponseChan != nil {
		req.ResponseChan <- res
	}
	e.exportResults <- res
}

func (e *ExportSystem) executeImport(req ImportRequest) {
	startTime := time.Now()
	file, err := os.Open(req.SourceFile)
	if err != nil {
		e.sendImportResult(req, ImportResult{Success: false, Error: err.Error()})
		return
	}
	defer file.Close()

	reader, err := NewCompressedReader(file)
	if err != nil {
		e.sendImportResult(req, ImportResult{Success: false, Error: err.Error()})
		return
	}

	totalPoints := 0
	for i := 0; i < len(reader.index); i++ {
		points, err := reader.ReadBlock(i)
		if err != nil {
			e.sendImportResult(req, ImportResult{Success: false, Error: err.Error()})
			return
		}

		if err := e.writeToInfluxDB(points); err != nil {
			e.sendImportResult(req, ImportResult{Success: false, Error: err.Error()})
			return
		}
		totalPoints += len(points)
	}

	result := ImportResult{
		RequestID:   req.ID,
		Success:     true,
		TotalPoints: totalPoints,
		Duration:    time.Since(startTime),
	}
	e.sendImportResult(req, result)
}

func (e *ExportSystem) writeToInfluxDB(points []Point) error {
	writeAPI := e.influxClient.WriteAPI(e.org, e.bucket)

	for _, p := range points {
		tags := map[string]string{
			"machine_id": p.Machine,
			"chamber_id": p.Chamber,
		}
		fields := map[string]interface{}{
			p.Symbol: p.Value,
		}

		pt := influxdb2.NewPoint("plc_data", tags, fields, p.Timestamp)
		writeAPI.WritePoint(pt)
	}

	writeAPI.Flush()
	return nil
}

func (e *ExportSystem) sendImportResult(req ImportRequest, res ImportResult) {
	e.importResults <- res
}
