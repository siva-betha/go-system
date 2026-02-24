package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"fiber-backend/internal/config"
	"fiber-backend/internal/database"
	"fiber-backend/internal/middleware"
	"fiber-backend/internal/modules/influx"
	"fiber-backend/internal/modules/user"

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
	cfg := config.Load()

	db, err := database.Connect(cfg.DBUrl)
	if err != nil {
		log.Fatal(err)
	}

	if err := database.RunMigrations(cfg.DBUrl, "up", 0); err != nil {
		log.Println("migrations:", err)
	}

	database.Seed(db)

	// Aggressive trimming is REQUIRED because .env files on Windows often have hidden \r
	getInfluxEnv := func(keys ...string) string {
		for _, key := range keys {
			if v := os.Getenv(key); v != "" {
				return strings.TrimSpace(v)
			}
		}
		return ""
	}

	influxURL := getInfluxEnv("INFLUX_URL", "INFLUXDB_URL")
	influxToken := getInfluxEnv("INFLUX_TOKEN", "INFLUXDB_TOKEN")
	influxOrg := getInfluxEnv("INFLUX_ORG", "INFLUXDB_ORG")
	influxBucket := getInfluxEnv("INFLUX_BUCKET", "INFLUXDB_BUCKET")
	influxUser := getInfluxEnv("INFLUX_USERNAME", "INFLUXDB_USERNAME")
	influxPass := getInfluxEnv("INFLUX_PASSWORD", "INFLUXDB_PASSWORD")

	influxClient, authMethod, authMasked := database.ConnectInflux(
		influxURL,
		influxToken,
		influxUser,
		influxPass,
	)
	log.Printf("InfluxDB Init: URL=[%s], Org=[%s], Bucket=[%s], AuthSource=%s", influxURL, influxOrg, influxBucket, authMethod)

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

	// ✅ PUBLIC auth routes (register, login, refresh, logout, profile)
	user.AuthRoutes(app.Group("/api/auth"), userRepo, tokenRepo)

	// ✅ JWT protected group
	api := app.Group("/api", middleware.JWT())

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
	user.Routes(api.Group("/users"), userRepo, tokenRepo)

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
	go app.Listen(":" + cfg.AppPort)

	// graceful shutdown
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = app.ShutdownWithContext(ctx)
	db.Close()
}
