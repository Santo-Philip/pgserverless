package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nexbic/platform/internal/project/repository"
	"github.com/nexbic/platform/internal/quota/models"
	"github.com/nexbic/platform/pkg/database"
)

type QuotaService struct {
	db            *database.DB
	projectRepo   *repository.ProjectRepository
}

func NewQuotaService(db *database.DB, projectRepo *repository.ProjectRepository) *QuotaService {
	return &QuotaService{
		db:          db,
		projectRepo: projectRepo,
	}
}

func (s *QuotaService) GetQuota(ctx context.Context, projectID uuid.UUID) (*models.Quota, *models.QuotaLimit, error) {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, nil, err
	}
	if project == nil {
		return nil, nil, fmt.Errorf("project not found")
	}

	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	periodEnd := periodStart.AddDate(0, 1, 0)

	var dbCount int
	s.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM databases WHERE project_id = $1`, projectID).Scan(&dbCount)

	var storageBytes int64
	s.db.Pool.QueryRow(ctx, `SELECT COALESCE(SUM(size_bytes), 0) FROM databases WHERE project_id = $1`, projectID).Scan(&storageBytes)

	var requestCount int64
	s.db.Pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(request_count), 0) FROM usage_logs
		WHERE project_id = $1 AND created_at >= $2`, projectID, periodStart).Scan(&requestCount)

	var apiKeyCount int
	s.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM api_keys WHERE project_id = $1`, projectID).Scan(&apiKeyCount)

	quota := &models.Quota{
		ProjectID:    projectID,
		DatabasesUsed: dbCount,
		StorageBytes: storageBytes,
		RequestsUsed: requestCount,
		APIKeysUsed:  apiKeyCount,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
	}

	planLimits := &models.QuotaLimit{
		MaxDatabases:   1,
		MaxStorageMB:   100,
		MaxConnections: 20,
		MaxRequests:    10000,
		MaxAPIKeys:     5,
	}

	if project.PlanID != nil {
		var maxDB, maxConns, maxReq, maxKeys int
		var maxStorage int64
		err := s.db.Pool.QueryRow(ctx, `
			SELECT max_databases, max_storage_mb, max_connections, max_requests, max_api_keys
			FROM plans WHERE id = $1`, *project.PlanID).Scan(
			&maxDB, &maxStorage, &maxConns, &maxReq, &maxKeys)
		if err == nil {
			planLimits = &models.QuotaLimit{
				MaxDatabases:   maxDB,
				MaxStorageMB:   maxStorage,
				MaxConnections: maxConns,
				MaxRequests:    maxReq,
				MaxAPIKeys:     maxKeys,
			}
		}
	}

	return quota, planLimits, nil
}
