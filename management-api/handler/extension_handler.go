package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/management-api/service"
	"github.com/nexbic/platform/shared/utils"
)

type ExtensionHandler struct {
	extensionService *service.ExtensionService
}

func NewExtensionHandler(extensionService *service.ExtensionService) *ExtensionHandler {
	return &ExtensionHandler{extensionService: extensionService}
}

type ToggleExtensionRequest struct {
	Name    string `json:"name"`
	Install bool   `json:"install"`
}

func (h *ExtensionHandler) List(c *fiber.Ctx) error {
	extensions, err := h.extensionService.ListExtensions(c.Context())
	if err != nil {
		return utils.InternalError(c, "Failed to list extensions")
	}
	return utils.OK(c, extensions)
}

func (h *ExtensionHandler) Toggle(c *fiber.Ctx) error {
	var req ToggleExtensionRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body")
	}

	if req.Name == "" {
		return utils.BadRequest(c, "Extension name is required")
	}

	if err := h.extensionService.ToggleExtension(c.Context(), req.Name, req.Install); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	action := "disabled"
	if req.Install {
		action = "enabled"
	}
	return utils.OK(c, fiber.Map{"message": "Extension " + action + " successfully", "name": req.Name, "installed": req.Install})
}
