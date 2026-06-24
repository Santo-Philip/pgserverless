package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nexbic/platform/shared/database"
	"github.com/nexbic/platform/shared/models"
)

type AppResolver struct {
	db         *database.DB
	cache      map[string]*models.App
	mu         sync.RWMutex
	cacheTTL   time.Duration
	lastFetch  time.Time
}

func NewAppResolver(db *database.DB) *AppResolver {
	return &AppResolver{
		db:        db,
		cache:     make(map[string]*models.App),
		cacheTTL: 30 * time.Second,
	}
}

func (r *AppResolver) ResolveBySlug(ctx context.Context, slug string) (*models.App, error) {
	r.mu.RLock()
	app, exists := r.cache[slug]
	cacheValid := exists && time.Since(r.lastFetch) < r.cacheTTL
	r.mu.RUnlock()

	if cacheValid {
		return app, nil
	}

	return r.fetchFromDB(ctx, slug)
}

func (r *AppResolver) fetchFromDB(ctx context.Context, slug string) (*models.App, error) {
	app := &models.App{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, org_id, owner_id, name, slug, schema_name, description, status, region, visibility, settings, created_at, updated_at
		FROM apps WHERE slug = $1 AND status = $2 AND deleted_at IS NULL
	`, slug, models.AppStatusActive).Scan(
		&app.ID, &app.OrgID, &app.OwnerID, &app.Name, &app.Slug, &app.SchemaName,
		&app.Description, &app.Status, &app.Region, &app.Visibility, &app.Settings, &app.CreatedAt, &app.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("app not found: %s", slug)
	}

	r.mu.Lock()
	r.cache[slug] = app
	r.lastFetch = time.Now()
	r.mu.Unlock()

	return app, nil
}

func (r *AppResolver) Invalidate(slug string) {
	r.mu.Lock()
	delete(r.cache, slug)
	r.mu.Unlock()
}

func (r *AppResolver) InvalidateAll() {
	r.mu.Lock()
	r.cache = make(map[string]*models.App)
	r.lastFetch = time.Time{}
	r.mu.Unlock()
}
