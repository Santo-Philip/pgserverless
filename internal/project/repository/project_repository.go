package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexbic/platform/internal/project/models"
	"github.com/nexbic/platform/pkg/database"
	"github.com/nexbic/platform/pkg/helpers"
)

type ProjectRepository struct {
	db *database.DB
}

func NewProjectRepository(db *database.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(ctx context.Context, project *models.Project) error {
	project.ID = uuid.New()
	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()

	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO projects (id, name, slug, description, plan_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		project.ID, project.Name, project.Slug, project.Description, project.PlanID, project.Status,
		project.CreatedAt, project.UpdatedAt,
	)
	return err
}

func (r *ProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	p := &models.Project{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, name, slug, description, plan_id, status, created_at, updated_at
		FROM projects WHERE id = $1`, id).Scan(
		&p.ID, &p.Name, &p.Slug, &p.Description, &p.PlanID, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, helpers.ErrNoRowsAsNil(err)
	}
	return p, nil
}

func (r *ProjectRepository) GetBySlug(ctx context.Context, slug string) (*models.Project, error) {
	p := &models.Project{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, name, slug, description, plan_id, status, created_at, updated_at
		FROM projects WHERE slug = $1`, slug).Scan(
		&p.ID, &p.Name, &p.Slug, &p.Description, &p.PlanID, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, helpers.ErrNoRowsAsNil(err)
	}
	return p, nil
}

func (r *ProjectRepository) List(ctx context.Context, limit, offset int) ([]models.Project, int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM projects`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, name, slug, description, plan_id, status, created_at, updated_at
		FROM projects ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.PlanID, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		projects = append(projects, p)
	}

	return projects, total, nil
}

func (r *ProjectRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}

	query := "UPDATE projects SET updated_at = NOW()"
	args := []any{}
	argIdx := 1

	if name, ok := updates["name"]; ok {
		query += ", name = $" + itoa(argIdx)
		args = append(args, name)
		argIdx++
	}
	if desc, ok := updates["description"]; ok {
		query += ", description = $" + itoa(argIdx)
		args = append(args, desc)
		argIdx++
	}
	if planID, ok := updates["plan_id"]; ok {
		query += ", plan_id = $" + itoa(argIdx)
		args = append(args, planID)
		argIdx++
	}

	query += " WHERE id = $" + itoa(argIdx)
	args = append(args, id)

	_, err := r.db.Pool.Exec(ctx, query, args...)
	return err
}

func (r *ProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
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
