package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/middleware"
	"github.com/nexbic/platform/internal/monitoring/handlers"
)

func RegisterMonitoringRoutes(router fiber.Router, handler *handlers.MonitoringHandler, authMW *middleware.AuthMiddleware) {
	m := router.Group("/monitoring", authMW.RequireAuth())

	m.Get("/sessions", handler.GetActiveSessions)
	m.Get("/queries/slow", handler.GetSlowQueries)
	m.Get("/queries/stats", handler.GetQueryStats)
	m.Get("/locks", handler.GetLocks)
	m.Get("/locks/waiting", handler.GetWaitingQueries)
	m.Post("/sessions/terminate", handler.TerminateSession)
	m.Post("/queries/cancel", handler.CancelQuery)
}
