package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/database/monitoring/models"
	"github.com/nexbic/platform/internal/database/monitoring/service"
	"github.com/nexbic/platform/pkg/response"
)

type MonitoringHandler struct {
	svc *service.MonitoringService
}

func NewMonitoringHandler(svc *service.MonitoringService) *MonitoringHandler {
	return &MonitoringHandler{svc: svc}
}

func (h *MonitoringHandler) GetActiveSessions(c *fiber.Ctx) error {
	sessions, err := h.svc.GetActiveSessions(c.Context())
	if err != nil {
		return response.InternalError(c, "failed to retrieve active sessions")
	}

	return response.OK(c, sessions)
}

func (h *MonitoringHandler) GetSlowQueries(c *fiber.Ctx) error {
	minSeconds := c.QueryFloat("min_seconds", 5)
	if minSeconds <= 0 {
		minSeconds = 5
	}

	queries, err := h.svc.GetSlowQueries(c.Context(), minSeconds)
	if err != nil {
		return response.InternalError(c, "failed to retrieve slow queries")
	}

	return response.OK(c, queries)
}

func (h *MonitoringHandler) GetLocks(c *fiber.Ctx) error {
	locks, err := h.svc.GetLocks(c.Context())
	if err != nil {
		return response.InternalError(c, "failed to retrieve locks")
	}

	return response.OK(c, locks)
}

func (h *MonitoringHandler) GetWaitingQueries(c *fiber.Ctx) error {
	queries, err := h.svc.GetWaitingQueries(c.Context())
	if err != nil {
		return response.InternalError(c, "failed to retrieve waiting queries")
	}

	return response.OK(c, queries)
}

func (h *MonitoringHandler) GetQueryStats(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 50)

	stats, err := h.svc.GetQueryStats(c.Context(), limit)
	if err != nil {
		return response.InternalError(c, "failed to retrieve query statistics")
	}

	return response.OK(c, stats)
}

func (h *MonitoringHandler) GetConnectionStats(c *fiber.Ctx) error {
	stats, err := h.svc.GetConnectionStats(c.Context())
	if err != nil {
		return response.InternalError(c, "failed to retrieve connection statistics")
	}

	return response.OK(c, stats)
}

func (h *MonitoringHandler) GetCacheStats(c *fiber.Ctx) error {
	stats, err := h.svc.GetCacheStats(c.Context())
	if err != nil {
		return response.InternalError(c, "failed to retrieve cache statistics")
	}

	return response.OK(c, stats)
}

func (h *MonitoringHandler) GetDatabaseStats(c *fiber.Ctx) error {
	stats, err := h.svc.GetDatabaseStats(c.Context())
	if err != nil {
		return response.InternalError(c, "failed to retrieve database statistics")
	}

	return response.OK(c, stats)
}

func (h *MonitoringHandler) GetTableStats(c *fiber.Ctx) error {
	schema := c.Query("schema", "")
	limit := c.QueryInt("limit", 50)

	tables, err := h.svc.GetTableStats(c.Context(), schema, limit)
	if err != nil {
		return response.InternalError(c, "failed to retrieve table statistics")
	}

	return response.OK(c, tables)
}

func (h *MonitoringHandler) GetIndexStats(c *fiber.Ctx) error {
	schema := c.Query("schema", "")
	limit := c.QueryInt("limit", 50)

	indexes, err := h.svc.GetIndexStats(c.Context(), schema, limit)
	if err != nil {
		return response.InternalError(c, "failed to retrieve index statistics")
	}

	return response.OK(c, indexes)
}

func (h *MonitoringHandler) TerminateSession(c *fiber.Ctx) error {
	var req models.TerminateRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if req.PID <= 0 {
		return response.BadRequest(c, "pid is required and must be a positive integer")
	}

	result, err := h.svc.TerminateSession(c.Context(), req.PID)
	if err != nil {
		return response.InternalError(c, "failed to terminate session")
	}

	return response.OK(c, result)
}

func (h *MonitoringHandler) CancelQuery(c *fiber.Ctx) error {
	var req models.CancelRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if req.PID <= 0 {
		return response.BadRequest(c, "pid is required and must be a positive integer")
	}

	result, err := h.svc.CancelQuery(c.Context(), req.PID)
	if err != nil {
		return response.InternalError(c, "failed to cancel query")
	}

	return response.OK(c, result)
}
