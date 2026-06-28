package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/middleware"
	"github.com/nexbic/platform/internal/monitoring/handlers"
)

func RegisterMonitoringRoutes(router fiber.Router, handler *handlers.MonitoringHandler, authMW *middleware.AuthMiddleware) {
	m := router.Group("/monitoring")

	m.Get("/sessions", authMW.RequireAuth(), handler.GetActiveSessions)
	m.Get("/slow-queries", authMW.RequireAuth(), handler.GetSlowQueries)
	m.Get("/locks", authMW.RequireAuth(), handler.GetLocks)
	m.Get("/waiting", authMW.RequireAuth(), handler.GetWaitingQueries)
	m.Get("/query-stats", authMW.RequireAuth(), handler.GetQueryStats)
	m.Get("/connections", authMW.RequireAuth(), handler.GetConnectionStats)
	m.Get("/cache", authMW.RequireAuth(), handler.GetCacheStats)
	m.Get("/databases", authMW.RequireAuth(), handler.GetDatabaseStats)
	m.Get("/tables", authMW.RequireAuth(), handler.GetTableStats)
	m.Get("/indexes", authMW.RequireAuth(), handler.GetIndexStats)
	m.Post("/terminate", authMW.RequireAuth(), handler.TerminateSession)
	m.Post("/cancel-query", authMW.RequireAuth(), handler.CancelQuery)
}
