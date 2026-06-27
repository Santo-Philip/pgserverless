package validation

import (
	"github.com/nexbic/platform/internal/apikey/dto"
	"github.com/nexbic/platform/internal/apikey/models"
	"github.com/nexbic/platform/pkg/validator"
)

func ValidateCreate(req *dto.CreateKeyRequest) *validator.Validator {
	v := validator.New()
	v.Required("name", req.Name)
	v.MaxLength("name", req.Name, 255)
	v.OneOf("key_type", string(req.KeyType), string(models.KeyTypeSystem), string(models.KeyTypeService), string(models.KeyTypeProject))
	return v
}
