package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"fiber-backend/internal/auth"
	"fiber-backend/internal/modules/apikey"

	"github.com/gofiber/fiber/v3"
)

// APIKey validates the X-API-Key header.
func APIKey(repo apikey.Repository) fiber.Handler {
	return func(c fiber.Ctx) error {
		key := c.Get("X-API-Key")
		if key == "" {
			return c.Next() // Fallback to JWT if allowed
		}

		// Parse key: prefix_secret
		parts := strings.Split(key, "_")
		if len(parts) != 2 {
			return c.Status(401).JSON(fiber.Map{"error": "invalid api key format"})
		}

		// Hash the provided key
		h := sha256.Sum256([]byte(key))
		hash := hex.EncodeToString(h[:])

		k, err := repo.GetByHash(context.Background(), hash)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "invalid or expired api key"})
		}

		// Update last used (fire and forget)
		go func() {
			_ = repo.UpdateLastUsed(context.Background(), k.ID)
		}()

		// Map API Key to synthetic Claims for downstream handlers
		claims := &auth.Claims{
			UserID:      k.UserID,
			Username:    "api_key_" + k.Prefix,
			Roles:       []string{"machine"},
			Permissions: map[string][]string{"*": k.Scopes},
		}

		c.Locals("user", claims)
		c.Locals("user_id", claims.UserID)
		c.Locals("is_api_key", true)

		return c.Next()
	}
}
