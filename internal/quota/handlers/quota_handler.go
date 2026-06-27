package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/quota/service"
	"github.com/nexbic/platform/pkg/helpers"
	"github.com/nexbic/platform/pkg/response"
)

type QuotaHandler struct {
	svc *service.QuotaService
}

func NewQuotaHandler(svc *service.QuotaService) *QuotaHandler {
	return &QuotaHandler{svc: svc}
}

func (h *QuotaHandler) GetQuota(c *fiber.Ctx) error {
	projectID, err := helpers.ParseUUIDParam(c, "project_id", "project")
	if err != nil {
		return err
	}

	quota, limits, err := h.svc.GetQuota(c.Context(), projectID)
	if err != nil {
		return response.NotFound(c, err.Error())
	}

	return response.OK(c, fiber.Map{
		"usage":  quota,
		"limits": limits,
	})
}
