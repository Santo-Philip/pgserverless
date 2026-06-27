package validation

import (
	"github.com/nexbic/platform/internal/project/dto"
	"github.com/nexbic/platform/pkg/validator"
)

func ValidateCreate(req *dto.CreateProjectRequest) *validator.Validator {
	v := validator.New()
	v.Required("name", req.Name)
	v.Required("slug", req.Slug)
	v.Slug("slug", req.Slug)
	v.MinLength("name", req.Name, 3)
	v.MaxLength("name", req.Name, 255)
	v.MaxLength("description", req.Description, 1000)
	return v
}

func ValidateUpdate(req *dto.UpdateProjectRequest) *validator.Validator {
	v := validator.New()
	if req.Name != nil {
		v.MinLength("name", *req.Name, 3)
		v.MaxLength("name", *req.Name, 255)
	}
	return v
}
