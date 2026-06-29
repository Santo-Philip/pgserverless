package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/nexbic/platform/internal/storage/models"
	"github.com/nexbic/platform/internal/storage/repository"
)

type StorageService struct {
	repo *repository.StorageRepo
}

func NewStorageService(repo *repository.StorageRepo) *StorageService {
	return &StorageService{repo: repo}
}

// ── Providers ──────────────────────────────────────────

func (s *StorageService) CreateProvider(ctx context.Context, req *models.CreateProviderRequest, userID uuid.UUID) (*models.StorageProvider, error) {
	if req.ProviderType == "" {
		return nil, fmt.Errorf("provider_type is required")
	}

	p := &models.StorageProvider{
		Name:         req.Name,
		ProviderType: req.ProviderType,
		Config:       req.Config,
		IsDefault:    req.IsDefault,
		IsEnabled:    true,
		CreatedBy:    &userID,
	}

	if p.Config == nil {
		p.Config = map[string]any{}
	}

	if p.IsDefault {
		if err := s.repo.ClearDefaultFlag(ctx); err != nil {
			return nil, fmt.Errorf("clear default flag: %w", err)
		}
	}

	if err := s.repo.CreateProvider(ctx, p); err != nil {
		return nil, fmt.Errorf("create provider: %w", err)
	}

	return p, nil
}

func (s *StorageService) ListProviders(ctx context.Context) ([]models.StorageProvider, error) {
	return s.repo.ListProviders(ctx)
}

func (s *StorageService) GetProvider(ctx context.Context, id uuid.UUID) (*models.StorageProvider, error) {
	return s.repo.GetProvider(ctx, id)
}

func (s *StorageService) UpdateProvider(ctx context.Context, id uuid.UUID, req *models.UpdateProviderRequest) error {
	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Config != nil {
		updates["config"] = *req.Config
	}
	if req.IsDefault != nil {
		if *req.IsDefault {
			if err := s.repo.ClearDefaultFlag(ctx); err != nil {
				return fmt.Errorf("clear default flag: %w", err)
			}
		}
		updates["is_default"] = *req.IsDefault
	}
	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
	}
	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}
	return s.repo.UpdateProvider(ctx, id, updates)
}

func (s *StorageService) DeleteProvider(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteProvider(ctx, id)
}

// ── Buckets ────────────────────────────────────────────

func (s *StorageService) CreateBucket(ctx context.Context, req *models.CreateBucketRequest, userID uuid.UUID) (*models.StorageBucket, error) {
	providerID, err := uuid.Parse(req.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("invalid provider_id")
	}

	provider, err := s.repo.GetProvider(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("provider not found")
	}
	if provider == nil {
		return nil, fmt.Errorf("provider not found")
	}

	b := &models.StorageBucket{
		Name:       req.Name,
		ProviderID: providerID,
		Path:       req.Path,
		IsPublic:   req.IsPublic,
		CreatedBy:  &userID,
	}

	if err := s.repo.CreateBucket(ctx, b); err != nil {
		return nil, fmt.Errorf("create bucket: %w", err)
	}

	return b, nil
}

func (s *StorageService) ListBuckets(ctx context.Context, providerID uuid.UUID) ([]models.StorageBucket, error) {
	return s.repo.ListBuckets(ctx, providerID)
}

func (s *StorageService) GetBucket(ctx context.Context, id uuid.UUID) (*models.StorageBucket, error) {
	return s.repo.GetBucket(ctx, id)
}

func (s *StorageService) DeleteBucket(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteBucket(ctx, id)
}

// ── Files ──────────────────────────────────────────────

func (s *StorageService) UploadFile(ctx context.Context, bucketID uuid.UUID, name, mimeType string, data []byte, userID uuid.UUID) (*models.StorageFile, error) {
	bucket, err := s.repo.GetBucket(ctx, bucketID)
	if err != nil {
		return nil, fmt.Errorf("bucket not found")
	}
	if bucket == nil {
		return nil, fmt.Errorf("bucket not found")
	}

	provider, err := s.repo.GetProvider(ctx, bucket.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("provider not found")
	}
	if provider == nil {
		return nil, fmt.Errorf("provider not found")
	}

	basePath := s.getProviderBasePath(provider)
	filePath := filepath.Join(basePath, bucket.Path, name)

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, fmt.Errorf("create directory: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0640); err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}

	f := &models.StorageFile{
		BucketID:  bucketID,
		Name:      name,
		Path:      filePath,
		MimeType:  mimeType,
		SizeBytes: int64(len(data)),
		Metadata:  map[string]any{},
		CreatedBy: &userID,
	}

	if err := s.repo.CreateFile(ctx, f); err != nil {
		os.Remove(filePath)
		return nil, fmt.Errorf("create file record: %w", err)
	}

	return f, nil
}

func (s *StorageService) ListFiles(ctx context.Context, bucketID uuid.UUID, limit, offset int) ([]models.StorageFile, int, error) {
	return s.repo.ListFiles(ctx, bucketID, limit, offset)
}

func (s *StorageService) GetFile(ctx context.Context, id uuid.UUID) (*models.StorageFile, error) {
	return s.repo.GetFile(ctx, id)
}

func (s *StorageService) DeleteFile(ctx context.Context, id uuid.UUID) error {
	f, err := s.repo.GetFile(ctx, id)
	if err != nil {
		return err
	}
	if f == nil {
		return fmt.Errorf("file not found")
	}

	if f.Path != "" {
		os.Remove(f.Path)
	}

	return s.repo.DeleteFile(ctx, id)
}

func (s *StorageService) ReadFile(ctx context.Context, id uuid.UUID) ([]byte, string, error) {
	f, err := s.repo.GetFile(ctx, id)
	if err != nil {
		return nil, "", err
	}
	if f == nil {
		return nil, "", fmt.Errorf("file not found")
	}

	data, err := os.ReadFile(f.Path)
	if err != nil {
		return nil, "", fmt.Errorf("read file: %w", err)
	}

	return data, f.MimeType, nil
}

func (s *StorageService) getProviderBasePath(p *models.StorageProvider) string {
	if p.ProviderType == models.ProviderTypeLocal {
		if path, ok := p.Config["base_path"].(string); ok && path != "" {
			return path
		}
		return "/data/storage"
	}
	return "/data/storage"
}
