package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/auth/handlers"
	"github.com/nexbic/platform/internal/middleware"
)

func RegisterAuthRoutes(router fiber.Router, handler *handlers.AuthHandler, authMW *middleware.AuthMiddleware) {
	// Public
	router.Post("/auth/login", handler.Login)
	router.Post("/auth/register", handler.Register)
	router.Post("/auth/refresh", handler.RefreshToken)
	router.Post("/auth/forgot-password", handler.ForgotPassword)
	router.Post("/auth/reset-password", handler.ResetPassword)
	router.Post("/auth/verify-email", handler.VerifyEmail)

	// Authenticated
	router.Get("/auth/me", authMW.RequireAuth(), handler.Me)
	router.Patch("/auth/password", authMW.RequireAuth(), handler.UpdatePassword)
	router.Post("/auth/verify-email/send", authMW.RequireAuth(), handler.SendVerification)

	// 2FA TOTP
	router.Post("/auth/totp/enable", authMW.RequireAuth(), handler.EnableTOTP)
	router.Post("/auth/totp/verify", authMW.RequireAuth(), handler.VerifyTOTP)
	router.Post("/auth/totp/disable", authMW.RequireAuth(), handler.DisableTOTP)

	// Devices & Sessions
	router.Get("/auth/devices", authMW.RequireAuth(), handler.ListDevices)
	router.Delete("/auth/devices/:id", authMW.RequireAuth(), handler.DeleteDevice)

	// Security Events
	router.Get("/auth/security-events", authMW.RequireAuth(), handler.ListSecurityEvents)

	// API Keys
	router.Get("/auth/api-keys", authMW.RequireAuth(), handler.ListAPIKeys)
	router.Post("/auth/api-keys", authMW.RequireAuth(), handler.CreateAPIKey)
	router.Delete("/auth/api-keys/:id", authMW.RequireAuth(), handler.RevokeAPIKey)

	// Admin User Management
	admin := router.Group("/admin/users", authMW.RequireAuth(), authMW.RequireRole("super_admin"))
	admin.Get("/", handler.ListUsers)
	admin.Get("/:id", handler.GetUser)
	admin.Post("/", handler.CreateUser)
	admin.Patch("/:id", handler.UpdateUser)
	admin.Patch("/:id/password", handler.UpdateUserPassword)
	admin.Delete("/:id", handler.DeleteUser)
}
