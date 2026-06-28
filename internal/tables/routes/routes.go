package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/tables/handlers"
)

func RegisterTablesRoutes(router fiber.Router, handler *handlers.TablesHandler) {
	tables := router.Group("/tables")
	tables.Get("/:schema/:table", handler.Query)
	tables.Post("/:schema/:table/rows", handler.Insert)
	tables.Patch("/:schema/:table/rows", handler.Update)
	tables.Delete("/:schema/:table/rows", handler.Delete)
	tables.Post("/:schema/:table/bulk", handler.BulkInsert)
	tables.Delete("/:schema/:table/bulk", handler.BulkDelete)
	tables.Post("/:schema/:table/search", handler.Search)
}
