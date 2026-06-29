package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/nexbic/platform/internal/files/models"
)

type FilesService struct {
	baseDir string
}

func NewFilesService(baseDir string) *FilesService {
	return &FilesService{baseDir: baseDir}
}

func userDir(baseDir string, userID uuid.UUID) string {
	return filepath.Join(baseDir, userID.String())
}

func (s *FilesService) Upload(ctx context.Context, userID uuid.UUID, name, mimeType string, reader io.Reader) (*models.File, error) {
	dir := userDir(s.baseDir, userID)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, fmt.Errorf("create directory: %w", err)
	}

	id := uuid.New().String()
	filename := fmt.Sprintf("%s-%s", id, name)
	filePath := filepath.Join(dir, filename)

	f, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	size, err := io.Copy(f, reader)
	if err != nil {
		os.Remove(filePath)
		return nil, fmt.Errorf("write file: %w", err)
	}

	file := &models.File{
		ID:        id,
		Name:      name,
		Size:      size,
		MimeType:  mimeType,
		CreatedAt: time.Now(),
	}
	return file, nil
}

func (s *FilesService) List(ctx context.Context, userID uuid.UUID) ([]models.File, error) {
	dir := userDir(s.baseDir, userID)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.File{}, nil
		}
		return nil, fmt.Errorf("read directory: %w", err)
	}

	var files []models.File
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		fileID, fileName := splitFileID(info.Name())
		if fileID == "" {
			continue
		}
		files = append(files, models.File{
			ID:        fileID,
			Name:      fileName,
			Size:      info.Size(),
			CreatedAt: info.ModTime(),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].CreatedAt.After(files[j].CreatedAt)
	})
	return files, nil
}

func (s *FilesService) Get(ctx context.Context, userID uuid.UUID, fileID string) (*models.File, io.ReadCloser, error) {
	dir := userDir(s.baseDir, userID)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("file not found")
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		id, name := splitFileID(entry.Name())
		if id == fileID {
			info, _ := entry.Info()
			f, err := os.Open(filepath.Join(dir, entry.Name()))
			if err != nil {
				return nil, nil, fmt.Errorf("file not found")
			}
			file := &models.File{
				ID: id,
				Name:      name,
				Size:      info.Size(),
				CreatedAt: info.ModTime(),
			}
			return file, f, nil
		}
	}
	return nil, nil, fmt.Errorf("file not found")
}

func (s *FilesService) Delete(ctx context.Context, userID uuid.UUID, fileID string) error {
	dir := userDir(s.baseDir, userID)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("file not found")
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		id, _ := splitFileID(entry.Name())
		if id == fileID {
			return os.Remove(filepath.Join(dir, entry.Name()))
		}
	}
	return fmt.Errorf("file not found")
}

func splitFileID(filename string) (id, name string) {
	for i := 0; i < len(filename); i++ {
		if filename[i] == '-' {
			potentialID := filename[:i]
			if _, err := uuid.Parse(potentialID); err == nil {
				return potentialID, filename[i+1:]
			}
		}
	}
	return "", filename
}
