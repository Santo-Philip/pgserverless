package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/apikey/handlers"
	"github.com/nexbic/platform/internal/middleware"
)

func RegisterAPIKeyRoutes(router fiber.Router, handler *handlers.APIKeyHandler, authMW *middleware.AuthMiddleware) {
	router.Post("/api-keys", authMW.RequireAuth(), handler.Create)
	router.Get("/api-keys", authMW.RequireAuth(), handler.List)
	router.Get("/projects/:project_id/api-keys", authMW.RequireAuth(), handler.ListByProject)
	router.Delete("/api-keys/:id", authMW.RequireAuth(), handler.Revoke)
}
