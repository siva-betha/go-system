package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters — tuned for security + reasonable performance.
const (
	argonTime    = 1
	argonMemory  = 64 * 1024 // 64 MB
	argonThreads = 4
	argonKeyLen  = 32
	saltLen      = 16
)

// HashPassword hashes a plain-text password using Argon2id.
// Returns a self-describing string: $argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>
func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generating salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

// CheckPassword compares a plain-text password against an Argon2id hash.
func CheckPassword(encoded, password string) bool {
	parts := strings.Split(encoded, "$")
	// format: $argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>
	if len(parts) != 6 {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}

	existing, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}

	candidate := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	return subtle.ConstantTimeCompare(existing, candidate) == 1
}
