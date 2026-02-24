package apikey

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	Repo Repository
}

func (h Handler) Create(c fiber.Ctx) error {
	var req struct {
		Name      string     `json:"name"`
		Scopes    []string   `json:"scopes"`
		ExpiresAt *time.Time `json:"expires_at"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	userID, _ := c.Locals("user_id").(string)

	rawKey, prefix, hash, err := generateSecureKey()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate key"})
	}

	k := &APIKey{
		UserID:    userID,
		Name:      req.Name,
		KeyHash:   hash,
		Prefix:    prefix,
		Scopes:    req.Scopes,
		ExpiresAt: req.ExpiresAt,
	}

	if err := h.Repo.Create(context.Background(), k); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(CreateAPIKeyResponse{
		ID:     k.ID, // Note: ID will be set by DB, but repo implementation for Mock might set it
		RawKey: fmt.Sprintf("%s_%s", prefix, rawKey),
	})
}

func (h Handler) List(c fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)

	keys, err := h.Repo.ListByUser(context.Background(), userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(keys)
}

func (h Handler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.Repo.Delete(context.Background(), id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

func generateSecureKey() (rawKey, prefix, hash string, err error) {
	// Generate prefix (6 chars)
	p := make([]byte, 3)
	if _, err := rand.Read(p); err != nil {
		return "", "", "", err
	}
	prefix = hex.EncodeToString(p)

	// Generate raw secret (32 bytes)
	s := make([]byte, 32)
	if _, err := rand.Read(s); err != nil {
		return "", "", "", err
	}
	rawKey = hex.EncodeToString(s)

	// Hash the combined string
	fullKey := fmt.Sprintf("%s_%s", prefix, rawKey)
	h := sha256.Sum256([]byte(fullKey))
	hash = hex.EncodeToString(h[:])

	return rawKey, prefix, hash, nil
}
