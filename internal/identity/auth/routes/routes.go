package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/identity/auth/handlers"
	"github.com/nexbic/platform/internal/middleware"
)

func RegisterAuthRoutes(router fiber.Router, handler *handlers.AuthHandler, authMW *middleware.AuthMiddleware) {
	RegisterDashboardAuthRoutes(router, handler, authMW)

	router.Get("/auth/oauth/google", handler.OAuthGoogle)
	router.Get("/auth/oauth/github", handler.OAuthGitHub)
	router.Post("/auth/oauth/callback", handler.OAuthCallback)
}

func RegisterDashboardAuthRoutes(router fiber.Router, handler *handlers.AuthHandler, authMW *middleware.AuthMiddleware) {
	router.Post("/auth/login", handler.Login)
	router.Post("/auth/refresh", handler.RefreshToken)

	router.Get("/auth/me", authMW.RequireAuth(), handler.Me)
	router.Patch("/auth/password", authMW.RequireAuth(), handler.UpdatePassword)

	router.Get("/auth/devices", authMW.RequireAuth(), handler.ListDevices)
	router.Delete("/auth/devices/:id", authMW.RequireAuth(), handler.DeleteDevice)

	router.Get("/auth/security-events", authMW.RequireAuth(), handler.ListSecurityEvents)

	router.Get("/auth/api-keys", authMW.RequireAuth(), handler.ListAPIKeys)
	router.Post("/auth/api-keys", authMW.RequireAuth(), handler.CreateAPIKey)
	router.Delete("/auth/api-keys/:id", authMW.RequireAuth(), handler.RevokeAPIKey)

	admin := router.Group("/admin/users", authMW.RequireAuth(), authMW.RequireRole("super_admin"))
	admin.Get("/", handler.ListUsers)
	admin.Get("/:id", handler.GetUser)
	admin.Post("/", handler.CreateUser)
	admin.Patch("/:id", handler.UpdateUser)
	admin.Patch("/:id/password", handler.UpdateUserPassword)
	admin.Delete("/:id", handler.DeleteUser)
}
