package influx

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type Handler struct {
	Client     influxdb2.Client
	Org        string
	Bucket     string
	AuthMethod string
	AuthMasked string
	URL        string
}

// QueryRange queries data points from InfluxDB for a given measurement and time range
// @Summary Query InfluxDB range
// @Description Query points from InfluxDB for a specific measurement with optional tag filters (last 1h)
// @Tags influx
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param measurement query string true "Measurement name"
// @Param range query string false "Range (e.g., -1h, -5m)" default("-1h")
// @Param limit query int false "Limit number of points" default(100)
// @Param chamber_id query string false "Chamber ID"
// @Param layer_id query string false "Layer ID"
// @Param wafer_id query string false "Wafer ID"
// @Param system query string false "System"
// @Success 200 {array} Point
// @Failure 500 {object} string
// @Router /influx/range [get]
func (h Handler) QueryRange(c fiber.Ctx) error {
	measurement := c.Query("measurement")
	chamberID := c.Query("chamber_id")
	layerID := c.Query("layer_id")
	waferID := c.Query("wafer_id")
	system := c.Query("system")
	rangeParam := c.Query("range", "-1h")
	limitStr := c.Query("limit", "100")

	// Strip quotes if they are passed in the URL (e.g. range="-1m")
	rangeClean := strings.Trim(rangeParam, "\"")

	query := `from(bucket: "` + h.Bucket + `")
|> range(start: ` + rangeClean + `)
|> filter(fn: (r) => r._measurement == "` + measurement + `")`

	if chamberID != "" {
		query += ` |> filter(fn: (r) => r.chamber_id == "` + chamberID + `")`
	}
	if layerID != "" {
		query += ` |> filter(fn: (r) => r.layer_id == "` + layerID + `")`
	}
	if waferID != "" {
		query += ` |> filter(fn: (r) => r.wafer_id == "` + waferID + `")`
	}
	if system != "" {
		query += ` |> filter(fn: (r) => r.system == "` + system + `")`
	}

	query += ` |> limit(n: ` + limitStr + `)`

	// Log query for debugging
	log.Printf("Executing Flux Query for bucket [%s] org [%s]:\n%s", h.Bucket, h.Org, query)

	api := h.Client.QueryAPI(h.Org)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := api.Query(ctx, query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":       err.Error(),
			"query":       query,
			"bucket_dbg":  "[" + h.Bucket + "]",
			"org_dbg":     "[" + h.Org + "]",
			"url_dbg":     "[" + h.URL + "]",
			"auth_method": h.AuthMethod,
			"auth_masked": h.AuthMasked,
			"auth_info":   "If auth_masked shows [empty] or unexpected length, check .env. Hidden chars are stripped.",
		})
	}

	out := []Point{}

	for result.Next() {
		rec := result.Record()
		p := Point{
			Time:        rec.Time().String(),
			Value:       rec.Value(),
			Field:       rec.Field(),
			Measurement: rec.Measurement(),
		}

		// Helper to safely get tag values
		if v, ok := rec.ValueByKey("chamber_id").(string); ok {
			p.ChamberID = v
		}
		if v, ok := rec.ValueByKey("destination").(string); ok {
			p.Destination = v
		}
		if v, ok := rec.ValueByKey("layer_id").(string); ok {
			p.LayerID = v
		}
		if v, ok := rec.ValueByKey("source").(string); ok {
			p.Source = v
		}
		if v, ok := rec.ValueByKey("system").(string); ok {
			p.System = v
		}
		if v, ok := rec.ValueByKey("telegraf_instance_id").(string); ok {
			p.TelegrafInstanceID = v
		}
		if v, ok := rec.ValueByKey("wafer_id").(string); ok {
			p.WaferID = v
		}

		out = append(out, p)
	}

	return c.JSON(out)
}

// Health checks InfluxDB connectivity
// @Summary InfluxDB Health
// @Description Check if InfluxDB is reachable
// @Tags influx
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} interface{}
// @Failure 500 {object} string
// @Router /influx/health [get]
func (h Handler) Health(c fiber.Ctx) error {
	ok, err := h.Client.Health(context.Background())
	if err != nil {
		return c.Status(500).JSON(err.Error())
	}

	return c.JSON(ok)
}
