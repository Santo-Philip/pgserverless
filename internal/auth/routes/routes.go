package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/auth/handlers"
	"github.com/nexbic/platform/internal/middleware"
)

func RegisterAuthRoutes(router fiber.Router, handler *handlers.AuthHandler, authMW *middleware.AuthMiddleware) {
	router.Post("/auth/login", handler.Login)
	router.Post("/auth/refresh", handler.RefreshToken)
	router.Get("/auth/me", authMW.RequireAuth(), handler.Me)
	router.Patch("/auth/password", authMW.RequireAuth(), handler.UpdatePassword)

	admin := router.Group("/admin/users", authMW.RequireAuth(), authMW.RequireRole("super_admin"))
	admin.Get("/", handler.ListUsers)
	admin.Get("/:id", handler.GetUser)
	admin.Post("/", handler.CreateUser)
	admin.Patch("/:id", handler.UpdateUser)
	admin.Patch("/:id/password", handler.UpdateUserPassword)
	admin.Delete("/:id", handler.DeleteUser)
}
