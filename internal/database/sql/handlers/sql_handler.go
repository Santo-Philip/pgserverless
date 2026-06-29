package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nexbic/platform/internal/database/sql/models"
	sqlService "github.com/nexbic/platform/internal/database/sql/service"
	"github.com/nexbic/platform/pkg/helpers"
	"github.com/nexbic/platform/pkg/response"
)

type SQLHandler struct {
	service *sqlService.SQLService
}

func NewSQLHandler(service *sqlService.SQLService) *SQLHandler {
	return &SQLHandler{service: service}
}

func (h *SQLHandler) Execute(c *fiber.Ctx) error {
	var req models.ExecuteRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if req.Query == "" {
		return response.BadRequest(c, "query is required")
	}

	start := time.Now()
	result, err := h.service.ExecuteSQL(c.Context(), req.Query, req.Params)
	duration := time.Since(start).Milliseconds()

	userID := helpers.GetUserID(c)

	resp := models.ExecuteResponse{
		DurationMs: duration,
	}

	if err != nil {
		resp.Error = err.Error()
		go h.service.LogQuery(c.Context(), userID, req.Query, duration, 0, "error", err.Error())
		return response.OK(c, resp)
	}

	if result.Columns != nil {
		resp.Columns = result.Columns
		resp.Rows = result.Rows
		resp.RowCount = len(result.Rows)
	}

	if result.RowsAffected > 0 {
		resp.RowsAffected = result.RowsAffected
	}

	go h.service.LogQuery(c.Context(), userID, req.Query, duration, resp.RowCount+int(resp.RowsAffected), "success", "")

	return response.OK(c, resp)
}

func (h *SQLHandler) Explain(c *fiber.Ctx) error {
	var req models.ExecuteRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if req.Query == "" {
		return response.BadRequest(c, "query is required")
	}

	result, err := h.service.ExplainQuery(c.Context(), req.Query)
	if err != nil {
		return response.BadRequest(c, "explain failed: "+err.Error())
	}

	return response.OK(c, result)
}

type cancelRequest struct {
	PID int `json:"pid"`
}

func (h *SQLHandler) Cancel(c *fiber.Ctx) error {
	var req cancelRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if req.PID == 0 {
		return response.BadRequest(c, "pid is required")
	}

	if err := h.service.CancelQuery(c.Context(), req.PID); err != nil {
		return response.BadRequest(c, "cancel failed: "+err.Error())
	}

	return response.OK(c, fiber.Map{"canceled": req.PID})
}

func (h *SQLHandler) GetHistory(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	if userID == uuid.Nil {
		return response.Unauthorized(c, "unauthorized")
	}

	p := helpers.ParsePagination(c)
	history, total, err := h.service.GetQueryHistory(c.Context(), userID, p.Limit, p.Offset)
	if err != nil {
		return response.InternalError(c, "failed to fetch query history")
	}

	return response.Paginated(c, history, total, p.Limit, p.Offset)
}

func (h *SQLHandler) GetSaved(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	if userID == uuid.Nil {
		return response.Unauthorized(c, "unauthorized")
	}

	queries, err := h.service.GetSavedQueries(c.Context(), userID)
	if err != nil {
		return response.InternalError(c, "failed to fetch saved queries")
	}

	return response.OK(c, queries)
}

type saveQueryRequest struct {
	Name        string `json:"name"`
	Query       string `json:"query"`
	Description string `json:"description"`
	IsShared    bool   `json:"is_shared"`
}

func (h *SQLHandler) Save(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	if userID == uuid.Nil {
		return response.Unauthorized(c, "unauthorized")
	}

	var req saveQueryRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if req.Name == "" {
		return response.BadRequest(c, "name is required")
	}

	if req.Query == "" {
		return response.BadRequest(c, "query is required")
	}

	saved, err := h.service.SaveQuery(c.Context(), userID, req.Name, req.Query, req.Description, req.IsShared)
	if err != nil {
		return response.InternalError(c, "failed to save query")
	}

	return response.Created(c, saved)
}

func (h *SQLHandler) DeleteSaved(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	if userID == uuid.Nil {
		return response.Unauthorized(c, "unauthorized")
	}

	id, err := helpers.ParseUUIDParam(c, "id", "saved query")
	if err != nil {
		return err
	}

	if err := h.service.DeleteSavedQuery(c.Context(), id, userID); err != nil {
		return response.NotFound(c, "saved query not found")
	}

	return response.NoContent(c)
}
