package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/schema/handlers"
)

func RegisterSchemaRoutes(router fiber.Router, handler *handlers.SchemaHandler) {
	g := router.Group("/schema")

	g.Post("/schemas", handler.CreateSchema)
	g.Delete("/schemas/:schema", handler.DropSchema)

	g.Post("/:schema/tables", handler.CreateTable)
	g.Delete("/:schema/tables/:table", handler.DropTable)

	g.Post("/:schema/tables/:table/columns", handler.AddColumn)
	g.Delete("/:schema/tables/:table/columns/:column", handler.DropColumn)
	g.Patch("/:schema/tables/:table/columns/:column", handler.AlterColumn)

	g.Post("/:schema/tables/:table/constraints", handler.AddConstraint)
	g.Delete("/:schema/tables/:table/constraints/:constraint", handler.DropConstraint)

	g.Get("/:schema/tables/:table/ddl", handler.GetTableDDL)

	g.Post("/:schema/indexes", handler.CreateIndex)
	g.Delete("/:schema/indexes/:name", handler.DropIndex)

	g.Post("/:schema/sequences", handler.CreateSequence)
	g.Delete("/:schema/sequences/:name", handler.DropSequence)
	g.Patch("/:schema/sequences/:name", handler.AlterSequence)
}
