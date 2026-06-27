package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexbic/platform/internal/audit/models"
	"github.com/nexbic/platform/pkg/database"
)

type AuditService struct {
	db *database.DB
}

func NewAuditService(db *database.DB) *AuditService {
	return &AuditService{db: db}
}

func (s *AuditService) Create(ctx context.Context, entry *models.AuditLog) error {
	entry.ID = uuid.New()
	entry.CreatedAt = time.Now()

	_, err := s.db.Pool.Exec(ctx, `
		INSERT INTO audit_logs (id, actor_id, action, resource, resource_id, metadata, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		entry.ID, entry.ActorID, entry.Action, entry.Resource, entry.ResourceID,
		entry.Metadata, entry.IPAddress, entry.UserAgent, entry.CreatedAt,
	)
	return err
}

func (s *AuditService) List(ctx context.Context, limit, offset int) ([]models.AuditLog, int, error) {
	var total int
	err := s.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, actor_id, action, resource, COALESCE(resource_id, ''), metadata, COALESCE(ip_address, ''), COALESCE(user_agent, ''), created_at
		FROM audit_logs ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var l models.AuditLog
		if err := rows.Scan(&l.ID, &l.ActorID, &l.Action, &l.Resource, &l.ResourceID,
			&l.Metadata, &l.IPAddress, &l.UserAgent, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}

	if logs == nil {
		logs = []models.AuditLog{}
	}

	return logs, total, nil
}

func (s *AuditService) ListByResource(ctx context.Context, resource, resourceID string, limit, offset int) ([]models.AuditLog, int, error) {
	var total int
	err := s.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs WHERE resource = $1 AND resource_id = $2`,
		resource, resourceID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, actor_id, action, resource, COALESCE(resource_id, ''), metadata, COALESCE(ip_address, ''), COALESCE(user_agent, ''), created_at
		FROM audit_logs WHERE resource = $1 AND resource_id = $2
		ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
		resource, resourceID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var l models.AuditLog
		if err := rows.Scan(&l.ID, &l.ActorID, &l.Action, &l.Resource, &l.ResourceID,
			&l.Metadata, &l.IPAddress, &l.UserAgent, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}

	if logs == nil {
		logs = []models.AuditLog{}
	}

	return logs, total, nil
}
