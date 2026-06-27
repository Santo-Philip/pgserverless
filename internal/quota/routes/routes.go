package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/middleware"
	"github.com/nexbic/platform/internal/quota/handlers"
)

func RegisterQuotaRoutes(router fiber.Router, handler *handlers.QuotaHandler, authMW *middleware.AuthMiddleware) {
	router.Get("/projects/:project_id/quota", authMW.RequireAuth(), handler.GetQuota)
}
