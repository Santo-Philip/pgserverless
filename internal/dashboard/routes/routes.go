package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/dashboard/handlers"
	"github.com/nexbic/platform/internal/middleware"
)

func RegisterDashboardRoutes(router fiber.Router, handler *handlers.DashboardHandler, authMW *middleware.AuthMiddleware) {
	g := router.Group("/dashboard", authMW.RequireAuth())
	g.Get("/overview", handler.Overview)
	g.Get("/stats", handler.Stats)
	g.Get("/schemas", handler.Schemas)
}
