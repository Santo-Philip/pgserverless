package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexbic/platform/internal/plan/dto"
	"github.com/nexbic/platform/internal/plan/models"
	"github.com/nexbic/platform/internal/plan/repository"
)

type PlanService struct {
	repo *repository.PlanRepository
}

func NewPlanService(repo *repository.PlanRepository) *PlanService {
	return &PlanService{repo: repo}
}

func (s *PlanService) Create(ctx context.Context, req *dto.CreatePlanRequest) (*models.Plan, error) {
	existing, err := s.repo.GetBySlug(ctx, req.Slug)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("plan slug already exists")
	}

	plan := &models.Plan{
		Name:           req.Name,
		Slug:           req.Slug,
		Description:    req.Description,
		MaxDatabases:   req.MaxDatabases,
		MaxStorageMB:   req.MaxStorageMB,
		MaxConnections: req.MaxConnections,
		MaxRequests:    req.MaxRequests,
		MaxAPIKeys:     req.MaxAPIKeys,
		Price:          req.Price,
		IsActive:       true,
	}

	if err := s.repo.Create(ctx, plan); err != nil {
		return nil, fmt.Errorf("create plan: %w", err)
	}

	return plan, nil
}

func (s *PlanService) List(ctx context.Context) ([]models.Plan, error) {
	return s.repo.List(ctx)
}

func (s *PlanService) GetByID(ctx context.Context, id uuid.UUID) (*models.Plan, error) {
	plan, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, fmt.Errorf("plan not found")
	}
	return plan, nil
}

func (s *PlanService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdatePlanRequest) error {
	plan, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if plan == nil {
		return fmt.Errorf("plan not found")
	}

	updates := make(map[string]any)
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.MaxDatabases != nil {
		updates["max_databases"] = *req.MaxDatabases
	}
	if req.MaxStorageMB != nil {
		updates["max_storage_mb"] = *req.MaxStorageMB
	}
	if req.MaxConnections != nil {
		updates["max_connections"] = *req.MaxConnections
	}
	if req.MaxRequests != nil {
		updates["max_requests"] = *req.MaxRequests
	}
	if req.MaxAPIKeys != nil {
		updates["max_api_keys"] = *req.MaxAPIKeys
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	return s.repo.Update(ctx, id, updates)
}
