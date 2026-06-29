package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/files/service"
	"github.com/nexbic/platform/pkg/helpers"
	"github.com/nexbic/platform/pkg/response"
)

type FilesHandler struct {
	svc *service.FilesService
}

func NewFilesHandler(svc *service.FilesService) *FilesHandler {
	return &FilesHandler{svc: svc}
}

func (h *FilesHandler) Upload(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)

	file, err := c.FormFile("file")
	if err != nil {
		return response.BadRequest(c, "file is required")
	}

	fh, err := file.Open()
	if err != nil {
		return response.InternalError(c, "failed to open file")
	}
	defer fh.Close()

	mimeType := file.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	f, err := h.svc.Upload(c.Context(), userID, file.Filename, mimeType, fh)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Created(c, f)
}

func (h *FilesHandler) List(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	files, err := h.svc.List(c.Context(), userID)
	if err != nil {
		return response.InternalError(c, "failed to list files")
	}
	return response.OK(c, files)
}

func (h *FilesHandler) Download(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	fileID := c.Params("id")

	f, reader, err := h.svc.Get(c.Context(), userID, fileID)
	if err != nil {
		return response.NotFound(c, "file not found")
	}
	defer reader.Close()

	c.Set("Content-Type", f.MimeType)
	c.Set("Content-Disposition", "attachment; filename="+f.Name)
	return c.SendStream(reader)
}

func (h *FilesHandler) Delete(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	fileID := c.Params("id")

	if err := h.svc.Delete(c.Context(), userID, fileID); err != nil {
		return response.NotFound(c, "file not found")
	}
	return response.NoContent(c)
}
