package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/apikey/dto"
	"github.com/nexbic/platform/internal/apikey/service"
	"github.com/nexbic/platform/internal/apikey/validation"
	"github.com/nexbic/platform/pkg/helpers"
	"github.com/nexbic/platform/pkg/response"
)

type APIKeyHandler struct {
	svc *service.APIKeyService
}

func NewAPIKeyHandler(svc *service.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{svc: svc}
}

func (h *APIKeyHandler) Create(c *fiber.Ctx) error {
	var req dto.CreateKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if v := validation.ValidateCreate(&req); v.HasErrors() {
		return response.BadRequest(c, v.Error())
	}

	userID := helpers.GetUserID(c)
	resp, err := h.svc.CreateKey(c.Context(), &req, userID)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Created(c, resp)
}

func (h *APIKeyHandler) List(c *fiber.Ctx) error {
	p := helpers.ParsePagination(c)
	keys, total, err := h.svc.List(c.Context(), p.Limit, p.Offset)
	if err != nil {
		return response.InternalError(c, "failed to list keys")
	}

	return response.Paginated(c, keys, total, p.Limit, p.Offset)
}

func (h *APIKeyHandler) ListByProject(c *fiber.Ctx) error {
	projectID, err := helpers.ParseUUIDParam(c, "project_id", "project")
	if err != nil {
		return err
	}

	p := helpers.ParsePagination(c)
	keys, total, err := h.svc.ListByProject(c.Context(), projectID, p.Limit, p.Offset)
	if err != nil {
		return response.InternalError(c, "failed to list keys")
	}

	return response.Paginated(c, keys, total, p.Limit, p.Offset)
}

func (h *APIKeyHandler) Revoke(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "key")
	if err != nil {
		return err
	}

	if err := h.svc.Revoke(c.Context(), id); err != nil {
		return response.NotFound(c, err.Error())
	}

	return response.NoContent(c)
}
