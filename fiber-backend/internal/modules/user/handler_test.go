package user

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"fiber-backend/internal/middleware"

	"github.com/gofiber/fiber/v3"
)

// ---------- helpers ----------

func TestMain(m *testing.M) {
	os.Setenv("JWT_SECRET", "handler-test-secret")
	os.Setenv("JWT_ACCESS_EXP_MIN", "5")
	os.Setenv("JWT_REFRESH_EXP_DAYS", "1")
	os.Exit(m.Run())
}

func setupApp(repo Repository, tokenRepo TokenRepository) *fiber.App {
	app := fiber.New()
	AuthRoutes(app.Group("/api/auth"), repo, tokenRepo)
	Routes(app.Group("/api/users", middleware.JWT()), repo, tokenRepo)
	return app
}

func doRequest(app *fiber.App, method, url string, body interface{}, token ...string) (*http.Response, map[string]interface{}) {
	var reqBody io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewReader(b)
	}

	req, _ := http.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if len(token) > 0 && token[0] != "" {
		req.Header.Set("Authorization", "Bearer "+token[0])
	}

	resp, _ := app.Test(req)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()

	return resp, result
}

// ---------- Register tests ----------

func TestRegister_Success(t *testing.T) {
	app := setupApp(NewMockRepo(), NewMockTokenRepo())

	resp, body := doRequest(app, "POST", "/api/auth/register", map[string]string{
		"username":  "Alice",
		"email":     "alice@test.com",
		"password":  "Strong!Pass1",
		"full_name": "Alice Smith",
	})

	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d — body: %v", resp.StatusCode, body)
	}

	if body["access_token"] == nil || body["access_token"] == "" {
		t.Error("response should contain access_token")
	}
	if body["refresh_token"] == nil || body["refresh_token"] == "" {
		t.Error("response should contain refresh_token")
	}

	userMap, ok := body["user"].(map[string]interface{})
	if !ok {
		t.Fatal("response should contain user object")
	}
	if userMap["username"] != "Alice" {
		t.Errorf("expected username Alice, got %v", userMap["username"])
	}
}

// ---------- Login tests ----------

func TestLogin_Success(t *testing.T) {
	repo := NewMockRepo()
	tokenRepo := NewMockTokenRepo()
	app := setupApp(repo, tokenRepo)

	// Pre-create user via repo or register
	doRequest(app, "POST", "/api/auth/register", map[string]string{
		"username": "Bob", "email": "bob@test.com", "password": "BobPass123!",
	})

	resp, body := doRequest(app, "POST", "/api/auth/login", map[string]string{
		"username": "Bob", "password": "BobPass123!",
	})

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d — body: %v", resp.StatusCode, body)
	}

	if body["access_token"] == nil {
		t.Error("response should contain access_token")
	}
}

// ---------- Profile tests ----------

func TestProfile_WithValidToken(t *testing.T) {
	app := setupApp(NewMockRepo(), NewMockTokenRepo())

	_, regBody := doRequest(app, "POST", "/api/auth/register", map[string]string{
		"username": "Dana", "email": "dana@test.com", "password": "DanaPass12!",
	})

	token := regBody["access_token"].(string)

	resp, body := doRequest(app, "GET", "/api/auth/me", nil, token)

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d — body: %v", resp.StatusCode, body)
	}

	if body["username"] != "Dana" {
		t.Errorf("expected username Dana, got %v", body["username"])
	}
}

func TestListUsers_WithAuth(t *testing.T) {
	app := setupApp(NewMockRepo(), NewMockTokenRepo())

	_, regBody := doRequest(app, "POST", "/api/auth/register", map[string]string{
		"username": "Hank", "email": "hank@test.com", "password": "HankPass123!",
	})

	token := regBody["access_token"].(string)

	resp, _ := doRequest(app, "GET", "/api/users/", nil, token)

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}
