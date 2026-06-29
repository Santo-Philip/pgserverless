package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/middleware"
	"github.com/nexbic/platform/internal/projects/handlers"
)

func RegisterProjectsRoutes(router fiber.Router, handler *handlers.ProjectsHandler, authMW *middleware.AuthMiddleware) {
	g := router.Group("/projects", authMW.RequireAuth())
	g.Get("/", handler.List)
	g.Post("/", handler.Create)
	g.Get("/:projectId", handler.Get)
	g.Patch("/:projectId", handler.Update)
	g.Delete("/:projectId", handler.Delete)
}
