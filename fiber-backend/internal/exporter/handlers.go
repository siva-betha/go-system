package exporter

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func (e *ExportSystem) RegisterRoutes(router fiber.Router) {
	group := router.Group("/export")
	group.Post("/start", e.handleExportStart)
	group.Get("/list", e.handleExportList)
	group.Get("/download/:id", e.handleExportDownload)

	importGroup := router.Group("/import")
	importGroup.Post("/start", e.handleImportStart)
}

func (e *ExportSystem) handleExportStart(c fiber.Ctx) error {
	var body struct {
		Start       string   `json:"start"`
		End         string   `json:"end"`
		Machines    []string `json:"machines"`
		Compression bool     `json:"compression"`
	}

	if err := c.Bind().JSON(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	startTime, _ := time.Parse(time.RFC3339, body.Start)
	endTime, _ := time.Parse(time.RFC3339, body.End)

	req := ExportRequest{
		ID: uuid.New().String(),
		TimeRange: TimeRange{
			Start: startTime,
			End:   endTime,
		},
		Machines:    body.Machines,
		Format:      "binary",
		Compression: body.Compression,
	}

	e.SubmitExport(req)

	return c.JSON(fiber.Map{
		"request_id": req.ID,
		"status":     "queued",
	})
}

func (e *ExportSystem) handleExportList(c fiber.Ctx) error {
	files, err := os.ReadDir(e.exportDir)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	var results []map[string]interface{}
	for _, f := range files {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".plc" {
			info, _ := f.Info()
			results = append(results, map[string]interface{}{
				"name": f.Name(),
				"size": info.Size(),
				"time": info.ModTime(),
			})
		}
	}

	return c.JSON(results)
}

func (e *ExportSystem) handleExportDownload(c fiber.Ctx) error {
	id := c.Params("id")
	// For simplicity, search for file in exportDir starting with export_ID
	files, _ := filepath.Glob(filepath.Join(e.exportDir, fmt.Sprintf("export_%s_*.plc", id)))
	if len(files) == 0 {
		return c.Status(http.StatusNotFound).SendString("File not found")
	}

	return c.Download(files[0])
}

func (e *ExportSystem) handleImportStart(c fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Save file temporarily
	tempPath := filepath.Join(os.TempDir(), file.Filename)
	if err := c.SaveFile(file, tempPath); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	req := ImportRequest{
		ID:         uuid.New().String(),
		SourceFile: tempPath,
		Mode:       "append",
	}

	e.SubmitImport(req)

	return c.JSON(fiber.Map{
		"request_id": req.ID,
		"status":     "queued",
	})
}
