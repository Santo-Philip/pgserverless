package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/dashboard/handlers"
	"github.com/nexbic/platform/internal/middleware"
)

func RegisterDashboardRoutes(router fiber.Router, handler *handlers.DashboardHandler, authMW *middleware.AuthMiddleware) {
	admin := router.Group("/admin/dashboard", authMW.RequireAuth(), authMW.RequireRole("super_admin"))
	admin.Get("/overview", handler.Overview)
	admin.Get("/stats", handler.Stats)
	admin.Get("/schemas", handler.Schemas)
}
