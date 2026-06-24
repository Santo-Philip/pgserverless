package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nexbic/platform/management-api/service"
	"github.com/nexbic/platform/shared/models"
	"github.com/nexbic/platform/shared/utils"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req models.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	result, err := h.authService.Register(c.Context(), req)
	if err != nil {
		return utils.BadRequest(c, err.Error())
	}

	return utils.Created(c, result)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	result, err := h.authService.Login(c.Context(), req)
	if err != nil {
		return utils.BadRequest(c, err.Error())
	}

	return utils.OK(c, result)
}

func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req models.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	result, err := h.authService.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		return utils.BadRequest(c, err.Error())
	}

	return utils.OK(c, result)
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	isAdmin, _ := c.Locals("is_super_admin").(bool)
	return utils.OK(c, map[string]any{
		"user_id":        userID,
		"email":          c.Locals("email"),
		"is_super_admin": isAdmin,
	})
}

func (h *AuthHandler) ListUsers(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	users, total, err := h.authService.ListUsers(c.Context(), limit, offset)
	if err != nil {
		return utils.InternalError(c, "failed to list users")
	}

	return utils.Paginated(c, users, total, limit, offset)
}

func (h *AuthHandler) SuspendUser(c *fiber.Ctx) error {
	actorIDStr := c.Locals("user_id").(string)
	actorID, err := uuid.Parse(actorIDStr)
	if err != nil {
		return utils.BadRequest(c, "invalid actor")
	}

	targetID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return utils.BadRequest(c, "invalid user id")
	}

	if err := h.authService.SuspendUser(c.Context(), actorID, targetID); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	return utils.OK(c, map[string]string{"status": "suspended"})
}

func (h *AuthHandler) ActivateUser(c *fiber.Ctx) error {
	targetID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return utils.BadRequest(c, "invalid user id")
	}

	if err := h.authService.ActivateUser(c.Context(), targetID); err != nil {
		return utils.InternalError(c, err.Error())
	}

	return utils.OK(c, map[string]string{"status": "active"})
}

func (h *AuthHandler) GetUser(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return utils.BadRequest(c, "invalid user id")
	}

	user, err := h.authService.GetUser(c.Context(), id)
	if err != nil {
		return utils.InternalError(c, "failed to get user")
	}
	if user == nil {
		return utils.NotFound(c, "user not found")
	}

	return utils.OK(c, user)
}
