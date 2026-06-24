package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nexbic/platform/management-api/service"
	"github.com/nexbic/platform/shared/utils"
)

type DomainHandler struct {
	domainService *service.DomainService
}

func NewDomainHandler(domainService *service.DomainService) *DomainHandler {
	return &DomainHandler{domainService: domainService}
}

func (h *DomainHandler) List(c *fiber.Ctx) error {
	appID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid app id")
	}

	domains, err := h.domainService.ListByApp(c.Context(), appID)
	if err != nil {
		return utils.InternalError(c, "failed to list domains")
	}

	return utils.OK(c, domains)
}

func (h *DomainHandler) Create(c *fiber.Ctx) error {
	appID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid app id")
	}

	var req struct {
		Domain string `json:"domain"`
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}
	if req.Domain == "" {
		return utils.BadRequest(c, "domain is required")
	}

	domain, err := h.domainService.Create(c.Context(), appID, req.Domain)
	if err != nil {
		return utils.BadRequest(c, err.Error())
	}

	return utils.Created(c, domain)
}

func (h *DomainHandler) Delete(c *fiber.Ctx) error {
	domainID, err := uuid.Parse(c.Params("domainId"))
	if err != nil {
		return utils.BadRequest(c, "invalid domain id")
	}

	if err := h.domainService.Delete(c.Context(), domainID); err != nil {
		return utils.InternalError(c, "failed to delete domain")
	}

	return utils.OK(c, map[string]string{"message": "domain deleted"})
}

func (h *DomainHandler) Verify(c *fiber.Ctx) error {
	domainID, err := uuid.Parse(c.Params("domainId"))
	if err != nil {
		return utils.BadRequest(c, "invalid domain id")
	}

	if err := h.domainService.Verify(c.Context(), domainID); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	return utils.OK(c, map[string]string{"message": "domain verified"})
}
