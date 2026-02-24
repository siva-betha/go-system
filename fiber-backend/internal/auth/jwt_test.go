package auth

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestMain(m *testing.M) {
	// Set test-only JWT secret
	os.Setenv("JWT_SECRET", "test-secret-key-for-unit-tests")
	os.Setenv("JWT_ACCESS_EXP_MIN", "5")
	os.Exit(m.Run())
}

func TestGenerate_ReturnsToken(t *testing.T) {
	token, err := Generate(1, "test@test.com", "user")
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if token == "" {
		t.Fatal("Generate returned empty token")
	}
}

func TestParseToken_ExtractsClaims(t *testing.T) {
	token, err := Generate(42, "alice@example.com", "admin")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	claims, err := ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}

	if claims.UserID != 42 {
		t.Errorf("expected UserID 42, got %d", claims.UserID)
	}
	if claims.Email != "alice@example.com" {
		t.Errorf("expected email alice@example.com, got %s", claims.Email)
	}
	if claims.Role != "admin" {
		t.Errorf("expected role admin, got %s", claims.Role)
	}
}

func TestParseToken_RejectsExpiredToken(t *testing.T) {
	// Create a token that expired 1 hour ago
	claims := Claims{
		UserID: 1,
		Email:  "expired@test.com",
		Role:   "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	_, err := ParseToken(signed)
	if err == nil {
		t.Error("ParseToken should reject an expired token")
	}
}

func TestParseToken_RejectsInvalidSignature(t *testing.T) {
	// Sign with a different secret
	claims := Claims{
		UserID: 1,
		Email:  "bad@test.com",
		Role:   "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte("wrong-secret"))

	_, err := ParseToken(signed)
	if err == nil {
		t.Error("ParseToken should reject a token signed with wrong secret")
	}
}
