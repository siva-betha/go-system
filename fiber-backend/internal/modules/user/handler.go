package user

import (
	"context"
	"errors"
	"strconv"
	"time"

	"fiber-backend/internal/auth"
	"fiber-backend/internal/validator"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

// Handler holds repository interfaces for user and token data access.
type Handler struct {
	Repo      Repository
	TokenRepo TokenRepository
}

// ---------- helper: issue token pair ----------

func (h Handler) issueTokenPair(ctx context.Context, u *User) (*AuthResponse, error) {
	accessToken, expiresIn, err := auth.GenerateAccessToken(u.ID, u.Email, u.Role)
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

// Register godoc
// @Summary     Register a new user
// @Description Creates a new user account with Argon2id hashed password and returns access + refresh tokens
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body     RegisterRequest true "Registration payload"
// @Success     201  {object} AuthResponse
// @Failure     400  {object} map[string]interface{} "Validation error or bad request"
// @Failure     409  {object} map[string]interface{} "Email already registered"
// @Failure     500  {object} map[string]interface{} "Internal server error"
// @Router      /auth/register [post]
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

	existing, _ := h.Repo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return c.Status(409).JSON(fiber.Map{"error": "email already registered"})
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to hash password"})
	}

	u := User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hash,
		Role:         "user",
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

// Login godoc
// @Summary     Log in with email and password
// @Description Authenticates a user and returns access + refresh tokens
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body     LoginRequest true "Login credentials"
// @Success     200  {object} AuthResponse
// @Failure     400  {object} map[string]interface{} "Validation error"
// @Failure     401  {object} map[string]interface{} "Invalid email or password"
// @Failure     500  {object} map[string]interface{} "Internal server error"
// @Router      /auth/login [post]
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

	u, err := h.Repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid email or password"})
	}

	if !auth.CheckPassword(u.PasswordHash, req.Password) {
		return c.Status(401).JSON(fiber.Map{"error": "invalid email or password"})
	}

	resp, err := h.issueTokenPair(ctx, u)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate tokens"})
	}

	return c.JSON(resp)
}

// Refresh godoc
// @Summary     Refresh access token
// @Description Exchange a valid refresh token for a new access + refresh token pair (token rotation)
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body     RefreshRequest true "Refresh token"
// @Success     200  {object} AuthResponse
// @Failure     400  {object} map[string]interface{} "Validation error"
// @Failure     401  {object} map[string]interface{} "Invalid or expired refresh token"
// @Failure     500  {object} map[string]interface{} "Internal server error"
// @Router      /auth/refresh [post]
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

	// Look up the hashed refresh token in DB
	tokenHash := auth.HashToken(req.RefreshToken)
	userID, expiresAt, err := h.TokenRepo.FindRefreshToken(ctx, tokenHash)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid refresh token"})
	}

	// Check expiry
	if time.Now().After(expiresAt) {
		// Clean up expired token
		_ = h.TokenRepo.DeleteRefreshToken(ctx, tokenHash)
		return c.Status(401).JSON(fiber.Map{"error": "refresh token expired"})
	}

	// Delete the old refresh token (rotation: each token is single-use)
	_ = h.TokenRepo.DeleteRefreshToken(ctx, tokenHash)

	// Fetch user to build response
	u, err := h.Repo.GetByID(ctx, userID)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "user not found"})
	}

	// Issue a new pair
	resp, err := h.issueTokenPair(ctx, u)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate tokens"})
	}

	return c.JSON(resp)
}

// Logout godoc
// @Summary     Log out (invalidate refresh token)
// @Description Invalidates the provided refresh token, preventing further use
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body     RefreshRequest true "Refresh token to invalidate"
// @Success     200  {object} map[string]interface{} "Logged out successfully"
// @Failure     400  {object} map[string]interface{} "Validation error"
// @Router      /auth/logout [post]
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

// Profile godoc
// @Summary     Get current user profile
// @Description Returns the authenticated user's profile data (requires JWT)
// @Tags        auth
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} UserResponse
// @Failure     401 {object} map[string]interface{} "Unauthorized"
// @Failure     404 {object} map[string]interface{} "User not found"
// @Router      /auth/me [get]
func (h Handler) Profile(c fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
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

// Create godoc
// @Summary     Create a new user (admin)
// @Description Admin endpoint to create a user with specified fields
// @Tags        users
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       user body     User true "User object"
// @Success     200  {object} User
// @Failure     400  {object} map[string]interface{} "Validation error"
// @Failure     500  {object} map[string]interface{} "Internal server error"
// @Router      /users [post]
func (h Handler) Create(c fiber.Ctx) error {
	var u User
	if err := c.Bind().Body(&u); err != nil {
		return fiber.ErrBadRequest
	}

	if err := validator.V.Struct(u); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := h.Repo.Create(ctx, &u); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(u)
}

// List godoc
// @Summary     List all users
// @Description Returns a list of all users in the system
// @Tags        users
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array}  UserResponse
// @Failure     500 {object} map[string]interface{} "Internal server error"
// @Router      /users [get]
func (h Handler) List(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	users, err := h.Repo.List(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(users)
}

// Update godoc
// @Summary     Update a user
// @Description Update an existing user by ID
// @Tags        users
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id   path     int  true "User ID"
// @Param       user body     User true "User object"
// @Success     200  {object} map[string]interface{}
// @Failure     400  {object} map[string]interface{} "Bad request"
// @Failure     404  {object} map[string]interface{} "User not found"
// @Router      /users/{id} [put]
func (h Handler) Update(c fiber.Ctx) error {
	idStr := c.Params("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return fiber.ErrBadRequest
	}

	var u User
	if err := c.Bind().Body(&u); err != nil {
		return fiber.ErrBadRequest
	}

	if err := validator.V.Struct(u); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = h.Repo.Update(ctx, id, &u)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}

	return c.JSON(fiber.Map{"updated": true})
}

// Delete godoc
// @Summary     Delete a user
// @Description Delete an existing user by ID
// @Tags        users
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id  path     int true "User ID"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{} "Bad request"
// @Failure     404 {object} map[string]interface{} "User not found"
// @Router      /users/{id} [delete]
func (h Handler) Delete(c fiber.Ctx) error {
	idStr := c.Params("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return fiber.ErrBadRequest
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = h.Repo.Delete(ctx, id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}

	return c.JSON(fiber.Map{"deleted": true})
}
