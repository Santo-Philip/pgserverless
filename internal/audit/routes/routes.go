package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/audit/handlers"
	"github.com/nexbic/platform/internal/middleware"
)

func RegisterAuditRoutes(router fiber.Router, handler *handlers.AuditHandler, authMW *middleware.AuthMiddleware) {
	router.Get("/audit-logs", authMW.RequireAuth(), handler.List)
	router.Get("/audit-logs/:resource/:resource_id", authMW.RequireAuth(), handler.ListByResource)
}
