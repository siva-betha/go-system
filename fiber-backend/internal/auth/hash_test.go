package auth

import (
	"testing"
)

func TestHashPassword_ProducesValidHash(t *testing.T) {
	hash, err := HashPassword("MyStr0ng!Pass")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if hash == "" {
		t.Fatal("HashPassword returned empty string")
	}

	// Must start with the argon2id identifier
	if hash[:9] != "$argon2id" {
		t.Errorf("hash does not start with $argon2id: %s", hash[:20])
	}
}

func TestHashPassword_UniqueSalts(t *testing.T) {
	h1, _ := HashPassword("samepassword")
	h2, _ := HashPassword("samepassword")

	if h1 == h2 {
		t.Error("two hashes of the same password should differ (unique salts)")
	}
}

func TestCheckPassword_Correct(t *testing.T) {
	password := "Correct!Horse42"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	if !CheckPassword(hash, password) {
		t.Error("CheckPassword should return true for the correct password")
	}
}

func TestCheckPassword_Wrong(t *testing.T) {
	hash, _ := HashPassword("RealPassword1!")

	if CheckPassword(hash, "WrongPassword1!") {
		t.Error("CheckPassword should return false for wrong password")
	}
}

func TestCheckPassword_MalformedHash(t *testing.T) {
	if CheckPassword("notahash", "anything") {
		t.Error("CheckPassword should return false for a malformed hash")
	}
}
