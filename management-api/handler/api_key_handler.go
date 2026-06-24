package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nexbic/platform/management-api/repository"
	"github.com/nexbic/platform/management-api/service"
	"github.com/nexbic/platform/shared/models"
	"github.com/nexbic/platform/shared/utils"
)

type APIKeyHandler struct {
	service    *service.APIKeyService
	keyRepo    *repository.APIKeyRepository
}

func NewAPIKeyHandler(service *service.APIKeyService, keyRepo *repository.APIKeyRepository) *APIKeyHandler {
	return &APIKeyHandler{
		service: service,
		keyRepo: keyRepo,
	}
}

func (h *APIKeyHandler) Create(c *fiber.Ctx) error {
	appID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid app id")
	}

	var req models.CreateAPIKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	userID, _ := uuid.Parse(c.Locals("user_id").(string))

	result, err := h.service.CreateKey(c.Context(), appID, userID, req)
	if err != nil {
		return utils.BadRequest(c, err.Error())
	}

	return utils.Created(c, result)
}

func (h *APIKeyHandler) List(c *fiber.Ctx) error {
	appID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid app id")
	}

	keys, err := h.keyRepo.ListByApp(c.Context(), appID)
	if err != nil {
		return utils.InternalError(c, "failed to list keys")
	}

	return utils.OK(c, keys)
}

func (h *APIKeyHandler) Deactivate(c *fiber.Ctx) error {
	keyID, err := uuid.Parse(c.Params("keyId"))
	if err != nil {
		return utils.BadRequest(c, "invalid key id")
	}

	if err := h.keyRepo.Deactivate(c.Context(), keyID); err != nil {
		return utils.InternalError(c, "failed to deactivate key")
	}

	return utils.OK(c, map[string]string{"message": "key deactivated"})
}
