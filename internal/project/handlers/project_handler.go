package handlers

import (
	"github.com/gofiber/fiber/v2"
	projectdto "github.com/nexbic/platform/internal/project/dto"
	projectservice "github.com/nexbic/platform/internal/project/service"
	projectvalidation "github.com/nexbic/platform/internal/project/validation"
	"github.com/nexbic/platform/pkg/helpers"
	"github.com/nexbic/platform/pkg/response"
)

type ProjectHandler struct {
	service *projectservice.ProjectService
}

func NewProjectHandler(service *projectservice.ProjectService) *ProjectHandler {
	return &ProjectHandler{service: service}
}

func (h *ProjectHandler) Create(c *fiber.Ctx) error {
	var req projectdto.CreateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if v := projectvalidation.ValidateCreate(&req); v.HasErrors() {
		return response.BadRequest(c, v.Error())
	}

	project, err := h.service.Create(c.Context(), &req)
	if err != nil {
		return response.Conflict(c, err.Error())
	}

	return response.Created(c, project)
}

func (h *ProjectHandler) GetByID(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "project")
	if err != nil {
		return err
	}

	project, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return response.NotFound(c, "project not found")
	}

	return response.OK(c, project)
}

func (h *ProjectHandler) List(c *fiber.Ctx) error {
	p := helpers.ParsePagination(c)

	projects, total, err := h.service.List(c.Context(), p.Limit, p.Offset)
	if err != nil {
		return response.InternalError(c, "failed to list projects")
	}

	return response.Paginated(c, projects, total, p.Limit, p.Offset)
}

func (h *ProjectHandler) Update(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "project")
	if err != nil {
		return err
	}

	var req projectdto.UpdateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if v := projectvalidation.ValidateUpdate(&req); v.HasErrors() {
		return response.BadRequest(c, v.Error())
	}

	if err := h.service.Update(c.Context(), id, &req); err != nil {
		return response.NotFound(c, err.Error())
	}

	project, _ := h.service.GetByID(c.Context(), id)
	return response.OK(c, project)
}

func (h *ProjectHandler) Delete(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "project")
	if err != nil {
		return err
	}

	if err := h.service.Delete(c.Context(), id); err != nil {
		return response.NotFound(c, err.Error())
	}

	return response.NoContent(c)
}
