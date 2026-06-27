package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/auth/handlers"
	"github.com/nexbic/platform/internal/middleware"
)

func RegisterAuthRoutes(router fiber.Router, handler *handlers.AuthHandler, authMW *middleware.AuthMiddleware) {
	router.Post("/auth/register", handler.Register)
	router.Post("/auth/login", handler.Login)
	router.Post("/auth/refresh", handler.RefreshToken)
	router.Get("/auth/me", authMW.RequireAuth(), handler.Me)
}
