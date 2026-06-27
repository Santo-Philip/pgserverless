package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/database/handlers"
	"github.com/nexbic/platform/internal/middleware"
)

func RegisterDatabaseRoutes(router fiber.Router, handler *handlers.DatabaseHandler, authMW *middleware.AuthMiddleware) {
	router.Post("/databases", authMW.RequireAuth(), handler.Create)
	router.Get("/databases/:id", authMW.RequireAuth(), handler.GetByID)
	router.Get("/projects/:project_id/databases", authMW.RequireAuth(), handler.ListByProject)
	router.Delete("/databases/:id", authMW.RequireAuth(), handler.Delete)

	router.Post("/databases/:id/sql", authMW.RequireAuth(), handler.RunSQL)
	router.Get("/databases/:id/tables", authMW.RequireAuth(), handler.ListTables)
	router.Post("/databases/:id/tables", authMW.RequireAuth(), handler.CreateTable)
	router.Get("/databases/:id/tables/:table", authMW.RequireAuth(), handler.GetTableData)
	router.Post("/databases/:id/tables/:table/rows", authMW.RequireAuth(), handler.InsertRow)
	router.Patch("/databases/:id/tables/:table/rows", authMW.RequireAuth(), handler.UpdateRow)
	router.Delete("/databases/:id/tables/:table/rows", authMW.RequireAuth(), handler.DeleteRow)
	router.Post("/databases/:id/tables/:table/columns", authMW.RequireAuth(), handler.AddColumn)

	router.Get("/extensions", authMW.RequireAuth(), handler.ListExtensions)
	router.Post("/extensions/toggle", authMW.RequireAuth(), handler.ToggleExtension)
}
