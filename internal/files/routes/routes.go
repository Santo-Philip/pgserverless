package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/files/handlers"
	"github.com/nexbic/platform/internal/middleware"
)

func RegisterFilesRoutes(router fiber.Router, handler *handlers.FilesHandler, authMW *middleware.AuthMiddleware) {
	g := router.Group("/files", authMW.RequireAuth())
	g.Get("/", handler.List)
	g.Post("/upload", handler.Upload)
	g.Get("/:id/download", handler.Download)
	g.Delete("/:id", handler.Delete)
}
