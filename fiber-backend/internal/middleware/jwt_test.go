package middleware

import (
	"net/http/httptest"
	"os"
	"testing"

	"fiber-backend/internal/auth"

	"github.com/gofiber/fiber/v3"
)

func TestMain(m *testing.M) {
	os.Setenv("JWT_SECRET", "middleware-test-secret")
	os.Setenv("JWT_ACCESS_EXP_MIN", "5")
	os.Exit(m.Run())
}

func TestJWT_MissingHeader(t *testing.T) {
	app := fiber.New()
	app.Get("/protected", JWT(), func(c fiber.Ctx) error {
		return c.SendStatus(200)
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != 401 {
		t.Errorf("expected 401 for missing header, got %d", resp.StatusCode)
	}
}

func TestJWT_InvalidToken(t *testing.T) {
	app := fiber.New()
	app.Get("/protected", JWT(), func(c fiber.Ctx) error {
		return c.SendStatus(200)
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	resp, _ := app.Test(req)

	if resp.StatusCode != 401 {
		t.Errorf("expected 401 for invalid token, got %d", resp.StatusCode)
	}
}

func TestJWT_ValidToken(t *testing.T) {
	app := fiber.New()

	// Middleware under test
	app.Use(JWT())

	app.Get("/protected", func(c fiber.Ctx) error {
		// Assert claims are set in Locals
		uid, ok := c.Locals("userID").(int64)
		if !ok || uid != 123 {
			return c.Status(500).JSON(fiber.Map{"error": "userID missing or wrong"})
		}

		role, ok := c.Locals("role").(string)
		if !ok || role != "admin" {
			return c.Status(500).JSON(fiber.Map{"error": "role missing or wrong"})
		}

		return c.SendStatus(200)
	})

	// Generate valid token
	token, _ := auth.Generate(123, "test@test.com", "admin")

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, _ := app.Test(req)

	if resp.StatusCode != 200 {
		t.Errorf("expected 200 for valid token, got %d", resp.StatusCode)
	}
}
