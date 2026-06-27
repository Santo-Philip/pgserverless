package validation

import (
	"github.com/nexbic/platform/internal/database/dto"
	"github.com/nexbic/platform/pkg/validator"
)

func ValidateCreate(req *dto.CreateDatabaseRequest) *validator.Validator {
	v := validator.New()
	v.Required("name", req.Name)
	v.Required("project_id", req.ProjectID)
	v.MinLength("name", req.Name, 3)
	v.MaxLength("name", req.Name, 63)
	return v
}

func ValidateCreateTable(req *dto.CreateTableRequest) *validator.Validator {
	v := validator.New()
	v.Required("name", req.Name)
	v.MinLength("name", req.Name, 1)
	v.MaxLength("name", req.Name, 63)
	if len(req.Columns) == 0 {
		v.Errors["columns"] = "at least one column is required"
	}
	return v
}

func ValidateAddColumn(req *dto.AddColumnRequest) *validator.Validator {
	v := validator.New()
	v.Required("name", req.Name)
	v.Required("type", req.Type)
	return v
}

func ValidateRunSQL(req *dto.RunSQLRequest) *validator.Validator {
	v := validator.New()
	v.Required("query", req.Query)
	return v
}
