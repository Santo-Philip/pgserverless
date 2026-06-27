package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexbic/platform/internal/plan/models"
	"github.com/nexbic/platform/pkg/database"
	"github.com/nexbic/platform/pkg/helpers"
)

type PlanRepository struct {
	db *database.DB
}

func NewPlanRepository(db *database.DB) *PlanRepository {
	return &PlanRepository{db: db}
}

func (r *PlanRepository) Create(ctx context.Context, plan *models.Plan) error {
	plan.ID = uuid.New()
	plan.CreatedAt = time.Now()
	plan.UpdatedAt = time.Now()

	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO plans (id, name, slug, description, max_databases, max_storage_mb,
			max_connections, max_requests, max_api_keys, price, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		plan.ID, plan.Name, plan.Slug, plan.Description, plan.MaxDatabases,
		plan.MaxStorageMB, plan.MaxConnections, plan.MaxRequests, plan.MaxAPIKeys,
		plan.Price, plan.IsActive, plan.CreatedAt, plan.UpdatedAt,
	)
	return err
}

func (r *PlanRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Plan, error) {
	p := &models.Plan{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, name, slug, description, max_databases, max_storage_mb,
			max_connections, max_requests, max_api_keys, price, is_active, created_at, updated_at
		FROM plans WHERE id = $1`, id).Scan(
		&p.ID, &p.Name, &p.Slug, &p.Description, &p.MaxDatabases,
		&p.MaxStorageMB, &p.MaxConnections, &p.MaxRequests, &p.MaxAPIKeys,
		&p.Price, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, helpers.ErrNoRowsAsNil(err)
	}
	return p, nil
}

func (r *PlanRepository) GetBySlug(ctx context.Context, slug string) (*models.Plan, error) {
	p := &models.Plan{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, name, slug, description, max_databases, max_storage_mb,
			max_connections, max_requests, max_api_keys, price, is_active, created_at, updated_at
		FROM plans WHERE slug = $1`, slug).Scan(
		&p.ID, &p.Name, &p.Slug, &p.Description, &p.MaxDatabases,
		&p.MaxStorageMB, &p.MaxConnections, &p.MaxRequests, &p.MaxAPIKeys,
		&p.Price, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, helpers.ErrNoRowsAsNil(err)
	}
	return p, nil
}

func (r *PlanRepository) List(ctx context.Context) ([]models.Plan, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, name, slug, description, max_databases, max_storage_mb,
			max_connections, max_requests, max_api_keys, price, is_active, created_at, updated_at
		FROM plans ORDER BY price ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []models.Plan
	for rows.Next() {
		var p models.Plan
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.MaxDatabases,
			&p.MaxStorageMB, &p.MaxConnections, &p.MaxRequests, &p.MaxAPIKeys,
			&p.Price, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}

	if plans == nil {
		plans = []models.Plan{}
	}

	return plans, nil
}

func (r *PlanRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}

	query := "UPDATE plans SET updated_at = NOW()"
	args := []any{}
	argIdx := 1

	for k, v := range updates {
		query += ", " + k + " = $" + itoa(argIdx)
		args = append(args, v)
		argIdx++
	}

	query += " WHERE id = $" + itoa(argIdx)
	args = append(args, id)

	_, err := r.db.Pool.Exec(ctx, query, args...)
	return err
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	s := ""
	for i > 0 {
		s = string(rune('0'+i%10)) + s
		i /= 10
	}
	return s
}
