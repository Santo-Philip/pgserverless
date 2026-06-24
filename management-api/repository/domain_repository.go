package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nexbic/platform/shared/database"
	"github.com/nexbic/platform/shared/models"
)

type DomainRepository struct {
	db *database.DB
}

func NewDomainRepository(db *database.DB) *DomainRepository {
	return &DomainRepository{db: db}
}

func (r *DomainRepository) Create(ctx context.Context, domain *models.Domain) error {
	domain.ID = models.NewID()
	domain.CreatedAt = models.Now()
	domain.UpdatedAt = models.Now()

	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO domains (id, app_id, domain, status, verified, verification_token, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, domain.ID, domain.AppID, domain.Domain, domain.Status, domain.Verified, domain.VerificationToken, domain.CreatedAt, domain.UpdatedAt)

	return err
}

func (r *DomainRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Domain, error) {
	domain := &models.Domain{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, app_id, domain, status, verified, verification_token, verified_at, created_at, updated_at
		FROM domains WHERE id = $1
	`, id).Scan(&domain.ID, &domain.AppID, &domain.Domain, &domain.Status, &domain.Verified, &domain.VerificationToken, &domain.VerifiedAt, &domain.CreatedAt, &domain.UpdatedAt)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return domain, err
}

func (r *DomainRepository) ListByApp(ctx context.Context, appID uuid.UUID) ([]models.Domain, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, app_id, domain, status, verified, verification_token, verified_at, created_at, updated_at
		FROM domains WHERE app_id = $1 ORDER BY created_at DESC
	`, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []models.Domain
	for rows.Next() {
		var d models.Domain
		if err := rows.Scan(&d.ID, &d.AppID, &d.Domain, &d.Status, &d.Verified, &d.VerificationToken, &d.VerifiedAt, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}

	return domains, nil
}

func (r *DomainRepository) Verify(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE domains SET verified = TRUE, status = 'active', verified_at = $1, updated_at = $1 WHERE id = $2
	`, now, id)
	return err
}

func (r *DomainRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM domains WHERE id = $1`, id)
	return err
}
