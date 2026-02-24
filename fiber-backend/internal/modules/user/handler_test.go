package user

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"fiber-backend/internal/auth"
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
		"name":     "Alice",
		"email":    "alice@test.com",
		"password": "Strong!Pass1",
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
	if userMap["email"] != "alice@test.com" {
		t.Errorf("expected email alice@test.com, got %v", userMap["email"])
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	app := setupApp(NewMockRepo(), NewMockTokenRepo())

	doRequest(app, "POST", "/api/auth/register", map[string]string{
		"name": "Alice", "email": "dup@test.com", "password": "Strong!Pass1",
	})

	resp, body := doRequest(app, "POST", "/api/auth/register", map[string]string{
		"name": "Alice2", "email": "dup@test.com", "password": "Strong!Pass2",
	})

	if resp.StatusCode != 409 {
		t.Fatalf("expected 409, got %d — body: %v", resp.StatusCode, body)
	}
}

func TestRegister_ValidationError(t *testing.T) {
	app := setupApp(NewMockRepo(), NewMockTokenRepo())

	resp, _ := doRequest(app, "POST", "/api/auth/register", map[string]string{
		"name":     "A",
		"email":    "not-an-email",
		"password": "short",
	})

	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

// ---------- Login tests ----------

func TestLogin_Success(t *testing.T) {
	repo := NewMockRepo()
	tokenRepo := NewMockTokenRepo()
	app := setupApp(repo, tokenRepo)

	doRequest(app, "POST", "/api/auth/register", map[string]string{
		"name": "Bob", "email": "bob@test.com", "password": "BobPass123!",
	})

	resp, body := doRequest(app, "POST", "/api/auth/login", map[string]string{
		"email": "bob@test.com", "password": "BobPass123!",
	})

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d — body: %v", resp.StatusCode, body)
	}

	if body["access_token"] == nil {
		t.Error("response should contain access_token")
	}
	if body["refresh_token"] == nil {
		t.Error("response should contain refresh_token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := NewMockRepo()
	tokenRepo := NewMockTokenRepo()
	app := setupApp(repo, tokenRepo)

	doRequest(app, "POST", "/api/auth/register", map[string]string{
		"name": "Charlie", "email": "charlie@test.com", "password": "Right!Pass1",
	})

	resp, _ := doRequest(app, "POST", "/api/auth/login", map[string]string{
		"email": "charlie@test.com", "password": "Wrong!Pass9",
	})

	if resp.StatusCode != 401 {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestLogin_NonExistentEmail(t *testing.T) {
	app := setupApp(NewMockRepo(), NewMockTokenRepo())

	resp, _ := doRequest(app, "POST", "/api/auth/login", map[string]string{
		"email": "noone@test.com", "password": "Whatever1!",
	})

	if resp.StatusCode != 401 {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

// ---------- Profile tests ----------

func TestProfile_WithValidToken(t *testing.T) {
	app := setupApp(NewMockRepo(), NewMockTokenRepo())

	_, regBody := doRequest(app, "POST", "/api/auth/register", map[string]string{
		"name": "Dana", "email": "dana@test.com", "password": "DanaPass12!",
	})

	token := regBody["access_token"].(string)

	resp, body := doRequest(app, "GET", "/api/auth/me", nil, token)

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d — body: %v", resp.StatusCode, body)
	}

	if body["email"] != "dana@test.com" {
		t.Errorf("expected email dana@test.com, got %v", body["email"])
	}
}

func TestProfile_WithoutToken(t *testing.T) {
	app := setupApp(NewMockRepo(), NewMockTokenRepo())

	resp, _ := doRequest(app, "GET", "/api/auth/me", nil)

	if resp.StatusCode != 401 {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestProfile_WithInvalidToken(t *testing.T) {
	app := setupApp(NewMockRepo(), NewMockTokenRepo())

	resp, _ := doRequest(app, "GET", "/api/auth/me", nil, "invalid.token.here")

	if resp.StatusCode != 401 {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

// ---------- Refresh token tests ----------

func TestRefresh_Success(t *testing.T) {
	app := setupApp(NewMockRepo(), NewMockTokenRepo())

	// Register to get tokens
	_, regBody := doRequest(app, "POST", "/api/auth/register", map[string]string{
		"name": "Eve", "email": "eve@test.com", "password": "EvePass123!",
	})

	refreshToken := regBody["refresh_token"].(string)

	// Use refresh token to get new pair
	resp, body := doRequest(app, "POST", "/api/auth/refresh", map[string]string{
		"refresh_token": refreshToken,
	})

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d — body: %v", resp.StatusCode, body)
	}

	if body["access_token"] == nil {
		t.Error("response should contain new access_token")
	}
	if body["refresh_token"] == nil {
		t.Error("response should contain new refresh_token")
	}

	// New refresh token should be different (rotation)
	newRefresh := body["refresh_token"].(string)
	if newRefresh == refreshToken {
		t.Error("new refresh token should differ from old one (rotation)")
	}
}

func TestRefresh_OldTokenInvalidAfterRotation(t *testing.T) {
	app := setupApp(NewMockRepo(), NewMockTokenRepo())

	_, regBody := doRequest(app, "POST", "/api/auth/register", map[string]string{
		"name": "Frank", "email": "frank@test.com", "password": "FrankPass1!",
	})

	oldRefresh := regBody["refresh_token"].(string)

	// First refresh — works
	doRequest(app, "POST", "/api/auth/refresh", map[string]string{
		"refresh_token": oldRefresh,
	})

	// Second refresh with same token — should fail (token was rotated)
	resp, _ := doRequest(app, "POST", "/api/auth/refresh", map[string]string{
		"refresh_token": oldRefresh,
	})

	if resp.StatusCode != 401 {
		t.Fatalf("expected 401 for reused refresh token, got %d", resp.StatusCode)
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	app := setupApp(NewMockRepo(), NewMockTokenRepo())

	resp, _ := doRequest(app, "POST", "/api/auth/refresh", map[string]string{
		"refresh_token": "completely-invalid-token",
	})

	if resp.StatusCode != 401 {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

// ---------- Logout tests ----------

func TestLogout_InvalidatesRefreshToken(t *testing.T) {
	app := setupApp(NewMockRepo(), NewMockTokenRepo())

	_, regBody := doRequest(app, "POST", "/api/auth/register", map[string]string{
		"name": "Grace", "email": "grace@test.com", "password": "GracePass1!",
	})

	refreshToken := regBody["refresh_token"].(string)

	// Logout
	resp, _ := doRequest(app, "POST", "/api/auth/logout", map[string]string{
		"refresh_token": refreshToken,
	})

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// Try to use the invalidated refresh token
	resp2, _ := doRequest(app, "POST", "/api/auth/refresh", map[string]string{
		"refresh_token": refreshToken,
	})

	if resp2.StatusCode != 401 {
		t.Fatalf("expected 401 after logout, got %d", resp2.StatusCode)
	}
}

// ---------- Protected user CRUD tests ----------

func TestListUsers_RequiresAuth(t *testing.T) {
	app := setupApp(NewMockRepo(), NewMockTokenRepo())

	resp, _ := doRequest(app, "GET", "/api/users/", nil)

	if resp.StatusCode != 401 {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestListUsers_WithAuth(t *testing.T) {
	app := setupApp(NewMockRepo(), NewMockTokenRepo())

	_, regBody := doRequest(app, "POST", "/api/auth/register", map[string]string{
		"name": "Hank", "email": "hank@test.com", "password": "HankPass123!",
	})

	token := regBody["access_token"].(string)

	resp, _ := doRequest(app, "GET", "/api/users/", nil, token)

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

// generateTestToken creates a JWT for test use.
func generateTestToken() string {
	token, _ := auth.Generate(1, "test@test.com", "user")
	return token
}
