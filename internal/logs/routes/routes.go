package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/logs/handlers"
	"github.com/nexbic/platform/internal/middleware"
)

func RegisterLogsRoutes(router fiber.Router, handler *handlers.LogsHandler, authMW *middleware.AuthMiddleware) {
	logs := router.Group("/logs", authMW.RequireAuth())

	logs.Get("/", handler.GetLogs)
	logs.Get("/errors", handler.GetErrorLogs)
}
