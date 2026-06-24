package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nexbic/platform/shared/database"
	"github.com/nexbic/platform/shared/models"
)

type APIKeyRepository struct {
	db *database.DB
}

func NewAPIKeyRepository(db *database.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) Create(ctx context.Context, key *models.APIKey) error {
	key.ID = models.NewID()
	key.CreatedAt = models.Now()
	key.UpdatedAt = models.Now()

	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO api_keys (id, app_id, user_id, name, key_type, key_hash, key_prefix, scopes, rate_limit, allowed_ips, expires_at, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, key.ID, key.AppID, key.UserID, key.Name, key.KeyType, key.KeyHash, key.KeyPrefix, key.Scopes, key.RateLimit, key.AllowedIPs, key.ExpiresAt, key.IsActive, key.CreatedAt, key.UpdatedAt)

	return err
}

func (r *APIKeyRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.APIKey, error) {
	key := &models.APIKey{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, app_id, user_id, name, key_type, key_hash, key_prefix, scopes, rate_limit, allowed_ips, last_used_at, expires_at, is_active, created_at, updated_at
		FROM api_keys WHERE id = $1
	`, id).Scan(&key.ID, &key.AppID, &key.UserID, &key.Name, &key.KeyType, &key.KeyHash, &key.KeyPrefix, &key.Scopes, &key.RateLimit, &key.AllowedIPs, &key.LastUsedAt, &key.ExpiresAt, &key.IsActive, &key.CreatedAt, &key.UpdatedAt)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return key, err
}

func (r *APIKeyRepository) ListByApp(ctx context.Context, appID uuid.UUID) ([]models.APIKey, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, app_id, user_id, name, key_type, key_hash, key_prefix, scopes, rate_limit, allowed_ips, last_used_at, expires_at, is_active, created_at, updated_at
		FROM api_keys WHERE app_id = $1 ORDER BY created_at DESC
	`, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []models.APIKey
	for rows.Next() {
		var key models.APIKey
		if err := rows.Scan(&key.ID, &key.AppID, &key.UserID, &key.Name, &key.KeyType, &key.KeyHash, &key.KeyPrefix, &key.Scopes, &key.RateLimit, &key.AllowedIPs, &key.LastUsedAt, &key.ExpiresAt, &key.IsActive, &key.CreatedAt, &key.UpdatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}

	return keys, nil
}

func (r *APIKeyRepository) Deactivate(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE api_keys SET is_active = FALSE, updated_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *APIKeyRepository) HashKey(rawKey string) string {
	return fmt.Sprintf("hashed_%s", rawKey[:16])
}
