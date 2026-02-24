package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"fiber-backend/internal/alerter"
	"fiber-backend/internal/collector"
	"fiber-backend/internal/config"
	"fiber-backend/internal/database"
	"fiber-backend/internal/exporter"
	"fiber-backend/internal/middleware"
	"fiber-backend/internal/modules/apikey"
	"fiber-backend/internal/modules/approval"
	"fiber-backend/internal/modules/audit"
	"fiber-backend/internal/modules/influx"
	"fiber-backend/internal/modules/machine_config"
	"fiber-backend/internal/modules/user"
	"fiber-backend/internal/plcengine"
	"fiber-backend/internal/streamer"

	_ "fiber-backend/docs"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title          Fiber Backend API
// @version        1.0
// @description    Go Fiber v3 backend with JWT authentication, Argon2 password hashing, and PostgreSQL.
// @host           localhost:3000
// @BasePath       /api
// @securityDefinitions.apikey BearerAuth
// @in   header
// @name Authorization
// @description Enter your JWT token. The `Bearer ` prefix is optional.
func main() {
	getEnv := func(key, fallback string) string {
		if v := os.Getenv(key); v != "" {
			return strings.TrimSpace(v)
		}
		return fallback
	}

	cfg := config.Load()
	dbURL := cfg.DBUrl

	db, err := database.Connect(dbURL)
	if err != nil {
		log.Fatal(err)
	}

	if err := database.RunMigrations(dbURL, "up", 0); err != nil {
		log.Println("migrations:", err)
	}

	database.Seed(db)

	influxURL := getEnv("INFLUX_URL", "http://localhost:8086")
	influxToken := getEnv("INFLUX_TOKEN", "")
	influxOrg := getEnv("INFLUX_ORG", "plc-org")
	influxBucket := getEnv("INFLUX_BUCKET", "plc-data")
	influxUser := getEnv("INFLUX_USERNAME", "admin")
	influxPass := getEnv("INFLUX_PASSWORD", "StrongPassword123!")

	influxClient, authMethod, authMasked := database.ConnectInflux(
		influxURL,
		influxToken,
		influxUser,
		influxPass,
	)
	log.Printf("InfluxDB Init: URL=[%s], Org=[%s], Bucket=[%s], AuthSource=%s", influxURL, influxOrg, influxBucket, authMethod)

	// --- Industrial System Initialization ---
	hub := streamer.NewHub()
	go hub.Run()

	// Data channel for cross-component values (PLC -> Engine -> Collector -> Kafka/UI)
	dataChan := make(chan plcengine.PLCValue, 10000)
	engine := plcengine.NewEngine(dataChan)

	col := collector.NewCollector(engine, hub)

	// --- Storage Monitoring Initialization ---
	storageConfig := alerter.AlerterConfig{
		CheckInterval:    5 * time.Minute,
		WarningPercent:   80,
		CriticalPercent:  90,
		EmergencyPercent: 95,
		Email: alerter.EmailConfig{
			SMTPHost: getEnv("SMTP_HOST", "smtp.gmail.com"),
			SMTPPort: 587,
			Username: getEnv("SMTP_USER", ""),
			Password: getEnv("SMTP_PASS", ""),
			From:     getEnv("SMTP_FROM", "alerts@example.com"),
			To:       strings.Split(getEnv("SMTP_TO", "admin@example.com"), ","),
		},
		Paths: map[string]string{
			"system":     getEnv("MONITOR_PATH_SYSTEM", "/"),
			"influxdb":   getEnv("MONITOR_PATH_INFLUX", "/var/lib/influxdb2"),
			"postgresql": getEnv("MONITOR_PATH_POSTGRES", "/var/lib/postgresql/data"),
			"logs":       getEnv("MONITOR_PATH_LOGS", "./logs"),
		},
		AutoCleanup: true,
	}
	storageMon := alerter.NewStorageMonitor(storageConfig, db)
	storageMon.Start()

	// --- Data Export/Import System Initialization ---
	exportDir := getEnv("EXPORT_DIR", "./exports")
	os.MkdirAll(exportDir, 0755)
	exportSystem := exporter.NewExportSystem(influxClient, influxOrg, influxBucket, exportDir)
	exportSystem.Start()

	// Fetch initial configs from DB to bootstrap the collector
	machineRepo := machine_config.PgRepo{DB: db}
	dbMachines, err := machineRepo.GetMachines(context.Background())
	if err == nil {
		var collectorConfigs []collector.MachineConfig
		for _, m := range dbMachines {
			// Map DB model to Collector config
			c := collector.MachineConfig{
				ID:       m.ID,
				Name:     m.Name,
				IP:       m.IP,
				AmsNetID: m.AmsNetID,
				Port:     m.Port,
			}
			for _, ch := range m.Chambers {
				cc := collector.ChamberConfig{
					ID:   ch.ID,
					Name: ch.Name,
				}
				for _, s := range ch.Symbols {
					cc.Symbols = append(cc.Symbols, collector.SymbolConfig{
						Name:     s.Name,
						DataType: s.DataType,
					})
				}
				c.Chambers = append(c.Chambers, cc)
			}
			collectorConfigs = append(collectorConfigs, c)
		}

		if len(collectorConfigs) > 0 {
			if err := col.Start(collectorConfigs); err != nil {
				log.Printf("Failed to start collector: %v", err)
			} else {
				log.Printf("Collector started with %d machines", len(collectorConfigs))
			}
		}
	}

	// ✅ create app ONLY ONCE — with error handler
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			return c.Status(code).JSON(fiber.Map{
				"request_id": c.Locals("reqid"),
				"error":      err.Error(),
			})
		},
	})

	// ✅ global middleware order
	app.Use(middleware.RequestID())
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(middleware.CORS())
	middleware.Security(app)

	// ✅ swagger initialization
	app.Get("/swagger/*", adaptor.HTTPHandler(httpSwagger.WrapHandler))

	// ✅ health endpoints
	// @Summary Health check
	// @Description Get the current health status of the application
	// @Tags system
	// @Produce json
	// @Success 200 {object} map[string]string
	// @Router /health [get]
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// ✅ WebSocket endpoint for real-time data
	app.Get("/ws", hub.NewHandler())

	// @Summary Database check
	// @Description Check database connectivity
	// @Tags system
	// @Produce plain
	// @Success 200 {string} string "DB ok"
	// @Failure 503 {string} string "Service Unavailable"
	// @Router /dbcheck [get]
	app.Get("/dbcheck", func(c fiber.Ctx) error {
		if err := db.Ping(context.Background()); err != nil {
			return fiber.ErrServiceUnavailable
		}
		return c.SendString("DB ok")
	})

	userRepo := user.PgRepo{DB: db}
	tokenRepo := user.PgTokenRepo{DB: db}

	// RBAC Phase 2: Audit and Approval
	auditRepo := audit.PgRepo{DB: db}
	auditSvc := audit.NewService(auditRepo)

	approvalRepo := approval.PgRepo{DB: db}
	approvalSvc := approval.NewService(approvalRepo)
	approvalHandler := approval.Handler{Service: approvalSvc}

	apiKeyRepo := apikey.PgRepo{DB: db}
	apiKeyHandler := apikey.Handler{Repo: apiKeyRepo}

	// ✅ PUBLIC auth routes (register, login, refresh, logout, profile)
	user.AuthRoutes(app.Group("/api/auth"), userRepo, tokenRepo, auditSvc)

	// ✅ machine_config routes (public)
	machine_config.Register(app.Group("/api"), machineRepo, approvalSvc, auditSvc)

	// ✅ JWT protected group
	api := app.Group("/api", middleware.JWT())

	// ✅ approval routes (protected)
	approval.Routes(api.Group("/approvals"), &approvalHandler)

	// ✅ audit routes (protected)
	audit.Routes(api.Group("/audit"), auditRepo)

	// ✅ api_key routes (protected)
	apikey.Routes(api.Group("/keys"), &apiKeyHandler)

	// @Summary Get current user (test)
	// @Description Returns simple auth status and request ID
	// @Tags user
	// @Security BearerAuth
	// @Produce json
	// @Success 200 {object} map[string]interface{}
	// @Failure 401 {object} map[string]interface{}
	// @Router /me [get]
	api.Get("/me", func(c fiber.Ctx) error {
		// This snippet was provided in the wrong context, assuming it was meant for an error handler within an InfluxDB route.
		// Reverting to original /me handler as per instruction to keep existing comments/empty lines and not make unrelated edits.
		return c.JSON(fiber.Map{
			"user":       "authorized",
			"request_id": c.Locals("reqid"),
		})
	})

	// ✅ admin user management routes (protected)
	user.Routes(api.Group("/users"), userRepo, tokenRepo, auditSvc)

	// ✅ storage monitoring routes (protected)
	storageMon.RegisterRoutes(api)

	// ✅ data export/import routes (protected)
	exportSystem.RegisterRoutes(api)

	// ✅ influx routes (protected)
	influx.Routes(
		api.Group("/influx"),
		influx.Handler{
			Client:     influxClient,
			Org:        influxOrg,
			Bucket:     influxBucket,
			AuthMethod: authMethod,
			AuthMasked: authMasked,
			URL:        influxURL,
		},
	)

	// ✅ start server
	port := getEnv("APP_PORT", cfg.AppPort)
	go app.Listen(":" + port)

	// graceful shutdown
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = app.ShutdownWithContext(ctx)
	exportSystem.Stop()
	storageMon.Stop()
	col.Stop()
	engine.Stop()
	db.Close()
}
