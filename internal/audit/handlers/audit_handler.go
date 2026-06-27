package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/audit/service"
	"github.com/nexbic/platform/pkg/helpers"
	"github.com/nexbic/platform/pkg/response"
)

type AuditHandler struct {
	svc *service.AuditService
}

func NewAuditHandler(svc *service.AuditService) *AuditHandler {
	return &AuditHandler{svc: svc}
}

func (h *AuditHandler) List(c *fiber.Ctx) error {
	p := helpers.ParsePagination(c)
	logs, total, err := h.svc.List(c.Context(), p.Limit, p.Offset)
	if err != nil {
		return response.InternalError(c, "failed to list audit logs")
	}

	return response.Paginated(c, logs, total, p.Limit, p.Offset)
}

func (h *AuditHandler) ListByResource(c *fiber.Ctx) error {
	resource := c.Params("resource")
	resourceID := c.Params("resource_id")
	if resource == "" || resourceID == "" {
		return response.BadRequest(c, "resource and resource_id are required")
	}

	p := helpers.ParsePagination(c)
	logs, total, err := h.svc.ListByResource(c.Context(), resource, resourceID, p.Limit, p.Offset)
	if err != nil {
		return response.InternalError(c, "failed to list audit logs")
	}

	return response.Paginated(c, logs, total, p.Limit, p.Offset)
}
