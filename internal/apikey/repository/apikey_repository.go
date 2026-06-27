package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexbic/platform/internal/apikey/models"
	"github.com/nexbic/platform/pkg/database"
	"github.com/nexbic/platform/pkg/helpers"
	"github.com/nexbic/platform/pkg/password"
)

type APIKeyRepository struct {
	db *database.DB
}

func NewAPIKeyRepository(db *database.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) Create(ctx context.Context, key *models.APIKey) error {
	key.ID = uuid.New()
	key.CreatedAt = time.Now()
	key.UpdatedAt = time.Now()

	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO api_keys (id, name, key_type, key_hash, key_prefix, scopes, project_id,
			rate_limit, allowed_ips, origins, expires_at, is_active, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		key.ID, key.Name, string(key.KeyType), key.KeyHash, key.KeyPrefix, key.Scopes,
		key.ProjectID, key.RateLimit, key.AllowedIPs, key.Origins, key.ExpiresAt,
		key.IsActive, key.CreatedBy, key.CreatedAt, key.UpdatedAt,
	)
	return err
}

func (r *APIKeyRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.APIKey, error) {
	key := &models.APIKey{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, name, key_type, key_hash, key_prefix, scopes, project_id,
			rate_limit, allowed_ips, origins, expires_at, is_active, revoked_at, created_by, created_at, updated_at
		FROM api_keys WHERE id = $1`, id).Scan(
		&key.ID, &key.Name, &key.KeyType, &key.KeyHash, &key.KeyPrefix, &key.Scopes,
		&key.ProjectID, &key.RateLimit, &key.AllowedIPs, &key.Origins, &key.ExpiresAt,
		&key.IsActive, &key.RevokedAt, &key.CreatedBy, &key.CreatedAt, &key.UpdatedAt,
	)
	if err != nil {
		return nil, helpers.ErrNoRowsAsNil(err)
	}
	return key, nil
}

func (r *APIKeyRepository) List(ctx context.Context, limit, offset int) ([]models.APIKey, int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM api_keys`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, name, key_type, key_hash, key_prefix, scopes, project_id,
			rate_limit, allowed_ips, origins, expires_at, is_active, revoked_at, created_by, created_at, updated_at
		FROM api_keys ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var keys []models.APIKey
	for rows.Next() {
		var k models.APIKey
		if err := rows.Scan(&k.ID, &k.Name, &k.KeyType, &k.KeyHash, &k.KeyPrefix, &k.Scopes,
			&k.ProjectID, &k.RateLimit, &k.AllowedIPs, &k.Origins, &k.ExpiresAt,
			&k.IsActive, &k.RevokedAt, &k.CreatedBy, &k.CreatedAt, &k.UpdatedAt); err != nil {
			return nil, 0, err
		}
		keys = append(keys, k)
	}

	return keys, total, nil
}

func (r *APIKeyRepository) ListByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]models.APIKey, int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM api_keys WHERE project_id = $1`, projectID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, name, key_type, key_hash, key_prefix, scopes, project_id,
			rate_limit, allowed_ips, origins, expires_at, is_active, revoked_at, created_by, created_at, updated_at
		FROM api_keys WHERE project_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		projectID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var keys []models.APIKey
	for rows.Next() {
		var k models.APIKey
		if err := rows.Scan(&k.ID, &k.Name, &k.KeyType, &k.KeyHash, &k.KeyPrefix, &k.Scopes,
			&k.ProjectID, &k.RateLimit, &k.AllowedIPs, &k.Origins, &k.ExpiresAt,
			&k.IsActive, &k.RevokedAt, &k.CreatedBy, &k.CreatedAt, &k.UpdatedAt); err != nil {
			return nil, 0, err
		}
		keys = append(keys, k)
	}

	return keys, total, nil
}

func (r *APIKeyRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE api_keys SET is_active = false, revoked_at = $1, updated_at = $2 WHERE id = $3`,
		now, now, id)
	return err
}

func (r *APIKeyRepository) HashKey(rawKey string) string {
	return password.HashKey(rawKey)
}
