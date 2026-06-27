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

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req authdto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if v := authvalidation.ValidateRegister(&req); v.HasErrors() {
		return response.BadRequest(c, v.Error())
	}

	resp, err := h.authService.Register(c.Context(), &req)
	if err != nil {
		return response.Conflict(c, err.Error())
	}

	return response.Created(c, resp)
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
