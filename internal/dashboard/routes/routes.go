package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/dashboard/handlers"
	"github.com/nexbic/platform/internal/middleware"
)

func RegisterDashboardRoutes(router fiber.Router, handler *handlers.DashboardHandler, authMW *middleware.AuthMiddleware) {
	g := router.Group("/dashboard", authMW.RequireAuth(), authMW.RequireRole("super_admin", "dba"))

	g.Get("/overview", middleware.AuditLog(middleware.AuditRead, "dashboard_overview"), handler.Overview)
}
