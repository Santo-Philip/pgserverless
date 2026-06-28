package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/extensions/handlers"
	"github.com/nexbic/platform/internal/middleware"
)

func RegisterExtensionRoutes(router fiber.Router, handler *handlers.ExtensionsHandler, authMW *middleware.AuthMiddleware) {
	router.Get("/extensions", authMW.RequireAuth(), handler.List)
	router.Post("/extensions/install", authMW.RequireAuth(), handler.Install)
	router.Post("/extensions/uninstall", authMW.RequireAuth(), handler.Uninstall)
}
