package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nexbic/platform/management-api/service"
	"github.com/nexbic/platform/shared/utils"
)

type AppHandler struct {
	appService *service.AppService
}

func NewAppHandler(appService *service.AppService) *AppHandler {
	return &AppHandler{appService: appService}
}

func (h *AppHandler) Create(c *fiber.Ctx) error {
	var req service.CreateAppRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	userID, _ := uuid.Parse(c.Locals("user_id").(string))

	result, err := h.appService.CreateApp(c.Context(), req, userID)
	if err != nil {
		return utils.BadRequest(c, err.Error())
	}

	return utils.Created(c, result)
}

func (h *AppHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid app id")
	}

	app, err := h.appService.GetApp(c.Context(), id)
	if err != nil {
		return utils.InternalError(c, "failed to get app")
	}
	if app == nil {
		return utils.NotFound(c, "app not found")
	}

	return utils.OK(c, app)
}

func (h *AppHandler) List(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	apps, total, err := h.appService.ListApps(c.Context(), nil, limit, offset)
	if err != nil {
		return utils.InternalError(c, "failed to list apps")
	}

	return utils.Paginated(c, apps, total, limit, offset)
}

func (h *AppHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid app id")
	}

	if err := h.appService.DeleteApp(c.Context(), id); err != nil {
		return utils.InternalError(c, "failed to delete app")
	}

	return utils.OK(c, map[string]string{"message": "app deleted"})
}

func (h *AppHandler) ListBackups(c *fiber.Ctx) error {
	return utils.OK(c, []interface{}{})
}

func (h *AppHandler) CreateBackup(c *fiber.Ctx) error {
	return utils.OK(c, map[string]string{"message": "backup started"})
}

func (h *AppHandler) GetSettings(c *fiber.Ctx) error {
	return utils.OK(c, map[string]interface{}{
		"region":          "us-east",
		"default_visibility": "public",
	})
}

func (h *AppHandler) UpdateSettings(c *fiber.Ctx) error {
	return utils.OK(c, map[string]string{"message": "settings updated"})
}
