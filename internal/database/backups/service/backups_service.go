package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nexbic/platform/internal/database/backups/models"
	"github.com/nexbic/platform/pkg/database"
	"github.com/nexbic/platform/pkg/helpers"
)

type BackupService struct {
	db            *database.DB
	backupDir     string
	dbHost        string
	dbPort        string
	dbUser        string
	dbPassword    string
	dbName        string
}

func NewBackupService(db *database.DB, backupDir, dbHost, dbPort, dbUser, dbPassword, dbName string) *BackupService {
	if backupDir == "" {
		backupDir = "/data/backups"
	}
	return &BackupService{
		db:         db,
		backupDir:  backupDir,
		dbHost:     dbHost,
		dbPort:     dbPort,
		dbUser:     dbUser,
		dbPassword: dbPassword,
		dbName:     dbName,
	}
}

func (s *BackupService) CreateBackup(ctx context.Context, req *models.CreateBackupRequest, userID uuid.UUID) (*models.BackupInfo, error) {
	if err := os.MkdirAll(s.backupDir, 0750); err != nil {
		return nil, fmt.Errorf("create backup directory: %w", err)
	}

	backupID := uuid.New()
	fileName := fmt.Sprintf("%s_%s_%d.dump", req.DatabaseName, req.Name, time.Now().Unix())
	filePath := filepath.Join(s.backupDir, fileName)

	now := time.Now()
	backup := &models.BackupInfo{
		ID:           backupID,
		Name:         req.Name,
		DatabaseName: req.DatabaseName,
		SizeBytes:    0,
		Status:       "running",
		Type:         req.Type,
		FilePath:     filePath,
		CompletedBy:  userID,
		CreatedAt:    now,
	}

	_, err := s.db.Pool.Exec(ctx, `
		INSERT INTO backup_history (id, name, database_name, size_bytes, status, type, file_path, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		backup.ID, backup.Name, backup.DatabaseName, backup.SizeBytes,
		backup.Status, backup.Type, backup.FilePath, backup.CompletedBy, backup.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert backup record: %w", err)
	}

	go s.runPgDump(backup)

	return backup, nil
}

func (s *BackupService) runPgDump(backup *models.BackupInfo) {
	ctx := context.Background()

	args := []string{
		"-h", s.dbHost,
		"-p", s.dbPort,
		"-U", s.dbUser,
		"-d", backup.DatabaseName,
		"-F", "c",
		"-f", backup.FilePath,
	}
	cmd := exec.CommandContext(ctx, "pg_dump", args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", s.dbPassword))

	output, err := cmd.CombinedOutput()
	if err != nil {
		errMsg := fmt.Sprintf("pg_dump failed: %s", string(output))
		if errMsg == "" {
			errMsg = err.Error()
		}
		s.db.Pool.Exec(ctx, `
			UPDATE backup_history SET status = 'failed', error_message = $1 WHERE id = $2`,
			errMsg, backup.ID)
		return
	}

	info, err := os.Stat(backup.FilePath)
	sizeBytes := int64(0)
	if err == nil {
		sizeBytes = info.Size()
	}

	now := time.Now()
	s.db.Pool.Exec(ctx, `
		UPDATE backup_history SET status = 'completed', size_bytes = $1, completed_at = $2 WHERE id = $3`,
		sizeBytes, now, backup.ID)
}

func (s *BackupService) ListBackups(ctx context.Context, limit, offset int) (*models.BackupHistory, error) {
	var total int
	err := s.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM backup_history`).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("count backups: %w", err)
	}

	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, name, database_name, size_bytes, status, type, file_path,
		       COALESCE(error_message, ''), created_by, completed_at, created_at
		FROM backup_history ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list backups: %w", err)
	}
	defer rows.Close()

	var backups []models.BackupInfo
	for rows.Next() {
		var b models.BackupInfo
		if err := rows.Scan(&b.ID, &b.Name, &b.DatabaseName, &b.SizeBytes,
			&b.Status, &b.Type, &b.FilePath, &b.ErrorMessage,
			&b.CompletedBy, &b.CompletedAt, &b.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan backup: %w", err)
		}
		backups = append(backups, b)
	}

	if backups == nil {
		backups = []models.BackupInfo{}
	}

	return &models.BackupHistory{
		Data:   backups,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (s *BackupService) GetBackup(ctx context.Context, id uuid.UUID) (*models.BackupInfo, error) {
	b := &models.BackupInfo{}
	err := s.db.Pool.QueryRow(ctx, `
		SELECT id, name, database_name, size_bytes, status, type, file_path,
		       COALESCE(error_message, ''), created_by, completed_at, created_at
		FROM backup_history WHERE id = $1`, id).Scan(
		&b.ID, &b.Name, &b.DatabaseName, &b.SizeBytes,
		&b.Status, &b.Type, &b.FilePath, &b.ErrorMessage,
		&b.CompletedBy, &b.CompletedAt, &b.CreatedAt)
	if err != nil {
		return nil, helpers.ErrNoRowsAsNil(err)
	}
	return b, nil
}

func (s *BackupService) DeleteBackup(ctx context.Context, id uuid.UUID) error {
	b, err := s.GetBackup(ctx, id)
	if err != nil {
		return err
	}
	if b == nil {
		return fmt.Errorf("backup not found")
	}

	if b.FilePath != "" {
		os.Remove(b.FilePath)
	}

	_, err = s.db.Pool.Exec(ctx, `DELETE FROM backup_history WHERE id = $1`, id)
	return err
}

func (s *BackupService) RestoreBackup(ctx context.Context, req *models.RestoreRequest, userID uuid.UUID) error {
	backupID, err := uuid.Parse(req.BackupID)
	if err != nil {
		return fmt.Errorf("invalid backup_id")
	}

	b, err := s.GetBackup(ctx, backupID)
	if err != nil {
		return err
	}
	if b == nil {
		return fmt.Errorf("backup not found")
	}
	if b.Status != "completed" {
		return fmt.Errorf("backup status is %s, must be completed", b.Status)
	}

	targetName := req.TargetName
	if targetName == "" {
		targetName = req.DatabaseName
	}

	args := []string{
		"-h", s.dbHost,
		"-p", s.dbPort,
		"-U", s.dbUser,
		"-d", targetName,
		"--clean",
		"-F", "c",
		b.FilePath,
	}
	cmd := exec.CommandContext(ctx, "pg_restore", args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", s.dbPassword))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pg_restore failed: %s", string(output))
	}

	return nil
}

func (s *BackupService) VerifyBackup(ctx context.Context, id uuid.UUID) (*models.BackupInfo, error) {
	b, err := s.GetBackup(ctx, id)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, fmt.Errorf("backup not found")
	}

	info, err := os.Stat(b.FilePath)
	if err != nil {
		b.Status = "corrupted"
		b.ErrorMessage = fmt.Sprintf("file not accessible: %s", err.Error())
	} else if info.Size() == 0 {
		b.Status = "corrupted"
		b.ErrorMessage = "backup file is empty"
	} else {
		b.Status = "verified"
		b.ErrorMessage = ""
		b.SizeBytes = info.Size()
	}

	s.db.Pool.Exec(ctx, `
		UPDATE backup_history SET status = $1, error_message = $2, size_bytes = $3 WHERE id = $4`,
		b.Status, b.ErrorMessage, b.SizeBytes, id)

	return b, nil
}

func (s *BackupService) GetBackupFilePath(ctx context.Context, id uuid.UUID) (string, error) {
	b, err := s.GetBackup(ctx, id)
	if err != nil {
		return "", err
	}
	if b == nil {
		return "", pgx.ErrNoRows
	}
	return b.FilePath, nil
}
