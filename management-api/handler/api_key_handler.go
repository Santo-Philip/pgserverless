package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/management-api/service"
	"github.com/nexbic/platform/shared/models"
	"github.com/nexbic/platform/shared/utils"
)

type APIKeyHandler struct {
	service *service.APIKeyService
}

func NewAPIKeyHandler(service *service.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{service: service}
}

func (h *APIKeyHandler) Create(c *fiber.Ctx) error {
	appID, err := utils.ParseUUIDParam(c, "id", "app")
	if err != nil {
		return utils.BadRequest(c, "invalid app id")
	}

	var req models.CreateAPIKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	userID, ok := utils.GetUserID(c)
	if !ok {
		return utils.BadRequest(c, "invalid user id")
	}

	result, err := h.service.CreateKey(c.Context(), appID, userID, req)
	if err != nil {
		return utils.BadRequest(c, err.Error())
	}

	return utils.Created(c, result)
}

func (h *APIKeyHandler) List(c *fiber.Ctx) error {
	appID, err := utils.ParseUUIDParam(c, "id", "app")
	if err != nil {
		return utils.BadRequest(c, "invalid app id")
	}

	keys, err := h.service.ListByApp(c.Context(), appID)
	if err != nil {
		return utils.InternalError(c, "failed to list keys")
	}

	return utils.OK(c, keys)
}

func (h *APIKeyHandler) Deactivate(c *fiber.Ctx) error {
	keyID, err := utils.ParseUUIDParam(c, "keyId", "key")
	if err != nil {
		return utils.BadRequest(c, "invalid key id")
	}

	if err := h.service.Deactivate(c.Context(), keyID); err != nil {
		return utils.InternalError(c, "failed to deactivate key")
	}

	return utils.OK(c, map[string]string{"message": "key deactivated"})
}
