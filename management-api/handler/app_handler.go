package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/management-api/service"
	"github.com/nexbic/platform/shared/utils"
)

type AppHandler struct {
	appService     *service.AppService
	settingsService *service.SettingsService
}

func NewAppHandler(appService *service.AppService, settingsService *service.SettingsService) *AppHandler {
	return &AppHandler{appService: appService, settingsService: settingsService}
}

func (h *AppHandler) Create(c *fiber.Ctx) error {
	var req service.CreateAppRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	userID, ok := utils.GetUserID(c)
	if !ok {
		return utils.BadRequest(c, "invalid user id")
	}

	result, err := h.appService.CreateApp(c.Context(), req, userID)
	if err != nil {
		return utils.BadRequest(c, err.Error())
	}

	return utils.Created(c, result)
}

func (h *AppHandler) GetByID(c *fiber.Ctx) error {
	id, err := utils.ParseUUIDParam(c, "id", "app")
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
	p := utils.ParsePagination(c)

	apps, total, err := h.appService.ListApps(c.Context(), nil, p.Limit, p.Offset)
	if err != nil {
		return utils.InternalError(c, "failed to list apps")
	}

	return utils.Paginated(c, apps, total, p.Limit, p.Offset)
}

func (h *AppHandler) Delete(c *fiber.Ctx) error {
	id, err := utils.ParseUUIDParam(c, "id", "app")
	if err != nil {
		return utils.BadRequest(c, "invalid app id")
	}

	if err := h.appService.DeleteApp(c.Context(), id); err != nil {
		return utils.InternalError(c, "failed to delete app")
	}

	return utils.OK(c, map[string]string{"message": "app deleted"})
}

func (h *AppHandler) GetSettings(c *fiber.Ctx) error {
	settings, err := h.settingsService.Get(c.Context())
	if err != nil {
		return utils.InternalError(c, "failed to load settings")
	}

	return utils.OK(c, settings)
}

func (h *AppHandler) UpdateSettings(c *fiber.Ctx) error {
	var updates map[string]interface{}
	if err := c.BodyParser(&updates); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	if err := h.settingsService.Update(c.Context(), updates); err != nil {
		return utils.InternalError(c, "failed to save settings")
	}

	return utils.OK(c, map[string]string{"message": "settings updated"})
}
