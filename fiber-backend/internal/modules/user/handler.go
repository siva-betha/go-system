package user

import (
	"context"
	"errors"
	"time"

	"fiber-backend/internal/auth"
	"fiber-backend/internal/modules/audit"
	"fiber-backend/internal/validator"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

// Handler holds repository interfaces for user and token data access.
type Handler struct {
	Repo         Repository
	TokenRepo    TokenRepository
	AuditService *audit.Service
}

// ---------- helper: issue token pair ----------

func (h Handler) issueTokenPair(ctx context.Context, u *User) (*AuthResponse, error) {
	roles, perms, err := h.Repo.GetRolesAndPermissions(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	accessToken, expiresIn, err := auth.GenerateAccessToken(u.ID, u.Username, roles, perms)
	if err != nil {
		return nil, err
	}

	rawRefresh, hashRefresh, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(auth.RefreshTokenExpiry())
	if err := h.TokenRepo.StoreRefreshToken(ctx, u.ID, hashRefresh, expiresAt); err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    expiresIn,
		User:         u.ToResponse(),
	}, nil
}

// ---------- Auth handlers ----------

func (h Handler) Register(c fiber.Ctx) error {
	var req RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := validator.V.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	existing, _ := h.Repo.GetByUsername(ctx, req.Username)
	if existing != nil {
		return c.Status(409).JSON(fiber.Map{"error": "username already registered"})
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to hash password"})
	}

	u := User{
		Username:     req.Username,
		Email:        req.Email,
		FullName:     req.FullName,
		PasswordHash: hash,
	}

	if err := h.Repo.Create(ctx, &u); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create user"})
	}

	resp, err := h.issueTokenPair(ctx, &u)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate tokens"})
	}

	return c.Status(201).JSON(resp)
}

func (h Handler) Login(c fiber.Ctx) error {
	var req LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := validator.V.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	u, err := h.Repo.GetByUsername(ctx, req.Username)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid username or password"})
	}

	if !auth.CheckPassword(u.PasswordHash, req.Password) {
		return c.Status(401).JSON(fiber.Map{"error": "invalid username or password"})
	}

	resp, err := h.issueTokenPair(ctx, u)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate tokens"})
	}

	return c.JSON(resp)
}

func (h Handler) Refresh(c fiber.Ctx) error {
	var req RefreshRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := validator.V.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tokenHash := auth.HashToken(req.RefreshToken)
	userID, expiresAt, err := h.TokenRepo.FindRefreshToken(ctx, tokenHash)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid refresh token"})
	}

	if time.Now().After(expiresAt) {
		_ = h.TokenRepo.DeleteRefreshToken(ctx, tokenHash)
		return c.Status(401).JSON(fiber.Map{"error": "refresh token expired"})
	}

	_ = h.TokenRepo.DeleteRefreshToken(ctx, tokenHash)

	u, err := h.Repo.GetByID(ctx, userID)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "user not found"})
	}

	resp, err := h.issueTokenPair(ctx, u)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate tokens"})
	}

	return c.JSON(resp)
}

func (h Handler) Logout(c fiber.Ctx) error {
	var req RefreshRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tokenHash := auth.HashToken(req.RefreshToken)
	_ = h.TokenRepo.DeleteRefreshToken(ctx, tokenHash)

	return c.JSON(fiber.Map{"message": "logged out successfully"})
}

func (h Handler) Profile(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return fiber.ErrUnauthorized
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	u, err := h.Repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.Status(404).JSON(fiber.Map{"error": "user not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "internal error"})
	}

	return c.JSON(u.ToResponse())
}

// ---------- Admin CRUD handlers ----------

func (h Handler) Create(c fiber.Ctx) error {
	var u User
	if err := c.Bind().Body(&u); err != nil {
		return fiber.ErrBadRequest
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := h.Repo.Create(ctx, &u); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	h.AuditService.Log(c, "CREATE", "user", &u.ID, nil, u)

	return c.JSON(u)
}

func (h Handler) List(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	users, err := h.Repo.List(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(users)
}

func (h Handler) Update(c fiber.Ctx) error {
	id := c.Params("id")

	var u User
	if err := c.Bind().Body(&u); err != nil {
		return fiber.ErrBadRequest
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	old, _ := h.Repo.GetByID(ctx, id)

	err := h.Repo.Update(ctx, id, &u)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}

	h.AuditService.Log(c, "UPDATE", "user", &id, old, u)

	return c.JSON(fiber.Map{"updated": true})
}

func (h Handler) Delete(c fiber.Ctx) error {
	id := c.Params("id")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	old, _ := h.Repo.GetByID(ctx, id)

	err := h.Repo.Delete(ctx, id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}

	h.AuditService.Log(c, "DELETE", "user", &id, old, nil)

	return c.JSON(fiber.Map{"deleted": true})
}
