package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/plan/dto"
	"github.com/nexbic/platform/internal/plan/service"
	"github.com/nexbic/platform/internal/plan/validation"
	"github.com/nexbic/platform/pkg/helpers"
	"github.com/nexbic/platform/pkg/response"
)

type PlanHandler struct {
	svc *service.PlanService
}

func NewPlanHandler(svc *service.PlanService) *PlanHandler {
	return &PlanHandler{svc: svc}
}

func (h *PlanHandler) Create(c *fiber.Ctx) error {
	var req dto.CreatePlanRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if v := validation.ValidateCreate(&req); v.HasErrors() {
		return response.BadRequest(c, v.Error())
	}

	plan, err := h.svc.Create(c.Context(), &req)
	if err != nil {
		return response.Conflict(c, err.Error())
	}

	return response.Created(c, plan)
}

func (h *PlanHandler) List(c *fiber.Ctx) error {
	plans, err := h.svc.List(c.Context())
	if err != nil {
		return response.InternalError(c, "failed to list plans")
	}

	return response.OK(c, plans)
}

func (h *PlanHandler) GetByID(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "plan")
	if err != nil {
		return err
	}

	plan, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		return response.NotFound(c, err.Error())
	}

	return response.OK(c, plan)
}

func (h *PlanHandler) Update(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "plan")
	if err != nil {
		return err
	}

	var req dto.UpdatePlanRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if err := h.svc.Update(c.Context(), id, &req); err != nil {
		return response.NotFound(c, err.Error())
	}

	plan, _ := h.svc.GetByID(c.Context(), id)
	return response.OK(c, plan)
}
