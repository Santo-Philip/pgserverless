package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nexbic/platform/management-api/repository"
	"github.com/nexbic/platform/shared/database"
	"github.com/nexbic/platform/shared/utils"
)

type LogHandler struct {
	db      *database.DB
	appRepo *repository.AppRepository
}

type AppLogEntry struct {
	ID             uuid.UUID `json:"id"`
	AppID          uuid.UUID `json:"app_id"`
	APIKeyID       *uuid.UUID `json:"api_key_id,omitempty"`
	UserID         *uuid.UUID `json:"user_id,omitempty"`
	Method         string    `json:"method"`
	Path           string    `json:"path"`
	StatusCode     int       `json:"status_code"`
	ResponseTimeMs int       `json:"response_time_ms"`
	CreatedAt      time.Time `json:"created_at"`
}

func NewLogHandler(db *database.DB, appRepo *repository.AppRepository) *LogHandler {
	return &LogHandler{db: db, appRepo: appRepo}
}

func (h *LogHandler) ListAppLogs(c *fiber.Ctx) error {
	appID, err := utils.ParseUUIDParam(c, "id", "app")
	if err != nil {
		return utils.BadRequest(c, "invalid app id")
	}

	app, err := h.appRepo.GetByID(c.Context(), appID)
	if err != nil || app == nil {
		return utils.NotFound(c, "app not found")
	}

	limit := c.QueryInt("limit", 100)
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	offset := c.QueryInt("offset", 0)
	if offset < 0 {
		offset = 0
	}

	rows, err := h.db.Pool.Query(c.Context(), `
		SELECT id, app_id, api_key_id, user_id, method, path, status_code, response_time_ms, created_at
		FROM usage_logs
		WHERE app_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, appID, limit, offset)
	if err != nil {
		return utils.InternalError(c, "failed to fetch logs")
	}
	defer rows.Close()

	var logs []AppLogEntry
	for rows.Next() {
		var entry AppLogEntry
		if err := rows.Scan(&entry.ID, &entry.AppID, &entry.APIKeyID, &entry.UserID, &entry.Method, &entry.Path, &entry.StatusCode, &entry.ResponseTimeMs, &entry.CreatedAt); err != nil {
			continue
		}
		logs = append(logs, entry)
	}

	if logs == nil {
		logs = []AppLogEntry{}
	}

	return utils.OK(c, logs)
}

func (h *LogHandler) ListGlobalLogs(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 100)
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	offset := c.QueryInt("offset", 0)
	if offset < 0 {
		offset = 0
	}

	rows, err := h.db.Pool.Query(c.Context(), `
		SELECT id, app_id, api_key_id, user_id, method, path, status_code, response_time_ms, created_at
		FROM usage_logs
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		if err == pgx.ErrNoRows {
			return utils.OK(c, []AppLogEntry{})
		}
		return utils.InternalError(c, "failed to fetch logs")
	}
	defer rows.Close()

	var logs []AppLogEntry
	for rows.Next() {
		var entry AppLogEntry
		if err := rows.Scan(&entry.ID, &entry.AppID, &entry.APIKeyID, &entry.UserID, &entry.Method, &entry.Path, &entry.StatusCode, &entry.ResponseTimeMs, &entry.CreatedAt); err != nil {
			continue
		}
		logs = append(logs, entry)
	}

	if logs == nil {
		logs = []AppLogEntry{}
	}

	return utils.OK(c, logs)
}
