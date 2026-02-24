package user

import (
	"fiber-backend/internal/middleware"

	"github.com/gofiber/fiber/v3"
)

// AuthRoutes registers the public auth endpoints (register, login, refresh, logout)
// and the protected profile endpoint.
func AuthRoutes(r fiber.Router, repo Repository, tokenRepo TokenRepository) {
	h := Handler{Repo: repo, TokenRepo: tokenRepo}

	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/refresh", h.Refresh)
	r.Post("/logout", h.Logout)

	// /auth/me requires a valid JWT
	r.Get("/me", middleware.JWT(), h.Profile)
}

// Routes registers admin-level user CRUD endpoints.
// These should be mounted behind JWT middleware by the caller.
func Routes(r fiber.Router, repo Repository, tokenRepo TokenRepository) {
	h := Handler{Repo: repo, TokenRepo: tokenRepo}

	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Put("/:id", h.Update)
	r.Delete("/:id", h.Delete)
}
