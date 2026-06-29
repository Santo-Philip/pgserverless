package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	authdto "github.com/nexbic/platform/internal/identity/auth/dto"
	authservice "github.com/nexbic/platform/internal/identity/auth/service"
	authvalidation "github.com/nexbic/platform/internal/identity/auth/validation"
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

	resp, err := h.authService.Login(c.Context(), &req, c.IP(), c.Get("User-Agent"))
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

func (h *AuthHandler) ListDevices(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	if userID == uuid.Nil {
		return response.Unauthorized(c, "not authenticated")
	}
	devices, err := h.authService.ListDevices(c.Context(), userID)
	if err != nil {
		return response.InternalError(c, "failed to list devices")
	}
	return response.OK(c, devices)
}

func (h *AuthHandler) DeleteDevice(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	if userID == uuid.Nil {
		return response.Unauthorized(c, "not authenticated")
	}
	id, err := helpers.ParseUUIDParam(c, "id", "device")
	if err != nil {
		return err
	}
	if err := h.authService.DeleteDevice(c.Context(), id, userID); err != nil {
		return response.InternalError(c, "failed to delete device")
	}
	return response.NoContent(c)
}

func (h *AuthHandler) ListSecurityEvents(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	if userID == uuid.Nil {
		return response.Unauthorized(c, "not authenticated")
	}
	p := helpers.ParsePagination(c)
	events, total, err := h.authService.ListSecurityEvents(c.Context(), userID, p.Limit, p.Offset)
	if err != nil {
		return response.InternalError(c, "failed to list security events")
	}
	return response.Paginated(c, events, total, p.Limit, p.Offset)
}

func (h *AuthHandler) CreateAPIKey(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	if userID == uuid.Nil {
		return response.Unauthorized(c, "not authenticated")
	}
	var req authdto.CreateAPIKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if v := authvalidation.ValidateCreateAPIKey(&req); v.HasErrors() {
		return response.BadRequest(c, v.Error())
	}
	resp, err := h.authService.CreateAPIKey(c.Context(), userID, &req)
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Created(c, resp)
}

func (h *AuthHandler) ListAPIKeys(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	if userID == uuid.Nil {
		return response.Unauthorized(c, "not authenticated")
	}
	keys, err := h.authService.ListAPIKeys(c.Context(), userID)
	if err != nil {
		return response.InternalError(c, "failed to list api keys")
	}
	return response.OK(c, keys)
}

func (h *AuthHandler) RevokeAPIKey(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	if userID == uuid.Nil {
		return response.Unauthorized(c, "not authenticated")
	}
	id, err := helpers.ParseUUIDParam(c, "id", "api_key")
	if err != nil {
		return err
	}
	if err := h.authService.RevokeAPIKey(c.Context(), id, userID); err != nil {
		return response.InternalError(c, "failed to revoke api key")
	}
	return response.NoContent(c)
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

func (h *AuthHandler) OAuthGoogle(c *fiber.Ctx) error {
	url := h.authService.GetOAuthURL("google")
	return response.OK(c, authdto.OAuthURLResponse{URL: url})
}

func (h *AuthHandler) OAuthGitHub(c *fiber.Ctx) error {
	url := h.authService.GetOAuthURL("github")
	return response.OK(c, authdto.OAuthURLResponse{URL: url})
}

func (h *AuthHandler) OAuthCallback(c *fiber.Ctx) error {
	var req authdto.OAuthCallbackRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	resp, err := h.authService.HandleOAuthCallback(c.Context(), req.Provider, req.Code)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}
	return response.OK(c, resp)
}
