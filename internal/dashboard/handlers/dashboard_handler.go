package handlers

import (
	"github.com/gofiber/fiber/v2"
	dashboardService "github.com/nexbic/platform/internal/dashboard/service"
	"github.com/nexbic/platform/pkg/response"
)

type DashboardHandler struct {
	service *dashboardService.DashboardService
}

func NewDashboardHandler(service *dashboardService.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) Overview(c *fiber.Ctx) error {
	overview, err := h.service.GetOverview(c.Context())
	if err != nil {
		return response.InternalError(c, "failed to fetch dashboard overview: "+err.Error())
	}

	return response.OK(c, overview)
}

func (h *DashboardHandler) Stats(c *fiber.Ctx) error {
	stats, err := h.service.GetOverview(c.Context())
	if err != nil {
		return response.InternalError(c, "failed to fetch stats: "+err.Error())
	}

	return response.OK(c, stats.Stats)
}

func (h *DashboardHandler) Schemas(c *fiber.Ctx) error {
	overview, err := h.service.GetOverview(c.Context())
	if err != nil {
		return response.InternalError(c, "failed to fetch schemas: "+err.Error())
	}

	return response.OK(c, overview.Schemas)
}
