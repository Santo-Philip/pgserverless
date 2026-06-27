package validation

import (
	"github.com/nexbic/platform/internal/plan/dto"
	"github.com/nexbic/platform/pkg/validator"
)

func ValidateCreate(req *dto.CreatePlanRequest) *validator.Validator {
	v := validator.New()
	v.Required("name", req.Name)
	v.Required("slug", req.Slug)
	v.Slug("slug", req.Slug)
	v.MinLength("name", req.Name, 3)
	v.MaxLength("name", req.Name, 255)
	v.MinInt("max_databases", req.MaxDatabases, 1)
	v.MinInt("max_connections", req.MaxConnections, 1)
	v.MinInt("max_requests", req.MaxRequests, 1)
	v.MinInt("max_api_keys", req.MaxAPIKeys, 1)
	return v
}
