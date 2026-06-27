package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/middleware"
	"github.com/nexbic/platform/internal/project/handlers"
)

func RegisterProjectRoutes(router fiber.Router, handler *handlers.ProjectHandler, authMW *middleware.AuthMiddleware) {
	router.Post("/projects", authMW.RequireAuth(), handler.Create)
	router.Get("/projects", authMW.RequireAuth(), handler.List)
	router.Get("/projects/:id", authMW.RequireAuth(), handler.GetByID)
	router.Patch("/projects/:id", authMW.RequireAuth(), handler.Update)
	router.Delete("/projects/:id", authMW.RequireAuth(), handler.Delete)
}
