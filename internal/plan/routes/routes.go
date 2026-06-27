package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/middleware"
	"github.com/nexbic/platform/internal/plan/handlers"
)

func RegisterPlanRoutes(router fiber.Router, handler *handlers.PlanHandler, authMW *middleware.AuthMiddleware) {
	router.Post("/plans", authMW.RequireAuth(), authMW.RequireRole("admin"), handler.Create)
	router.Get("/plans", authMW.RequireAuth(), handler.List)
	router.Get("/plans/:id", authMW.RequireAuth(), handler.GetByID)
	router.Patch("/plans/:id", authMW.RequireAuth(), authMW.RequireRole("admin"), handler.Update)
}
