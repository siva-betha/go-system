package auth

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestMain(m *testing.M) {
	os.Setenv("JWT_SECRET", "test-secret-key-for-unit-tests")
	os.Setenv("JWT_ACCESS_EXP_MIN", "5")
	os.Exit(m.Run())
}

func TestGenerateAndParse(t *testing.T) {
	userID := "550e8400-e29b-41d4-a716-446655440000"
	username := "testuser"
	roles := []string{"admin"}
	perms := map[string][]string{"machine": {"read"}}

	token, _, err := GenerateAccessToken(userID, username, roles, perms)
	if err != nil {
		t.Fatalf("failed to generate: %v", err)
	}

	claims, err := ParseToken(token)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected userID %s, got %s", userID, claims.UserID)
	}
	if claims.Username != username {
		t.Errorf("expected username %s, got %s", username, claims.Username)
	}
	if len(claims.Roles) != len(roles) || claims.Roles[0] != roles[0] {
		t.Errorf("expected roles %v, got %v", roles, claims.Roles)
	}
}

func TestParseToken_RejectsExpiredToken(t *testing.T) {
	claims := Claims{
		UserID:   "expired-user",
		Username: "expired",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
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
	claims := Claims{
		UserID:   "bad-sig-user",
		Username: "badsig",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte("wrong-secret"))

	_, err := ParseToken(signed)
	if err == nil {
		t.Error("ParseToken should reject a token signed with wrong secret")
	}
}
