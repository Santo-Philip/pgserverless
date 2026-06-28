package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	authdto "github.com/nexbic/platform/internal/auth/dto"
	authservice "github.com/nexbic/platform/internal/auth/service"
	authvalidation "github.com/nexbic/platform/internal/auth/validation"
	"github.com/nexbic/platform/pkg/helpers"
	"github.com/nexbic/platform/pkg/response"
)

type AuthHandler struct {
	authService *authservice.AuthService
}

func NewAuthHandler(authService *authservice.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req authdto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if v := authvalidation.ValidateLogin(&req); v.HasErrors() {
		return response.BadRequest(c, v.Error())
	}

	resp, err := h.authService.Login(c.Context(), &req)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	return response.OK(c, resp)
}

func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req authdto.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if v := authvalidation.ValidateRefreshToken(&req); v.HasErrors() {
		return response.BadRequest(c, v.Error())
	}

	resp, err := h.authService.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	return response.OK(c, resp)
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	if userID == uuid.Nil {
		return response.Unauthorized(c, "not authenticated")
	}

	user, err := h.authService.GetUser(c.Context(), userID)
	if err != nil || user == nil {
		return response.NotFound(c, "user not found")
	}

	return response.OK(c, user)
}

func (h *AuthHandler) ListUsers(c *fiber.Ctx) error {
	p := helpers.ParsePagination(c)
	users, total, err := h.authService.ListUsers(c.Context(), p.Limit, p.Offset)
	if err != nil {
		return response.InternalError(c, "failed to list users")
	}
	return response.Paginated(c, users, total, p.Limit, p.Offset)
}

func (h *AuthHandler) GetUser(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "user")
	if err != nil {
		return err
	}

	user, err := h.authService.GetUser(c.Context(), id)
	if err != nil || user == nil {
		return response.NotFound(c, "user not found")
	}

	return response.OK(c, user)
}

func (h *AuthHandler) CreateUser(c *fiber.Ctx) error {
	var req authdto.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if v := authvalidation.ValidateCreateUser(&req); v.HasErrors() {
		return response.BadRequest(c, v.Error())
	}

	user, err := h.authService.CreateUser(c.Context(), &req)
	if err != nil {
		return response.Conflict(c, err.Error())
	}

	return response.Created(c, user)
}

func (h *AuthHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "user")
	if err != nil {
		return err
	}

	var req authdto.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if v := authvalidation.ValidateUpdateUser(&req); v.HasErrors() {
		return response.BadRequest(c, v.Error())
	}

	user, err := h.authService.UpdateUser(c.Context(), id, &req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.OK(c, user)
}

func (h *AuthHandler) UpdatePassword(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	if userID == uuid.Nil {
		return response.Unauthorized(c, "not authenticated")
	}

	var req authdto.UpdatePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if v := authvalidation.ValidateUpdatePassword(&req); v.HasErrors() {
		return response.BadRequest(c, v.Error())
	}

	if err := h.authService.UpdatePassword(c.Context(), userID, &req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.OK(c, "password updated")
}

func (h *AuthHandler) UpdateUserPassword(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "user")
	if err != nil {
		return err
	}

	var req authdto.UpdateUserPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if v := authvalidation.ValidateUpdateUserPassword(&req); v.HasErrors() {
		return response.BadRequest(c, v.Error())
	}

	if err := h.authService.UpdateUserPassword(c.Context(), id, &req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.OK(c, "password updated")
}

func (h *AuthHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "user")
	if err != nil {
		return err
	}

	if err := h.authService.DeleteUser(c.Context(), id); err != nil {
		return response.InternalError(c, "failed to delete user")
	}

	return response.OK(c, "user deleted")
}
