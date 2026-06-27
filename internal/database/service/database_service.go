package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexbic/platform/internal/database/dto"
	"github.com/nexbic/platform/internal/database/models"
	"github.com/nexbic/platform/internal/database/repository"
	"github.com/nexbic/platform/pkg/password"
)

type DatabaseService struct {
	repo *repository.DatabaseRepository
}

func NewDatabaseService(repo *repository.DatabaseRepository) *DatabaseService {
	return &DatabaseService{repo: repo}
}

func (s *DatabaseService) Create(ctx context.Context, req *dto.CreateDatabaseRequest) (*models.Database, error) {
	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("invalid project_id")
	}

	rawPassword, err := password.GenerateToken(16)
	if err != nil {
		return nil, fmt.Errorf("generate password: %w", err)
	}

	dbUser := fmt.Sprintf("db_user_%s", req.Name)
	dbEntry := &models.Database{
		ProjectID:  projectID,
		Name:       req.Name,
		SchemaName: req.Name,
		DBUser:     dbUser,
		DBPassword: rawPassword,
		Status:     "provisioning",
		SizeBytes:  0,
	}

	if err := s.repo.ProvisionSchema(ctx, dbEntry); err != nil {
		return nil, fmt.Errorf("provision schema: %w", err)
	}

	dbEntry.Status = "active"
	dbEntry.ConnString = fmt.Sprintf("postgres://%s:%s@localhost:5432/%s?sslmode=disable",
		dbEntry.DBUser, rawPassword, dbEntry.Name)

	if err := s.repo.Create(ctx, dbEntry); err != nil {
		return nil, fmt.Errorf("create database record: %w", err)
	}

	return dbEntry, nil
}

func (s *DatabaseService) GetByID(ctx context.Context, id uuid.UUID) (*models.Database, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *DatabaseService) ListByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]models.Database, int, error) {
	return s.repo.ListByProject(ctx, projectID, limit, offset)
}

func (s *DatabaseService) Delete(ctx context.Context, id uuid.UUID) error {
	dbEntry, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if dbEntry == nil {
		return fmt.Errorf("database not found")
	}

	if err := s.repo.DropSchema(ctx, dbEntry); err != nil {
		return fmt.Errorf("drop schema: %w", err)
	}

	return s.repo.Delete(ctx, id)
}

func (s *DatabaseService) RunSQL(ctx context.Context, dbID uuid.UUID, query string) ([]map[string]any, error) {
	return s.repo.RunSQL(ctx, dbID, query)
}

func (s *DatabaseService) ListTables(ctx context.Context, dbID uuid.UUID) ([]models.TableInfo, error) {
	return s.repo.ListTables(ctx, dbID)
}

func (s *DatabaseService) GetTableData(ctx context.Context, dbID uuid.UUID, table string, limit, offset int) ([]map[string]any, error) {
	return s.repo.GetTableData(ctx, dbID, table, limit, offset)
}

func (s *DatabaseService) CreateTable(ctx context.Context, dbID uuid.UUID, req *dto.CreateTableRequest) error {
	return s.repo.CreateTable(ctx, dbID, req.Name, req.Columns)
}

func (s *DatabaseService) AddColumn(ctx context.Context, dbID uuid.UUID, table string, req *dto.AddColumnRequest) error {
	col := &models.TableColumn{
		Name:         req.Name,
		Type:         req.Type,
		Nullable:     req.Nullable,
		DefaultValue: req.DefaultValue,
	}
	return s.repo.AddColumn(ctx, dbID, table, col)
}

func (s *DatabaseService) InsertRow(ctx context.Context, dbID uuid.UUID, table string, req *dto.InsertRowRequest) (map[string]any, error) {
	return s.repo.InsertRow(ctx, dbID, table, req.Values)
}

func (s *DatabaseService) UpdateRow(ctx context.Context, dbID uuid.UUID, table string, req *dto.UpdateRowRequest) ([]map[string]any, error) {
	return s.repo.UpdateRow(ctx, dbID, table, req.Values, req.Where)
}

func (s *DatabaseService) DeleteRow(ctx context.Context, dbID uuid.UUID, table string, req *dto.DeleteRowRequest) (int64, error) {
	return s.repo.DeleteRow(ctx, dbID, table, req.Where)
}

func (s *DatabaseService) ListExtensions(ctx context.Context) ([]models.Extension, error) {
	exts, err := s.repo.ListAvailableExtensions(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []models.Extension
	for _, ext := range exts {
		if !repository.IsExtensionBlocked(ext.Name) {
			filtered = append(filtered, ext)
		}
	}

	return filtered, nil
}

func (s *DatabaseService) ToggleExtension(ctx context.Context, req *dto.ToggleExtensionRequest) error {
	if repository.IsExtensionBlocked(req.Name) {
		return fmt.Errorf("extension %s is blocked", req.Name)
	}
	return s.repo.ToggleExtension(ctx, req.Name, req.Install)
}
