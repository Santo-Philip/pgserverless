package service

import (
	"context"
	"fmt"

	"github.com/nexbic/platform/management-api/repository"
	"github.com/nexbic/platform/shared/models"
	"github.com/nexbic/platform/shared/utils"

	"github.com/google/uuid"
)

type APIKeyService struct {
	keyRepo *repository.APIKeyRepository
}

func NewAPIKeyService(keyRepo *repository.APIKeyRepository) *APIKeyService {
	return &APIKeyService{keyRepo: keyRepo}
}

func (s *APIKeyService) CreateKey(ctx context.Context, appID uuid.UUID, userID uuid.UUID, req models.CreateAPIKeyRequest) (*models.APIKeyResponse, error) {
	v := utils.NewValidator()
	v.Required("name", req.Name)
	if req.KeyType == "" {
		req.KeyType = models.KeyTypeSecret
	}
	v.OneOf("key_type", string(req.KeyType), string(models.KeyTypePublishable), string(models.KeyTypeSecret), string(models.KeyTypeService), string(models.KeyTypeAdmin))
	if v.HasErrors() {
		return nil, fmt.Errorf("validation: %s", v.Error())
	}

	rawKey, prefix, err := utils.GenerateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}

	key := &models.APIKey{
		AppID:     appID,
		UserID:    &userID,
		Name:      req.Name,
		KeyType:   req.KeyType,
		KeyHash:   s.keyRepo.HashKey(rawKey),
		KeyPrefix: prefix,
		Scopes:    req.Scopes,
		RateLimit: req.RateLimit,
		ExpiresAt: req.ExpiresAt,
		IsActive:  true,
	}

	if key.Scopes == nil {
		key.Scopes = []string{}
	}
	if key.RateLimit <= 0 {
		key.RateLimit = 1000
	}

	if err := s.keyRepo.Create(ctx, key); err != nil {
		return nil, fmt.Errorf("create key: %w", err)
	}

	return &models.APIKeyResponse{
		ID:        key.ID,
		Name:      key.Name,
		KeyType:   key.KeyType,
		KeyPrefix: prefix,
		RawKey:    rawKey,
		Scopes:    key.Scopes,
		CreatedAt: key.CreatedAt,
	}, nil
}

func (s *APIKeyService) ListByApp(ctx context.Context, appID uuid.UUID) ([]models.APIKey, error) {
	return s.keyRepo.ListByApp(ctx, appID)
}

func (s *APIKeyService) Deactivate(ctx context.Context, keyID uuid.UUID) error {
	return s.keyRepo.Deactivate(ctx, keyID)
}
