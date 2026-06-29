package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/database/extensions/service"
	"github.com/nexbic/platform/pkg/response"
)

type ExtensionsHandler struct {
	svc *service.ExtensionsService
}

func NewExtensionsHandler(svc *service.ExtensionsService) *ExtensionsHandler {
	return &ExtensionsHandler{svc: svc}
}

func (h *ExtensionsHandler) List(c *fiber.Ctx) error {
	extensions, err := h.svc.ListExtensions(c.Context())
	if err != nil {
		return response.InternalError(c, "failed to list extensions")
	}
	return response.OK(c, extensions)
}

type installRequest struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

func (h *ExtensionsHandler) Install(c *fiber.Ctx) error {
	var req installRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Name == "" {
		return response.BadRequest(c, "name is required")
	}
	if err := h.svc.InstallExtension(c.Context(), req.Name, req.Version); err != nil {
		return response.BadRequest(c, "install failed: "+err.Error())
	}
	return response.OK(c, fiber.Map{"extension": req.Name, "version": req.Version})
}

func (h *ExtensionsHandler) Uninstall(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return response.BadRequest(c, "name is required")
	}
	if err := h.svc.UninstallExtension(c.Context(), name); err != nil {
		return response.BadRequest(c, "uninstall failed: "+err.Error())
	}
	return response.OK(c, fiber.Map{"extension": name})
}
