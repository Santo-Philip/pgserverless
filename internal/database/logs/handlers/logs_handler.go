package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/database/logs/models"
	"github.com/nexbic/platform/internal/database/logs/service"
	"github.com/nexbic/platform/pkg/helpers"
	"github.com/nexbic/platform/pkg/response"
)

type LogsHandler struct {
	svc *service.LogsService
}

func NewLogsHandler(svc *service.LogsService) *LogsHandler {
	return &LogsHandler{svc: svc}
}

func (h *LogsHandler) GetLogs(c *fiber.Ctx) error {
	q, err := parseLogQuery(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	resp, err := h.svc.GetLogs(c.Context(), q)
	if err != nil {
		return response.InternalError(c, "failed to retrieve logs")
	}

	return response.Paginated(c, resp.Entries, resp.Total, resp.Limit, resp.Offset)
}

func (h *LogsHandler) GetQueryLogs(c *fiber.Ctx) error {
	p := helpers.ParsePagination(c)
	resp, err := h.svc.GetQueryLogs(c.Context(), p.Limit, p.Offset)
	if err != nil {
		return response.InternalError(c, "failed to retrieve query logs")
	}
	return response.Paginated(c, resp.Entries, resp.Total, resp.Limit, resp.Offset)
}

func (h *LogsHandler) GetErrorLogs(c *fiber.Ctx) error {
	p := helpers.ParsePagination(c)
	severity := c.Query("severity", "")
	resp, err := h.svc.GetErrorLogs(c.Context(), severity, p.Limit, p.Offset)
	if err != nil {
		return response.InternalError(c, "failed to retrieve error logs")
	}
	return response.Paginated(c, resp.Entries, resp.Total, resp.Limit, resp.Offset)
}

func (h *LogsHandler) GetAuthLogs(c *fiber.Ctx) error {
	p := helpers.ParsePagination(c)
	resp, err := h.svc.GetAuthLogs(c.Context(), p.Limit, p.Offset)
	if err != nil {
		return response.InternalError(c, "failed to retrieve auth logs")
	}
	return response.Paginated(c, resp.Entries, resp.Total, resp.Limit, resp.Offset)
}

func (h *LogsHandler) GetConnectionLogs(c *fiber.Ctx) error {
	p := helpers.ParsePagination(c)
	resp, err := h.svc.GetConnectionLogs(c.Context(), p.Limit, p.Offset)
	if err != nil {
		return response.InternalError(c, "failed to retrieve connection logs")
	}
	return response.Paginated(c, resp.Entries, resp.Total, resp.Limit, resp.Offset)
}

func parseLogQuery(c *fiber.Ctx) (models.LogQuery, error) {
	q := models.LogQuery{
		Severity: c.Query("severity"),
		Database: c.Query("database"),
		User:     c.Query("user"),
		Search:   c.Query("search"),
	}

	p := helpers.ParsePagination(c)
	q.Limit = p.Limit
	q.Offset = p.Offset

	if startStr := c.Query("start_time"); startStr != "" {
		t, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			return q, err
		}
		q.StartTime = &t
	}

	if endStr := c.Query("end_time"); endStr != "" {
		t, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			return q, err
		}
		q.EndTime = &t
	}

	return q, nil
}
