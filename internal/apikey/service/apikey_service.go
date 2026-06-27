package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexbic/platform/internal/apikey/dto"
	"github.com/nexbic/platform/internal/apikey/models"
	"github.com/nexbic/platform/internal/apikey/repository"
	"github.com/nexbic/platform/pkg/password"
)

type APIKeyService struct {
	repo *repository.APIKeyRepository
}

func NewAPIKeyService(repo *repository.APIKeyRepository) *APIKeyService {
	return &APIKeyService{repo: repo}
}

func (s *APIKeyService) CreateKey(ctx context.Context, req *dto.CreateKeyRequest, createdBy uuid.UUID) (*dto.KeyResponse, error) {
	if req.RateLimit <= 0 {
		req.RateLimit = 1000
	}

	rawKey, keyHash, keyPrefix, err := password.GenerateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}

	var projectID *uuid.UUID
	if req.ProjectID != "" {
		if pid, err := uuid.Parse(req.ProjectID); err == nil {
			projectID = &pid
		}
	}

	if req.KeyType == models.KeyTypeProject && projectID == nil {
		return nil, fmt.Errorf("project_id is required for project keys")
	}

	key := &models.APIKey{
		Name:       req.Name,
		KeyType:    req.KeyType,
		KeyHash:    keyHash,
		KeyPrefix:  keyPrefix,
		Scopes:     req.Scopes,
		ProjectID:  projectID,
		RateLimit:  req.RateLimit,
		AllowedIPs: req.IPs,
		Origins:    req.Origins,
		ExpiresAt:  req.ExpiresAt,
		IsActive:   true,
		CreatedBy:  createdBy,
	}

	if err := s.repo.Create(ctx, key); err != nil {
		return nil, fmt.Errorf("create key: %w", err)
	}

	return &dto.KeyResponse{
		ID:        key.ID,
		Name:      key.Name,
		KeyType:   key.KeyType,
		KeyPrefix: key.KeyPrefix,
		RawKey:    rawKey,
		Scopes:    key.Scopes,
		CreatedAt: key.CreatedAt,
	}, nil
}

func (s *APIKeyService) List(ctx context.Context, limit, offset int) ([]models.APIKey, int, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *APIKeyService) ListByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]models.APIKey, int, error) {
	return s.repo.ListByProject(ctx, projectID, limit, offset)
}

func (s *APIKeyService) Revoke(ctx context.Context, id uuid.UUID) error {
	key, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if key == nil {
		return fmt.Errorf("key not found")
	}

	return s.repo.Revoke(ctx, id)
}
