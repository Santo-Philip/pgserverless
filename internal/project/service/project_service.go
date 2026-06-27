package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	projectdto "github.com/nexbic/platform/internal/project/dto"
	projectmodels "github.com/nexbic/platform/internal/project/models"
	projectrepo "github.com/nexbic/platform/internal/project/repository"
)

type ProjectService struct {
	repo *projectrepo.ProjectRepository
}

func NewProjectService(repo *projectrepo.ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) Create(ctx context.Context, req *projectdto.CreateProjectRequest) (*projectmodels.Project, error) {
	existing, err := s.repo.GetBySlug(ctx, req.Slug)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("slug already taken")
	}

	var planID *uuid.UUID
	if req.PlanID != "" {
		if pid, err := uuid.Parse(req.PlanID); err == nil {
			planID = &pid
		}
	}

	project := &projectmodels.Project{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		PlanID:      planID,
		Status:      "active",
	}

	if err := s.repo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}

	return project, nil
}

func (s *ProjectService) GetByID(ctx context.Context, id uuid.UUID) (*projectmodels.Project, error) {
	project, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, fmt.Errorf("project not found")
	}
	return project, nil
}

func (s *ProjectService) List(ctx context.Context, limit, offset int) ([]projectmodels.Project, int, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *ProjectService) Update(ctx context.Context, id uuid.UUID, req *projectdto.UpdateProjectRequest) error {
	project, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if project == nil {
		return fmt.Errorf("project not found")
	}

	updates := make(map[string]any)
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.PlanID != nil {
		if pid, err := uuid.Parse(*req.PlanID); err == nil {
			updates["plan_id"] = pid
		}
	}

	return s.repo.Update(ctx, id, updates)
}

func (s *ProjectService) Delete(ctx context.Context, id uuid.UUID) error {
	project, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if project == nil {
		return fmt.Errorf("project not found")
	}

	return s.repo.Delete(ctx, id)
}
