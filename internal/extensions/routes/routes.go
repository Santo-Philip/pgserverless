package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/extensions/handlers"
	"github.com/nexbic/platform/internal/middleware"
)

func RegisterExtensionRoutes(router fiber.Router, handler *handlers.ExtensionsHandler, authMW *middleware.AuthMiddleware) {
	g := router.Group("/extensions", authMW.RequireAuth())
	g.Get("/", handler.List)
	g.Post("/", handler.Install)
	g.Delete("/:name", handler.Uninstall)
}
