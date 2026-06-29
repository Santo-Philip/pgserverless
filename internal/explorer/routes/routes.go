package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/explorer/handlers"
	"github.com/nexbic/platform/internal/middleware"
)

func RegisterExplorerRoutes(router fiber.Router, handler *handlers.ExplorerHandler, authMW *middleware.AuthMiddleware) {
	g := router.Group("/explorer", authMW.RequireAuth())

	g.Get("/schemas", handler.ListSchemas)
	g.Get("/schemas/:schema/:resource", handler.ListResource)
	g.Get("/schemas/:schema/tables/:table", handler.GetTableDetails)
	g.Get("/extensions", handler.ListExtensions)
}
