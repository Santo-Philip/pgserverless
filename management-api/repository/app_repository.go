package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nexbic/platform/shared/database"
	"github.com/nexbic/platform/shared/models"
)

type AppRepository struct {
	db *database.DB
}

func NewAppRepository(db *database.DB) *AppRepository {
	return &AppRepository{db: db}
}

func (r *AppRepository) Create(ctx context.Context, app *models.App) error {
	app.ID = models.NewID()
	app.CreatedAt = models.Now()
	app.UpdatedAt = models.Now()

	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO apps (id, org_id, owner_id, name, slug, schema_name, description, status, region, visibility, settings, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, app.ID, app.OrgID, app.OwnerID, app.Name, app.Slug, app.SchemaName, app.Description, app.Status, app.Region, app.Visibility, app.Settings, app.CreatedAt, app.UpdatedAt)

	return err
}

func (r *AppRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.App, error) {
	app := &models.App{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, org_id, owner_id, name, slug, schema_name, description, status, region, visibility, settings, created_at, updated_at, deleted_at
		FROM apps WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(&app.ID, &app.OrgID, &app.OwnerID, &app.Name, &app.Slug, &app.SchemaName, &app.Description, &app.Status, &app.Region, &app.Visibility, &app.Settings, &app.CreatedAt, &app.UpdatedAt, &app.DeletedAt)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return app, err
}

func (r *AppRepository) GetBySlug(ctx context.Context, slug string) (*models.App, error) {
	app := &models.App{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, org_id, owner_id, name, slug, schema_name, description, status, region, visibility, settings, created_at, updated_at, deleted_at
		FROM apps WHERE slug = $1 AND deleted_at IS NULL
	`, slug).Scan(&app.ID, &app.OrgID, &app.OwnerID, &app.Name, &app.Slug, &app.SchemaName, &app.Description, &app.Status, &app.Region, &app.Visibility, &app.Settings, &app.CreatedAt, &app.UpdatedAt, &app.DeletedAt)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return app, err
}

func (r *AppRepository) List(ctx context.Context, orgID *uuid.UUID, limit, offset int) ([]models.App, int, error) {
	var total int
	args := []interface{}{}
	where := "WHERE deleted_at IS NULL"
	argIdx := 1

	if orgID != nil {
		where += fmt.Sprintf(" AND org_id = $%d", argIdx)
		args = append(args, *orgID)
		argIdx++
	}

	err := r.db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM apps "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	rows, err := r.db.Pool.Query(ctx, fmt.Sprintf(`
		SELECT id, org_id, owner_id, name, slug, schema_name, description, status, region, visibility, settings, created_at, updated_at, deleted_at
		FROM apps %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var apps []models.App
	for rows.Next() {
		var app models.App
		if err := rows.Scan(&app.ID, &app.OrgID, &app.OwnerID, &app.Name, &app.Slug, &app.SchemaName, &app.Description, &app.Status, &app.Region, &app.Visibility, &app.Settings, &app.CreatedAt, &app.UpdatedAt, &app.DeletedAt); err != nil {
			return nil, 0, err
		}
		apps = append(apps, app)
	}

	return apps, total, nil
}

func (r *AppRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE apps SET status = $1, deleted_at = $2, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL
	`, models.AppStatusDeleted, now, id)
	return err
}
