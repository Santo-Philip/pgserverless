package validation

import (
	"github.com/nexbic/platform/internal/auth/dto"
	"github.com/nexbic/platform/pkg/validator"
)

func ValidateRegister(req *dto.RegisterRequest) *validator.Validator {
	v := validator.New()
	v.Required("email", req.Email)
	v.Required("password", req.Password)
	v.Email("email", req.Email)
	v.MinLength("password", req.Password, 8)
	v.MaxLength("password", req.Password, 128)
	v.MaxLength("name", req.Name, 255)
	return v
}

func ValidateLogin(req *dto.LoginRequest) *validator.Validator {
	v := validator.New()
	v.Required("email", req.Email)
	v.Required("password", req.Password)
	return v
}

func ValidateRefreshToken(req *dto.RefreshTokenRequest) *validator.Validator {
	v := validator.New()
	v.Required("refresh_token", req.RefreshToken)
	return v
}
