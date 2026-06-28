package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/logs/handlers"
	"github.com/nexbic/platform/internal/middleware"
)

func RegisterLogsRoutes(router fiber.Router, handler *handlers.LogsHandler, authMW *middleware.AuthMiddleware) {
	logs := router.Group("/logs")

	logs.Get("/", authMW.RequireAuth(), handler.GetLogs)
	logs.Get("/errors", authMW.RequireAuth(), handler.GetErrorLogs)
	logs.Get("/queries", authMW.RequireAuth(), handler.GetQueryLogs)
	logs.Get("/auth", authMW.RequireAuth(), handler.GetAuthLogs)
	logs.Get("/connections", authMW.RequireAuth(), handler.GetConnectionLogs)
}
