package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/middleware"
	"github.com/nexbic/platform/internal/sql/handlers"
)

func RegisterSQLRoutes(router fiber.Router, handler *handlers.SQLHandler, authMW *middleware.AuthMiddleware) {
	sql := router.Group("/sql")

	sql.Post("/execute", authMW.RequireAuth(), handler.Execute)
	sql.Post("/explain", authMW.RequireAuth(), handler.Explain)
	sql.Post("/cancel", authMW.RequireAuth(), handler.Cancel)

	sql.Get("/history", authMW.RequireAuth(), handler.GetHistory)

	sql.Get("/saved", authMW.RequireAuth(), handler.GetSaved)
	sql.Post("/saved", authMW.RequireAuth(), handler.Save)
	sql.Delete("/saved/:id", authMW.RequireAuth(), handler.DeleteSaved)
}
