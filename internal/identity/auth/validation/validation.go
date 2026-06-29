package validation

import (
	"github.com/nexbic/platform/internal/identity/auth/dto"
	"github.com/nexbic/platform/pkg/validator"
)

func ValidateLogin(req *dto.LoginRequest) *validator.Validator {
	v := validator.New()
	v.Required("email", req.Email)
	v.Required("password", req.Password)
	v.Email("email", req.Email)
	return v
}

func ValidateRefreshToken(req *dto.RefreshTokenRequest) *validator.Validator {
	v := validator.New()
	v.Required("refresh_token", req.RefreshToken)
	return v
}

func ValidateCreateAPIKey(req *dto.CreateAPIKeyRequest) *validator.Validator {
	v := validator.New()
	v.Required("name", req.Name)
	v.MaxLength("name", req.Name, 255)
	return v
}

func ValidateCreateUser(req *dto.CreateUserRequest) *validator.Validator {
	v := validator.New()
	v.Required("email", req.Email)
	v.Required("password", req.Password)
	v.Email("email", req.Email)
	v.MinLength("password", req.Password, 8)
	v.MaxLength("password", req.Password, 128)
	v.MaxLength("name", req.Name, 255)
	v.Required("role", req.Role)
	v.OneOf("role", req.Role, "super_admin", "dba", "developer", "read_only")
	return v
}

func ValidateUpdateUser(req *dto.UpdateUserRequest) *validator.Validator {
	v := validator.New()
	if req.Role != "" {
		v.OneOf("role", req.Role, "super_admin", "dba", "developer", "read_only")
	}
	return v
}

func ValidateUpdatePassword(req *dto.UpdatePasswordRequest) *validator.Validator {
	v := validator.New()
	v.Required("current_password", req.CurrentPassword)
	v.Required("new_password", req.NewPassword)
	v.MinLength("new_password", req.NewPassword, 8)
	return v
}

func ValidateUpdateUserPassword(req *dto.UpdateUserPasswordRequest) *validator.Validator {
	v := validator.New()
	v.Required("new_password", req.NewPassword)
	v.MinLength("new_password", req.NewPassword, 8)
	return v
}
