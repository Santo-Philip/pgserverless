package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nexbic/platform/internal/database/storage/models"
	"github.com/nexbic/platform/pkg/database"
	"github.com/nexbic/platform/pkg/helpers"
)

type StorageRepo struct {
	db *database.DB
}

func NewStorageRepo(db *database.DB) *StorageRepo {
	return &StorageRepo{db: db}
}

// ── Providers ─────────────────────────────────────────

func (r *StorageRepo) CreateProvider(ctx context.Context, p *models.StorageProvider) error {
	p.ID = uuid.New()
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()

	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO storage_providers (id, name, provider_type, config, is_default, is_enabled, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		p.ID, p.Name, string(p.ProviderType), p.Config, p.IsDefault, p.IsEnabled, p.CreatedBy, p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *StorageRepo) ListProviders(ctx context.Context) ([]models.StorageProvider, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, name, provider_type, config, is_default, is_enabled, created_by, created_at, updated_at
		FROM storage_providers ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []models.StorageProvider
	for rows.Next() {
		var p models.StorageProvider
		var pType string
		if err := rows.Scan(&p.ID, &p.Name, &pType, &p.Config, &p.IsDefault, &p.IsEnabled, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.ProviderType = models.ProviderType(pType)
		providers = append(providers, p)
	}
	if providers == nil {
		providers = []models.StorageProvider{}
	}
	return providers, nil
}

func (r *StorageRepo) GetProvider(ctx context.Context, id uuid.UUID) (*models.StorageProvider, error) {
	p := &models.StorageProvider{}
	var pType string
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, name, provider_type, config, is_default, is_enabled, created_by, created_at, updated_at
		FROM storage_providers WHERE id = $1`, id).Scan(
		&p.ID, &p.Name, &pType, &p.Config, &p.IsDefault, &p.IsEnabled, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, helpers.ErrNoRowsAsNil(err)
	}
	p.ProviderType = models.ProviderType(pType)
	return p, nil
}

func (r *StorageRepo) GetDefaultProvider(ctx context.Context) (*models.StorageProvider, error) {
	p := &models.StorageProvider{}
	var pType string
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, name, provider_type, config, is_default, is_enabled, created_by, created_at, updated_at
		FROM storage_providers WHERE is_default = TRUE AND is_enabled = TRUE LIMIT 1`).Scan(
		&p.ID, &p.Name, &pType, &p.Config, &p.IsDefault, &p.IsEnabled, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, helpers.ErrNoRowsAsNil(err)
	}
	p.ProviderType = models.ProviderType(pType)
	return p, nil
}

func (r *StorageRepo) UpdateProvider(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	// Build SET clause dynamically from the updates map
	query := "UPDATE storage_providers SET "
	args := []any{}
	i := 1
	for col, val := range updates {
		if i > 1 {
			query += ", "
		}
		query += col + " = $" + fmt.Sprintf("%d", i)
		args = append(args, val)
		i++
	}
	query += ", updated_at = $" + fmt.Sprintf("%d", i)
	args = append(args, time.Now())
	i++
	query += " WHERE id = $" + fmt.Sprintf("%d", i)
	args = append(args, id)

	_, err := r.db.Pool.Exec(ctx, query, args...)
	return err
}

func (r *StorageRepo) DeleteProvider(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM storage_providers WHERE id = $1`, id)
	return err
}

func (r *StorageRepo) ClearDefaultFlag(ctx context.Context) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE storage_providers SET is_default = FALSE, updated_at = NOW()`)
	return err
}

// ── Buckets ────────────────────────────────────────────

func (r *StorageRepo) CreateBucket(ctx context.Context, b *models.StorageBucket) error {
	b.ID = uuid.New()
	b.CreatedAt = time.Now()
	b.UpdatedAt = time.Now()

	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO storage_buckets (id, name, provider_id, path, is_public, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		b.ID, b.Name, b.ProviderID, b.Path, b.IsPublic, b.CreatedBy, b.CreatedAt, b.UpdatedAt)
	return err
}

func (r *StorageRepo) ListBuckets(ctx context.Context, providerID uuid.UUID) ([]models.StorageBucket, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, name, provider_id, path, is_public, created_by, created_at, updated_at
		FROM storage_buckets WHERE provider_id = $1 ORDER BY name`, providerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var buckets []models.StorageBucket
	for rows.Next() {
		var b models.StorageBucket
		if err := rows.Scan(&b.ID, &b.Name, &b.ProviderID, &b.Path, &b.IsPublic, &b.CreatedBy, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		buckets = append(buckets, b)
	}
	if buckets == nil {
		buckets = []models.StorageBucket{}
	}
	return buckets, nil
}

func (r *StorageRepo) GetBucket(ctx context.Context, id uuid.UUID) (*models.StorageBucket, error) {
	b := &models.StorageBucket{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, name, provider_id, path, is_public, created_by, created_at, updated_at
		FROM storage_buckets WHERE id = $1`, id).Scan(
		&b.ID, &b.Name, &b.ProviderID, &b.Path, &b.IsPublic, &b.CreatedBy, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return nil, helpers.ErrNoRowsAsNil(err)
	}
	return b, nil
}

func (r *StorageRepo) DeleteBucket(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM storage_buckets WHERE id = $1`, id)
	return err
}

// ── Files ──────────────────────────────────────────────

func (r *StorageRepo) CreateFile(ctx context.Context, f *models.StorageFile) error {
	f.ID = uuid.New()
	f.CreatedAt = time.Now()
	f.UpdatedAt = time.Now()

	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO storage_files (id, bucket_id, name, path, mime_type, size_bytes, md5_hash, metadata, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		f.ID, f.BucketID, f.Name, f.Path, f.MimeType, f.SizeBytes, f.MD5Hash, f.Metadata, f.CreatedBy, f.CreatedAt, f.UpdatedAt)
	return err
}

func (r *StorageRepo) ListFiles(ctx context.Context, bucketID uuid.UUID, limit, offset int) ([]models.StorageFile, int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM storage_files WHERE bucket_id = $1`, bucketID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, bucket_id, name, path, mime_type, size_bytes, COALESCE(md5_hash, ''), metadata, created_by, created_at, updated_at
		FROM storage_files WHERE bucket_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		bucketID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var files []models.StorageFile
	for rows.Next() {
		var f models.StorageFile
		if err := rows.Scan(&f.ID, &f.BucketID, &f.Name, &f.Path, &f.MimeType, &f.SizeBytes, &f.MD5Hash, &f.Metadata, &f.CreatedBy, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, 0, err
		}
		files = append(files, f)
	}
	if files == nil {
		files = []models.StorageFile{}
	}
	return files, total, nil
}

func (r *StorageRepo) GetFile(ctx context.Context, id uuid.UUID) (*models.StorageFile, error) {
	f := &models.StorageFile{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, bucket_id, name, path, mime_type, size_bytes, COALESCE(md5_hash, ''), metadata, created_by, created_at, updated_at
		FROM storage_files WHERE id = $1`, id).Scan(
		&f.ID, &f.BucketID, &f.Name, &f.Path, &f.MimeType, &f.SizeBytes, &f.MD5Hash, &f.Metadata, &f.CreatedBy, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, helpers.ErrNoRowsAsNil(err)
	}
	return f, nil
}

func (r *StorageRepo) DeleteFile(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM storage_files WHERE id = $1`, id)
	return err
}


