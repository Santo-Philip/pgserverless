package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/explorer/handlers"
	"github.com/nexbic/platform/internal/middleware"
)

func RegisterExplorerRoutes(router fiber.Router, handler *handlers.ExplorerHandler, authMW *middleware.AuthMiddleware) {
	g := router.Group("/explorer", authMW.RequireAuth())

	g.Get("/schemas", handler.ListSchemas)
	g.Get("/schemas/:schema/tables", handler.ListTables)
	g.Get("/schemas/:schema/tables/:table", handler.GetTableDetails)
	g.Get("/schemas/:schema/views", handler.ListViews)
	g.Get("/schemas/:schema/functions", handler.ListFunctions)
	g.Get("/schemas/:schema/procedures", handler.ListProcedures)
	g.Get("/schemas/:schema/triggers", handler.ListTriggers)
	g.Get("/schemas/:schema/indexes", handler.ListIndexes)
	g.Get("/schemas/:schema/constraints", handler.ListConstraints)
	g.Get("/schemas/:schema/sequences", handler.ListSequences)
	g.Get("/schemas/:schema/materialized-views", handler.ListMaterializedViews)
	g.Get("/extensions", handler.ListExtensions)
}
