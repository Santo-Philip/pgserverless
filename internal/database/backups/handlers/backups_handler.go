package handlers

import (
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/database/backups/models"
	"github.com/nexbic/platform/internal/database/backups/service"
	"github.com/nexbic/platform/pkg/helpers"
	"github.com/nexbic/platform/pkg/response"
)

type BackupHandler struct {
	svc *service.BackupService
}

func NewBackupHandler(svc *service.BackupService) *BackupHandler {
	return &BackupHandler{svc: svc}
}

func (h *BackupHandler) CreateBackup(c *fiber.Ctx) error {
	var req models.CreateBackupRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if req.Name == "" || req.DatabaseName == "" {
		return response.BadRequest(c, "name and database_name are required")
	}
	if req.Type == "" {
		req.Type = "manual"
	}

	userID := helpers.GetUserID(c)
	backup, err := h.svc.CreateBackup(c.Context(), &req, userID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Created(c, backup)
}

func (h *BackupHandler) ListBackups(c *fiber.Ctx) error {
	p := helpers.ParsePagination(c)
	history, err := h.svc.ListBackups(c.Context(), p.Limit, p.Offset)
	if err != nil {
		return response.InternalError(c, "failed to list backups")
	}

	return response.Paginated(c, history.Data, history.Total, history.Limit, history.Offset)
}

func (h *BackupHandler) GetBackup(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "backup")
	if err != nil {
		return err
	}

	backup, err := h.svc.GetBackup(c.Context(), id)
	if err != nil || backup == nil {
		return response.NotFound(c, "backup not found")
	}

	return response.OK(c, backup)
}

func (h *BackupHandler) DeleteBackup(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "backup")
	if err != nil {
		return err
	}

	if err := h.svc.DeleteBackup(c.Context(), id); err != nil {
		return response.Conflict(c, err.Error())
	}

	return response.NoContent(c)
}

func (h *BackupHandler) RestoreBackup(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "backup")
	if err != nil {
		return err
	}

	var req models.RestoreRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	req.BackupID = id.String()

	userID := helpers.GetUserID(c)
	if err := h.svc.RestoreBackup(c.Context(), &req, userID); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.OK(c, fiber.Map{"message": "restore initiated"})
}

func (h *BackupHandler) DownloadBackup(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "backup")
	if err != nil {
		return err
	}

	filePath, err := h.svc.GetBackupFilePath(c.Context(), id)
	if err != nil || filePath == "" {
		return response.NotFound(c, "backup file not found")
	}

	c.Set("Content-Type", "application/octet-stream")
	c.Set("Content-Disposition", "attachment; filename=\""+filepath.Base(filePath)+"\"")
	return c.SendFile(filePath)
}

func (h *BackupHandler) VerifyBackup(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "backup")
	if err != nil {
		return err
	}

	backup, err := h.svc.VerifyBackup(c.Context(), id)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.OK(c, backup)
}
