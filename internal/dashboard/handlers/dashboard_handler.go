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
