package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nexbic/platform/shared/database"
)

type SettingsService struct {
	db *database.DB
}

func NewSettingsService(db *database.DB) *SettingsService {
	return &SettingsService{db: db}
}

func (s *SettingsService) Get(ctx context.Context) (map[string]interface{}, error) {
	var rawSettings []byte
	err := s.db.QueryRow(ctx, "SELECT settings::text FROM platform_settings WHERE id = 1").Scan(&rawSettings)
	if err != nil {
		return nil, fmt.Errorf("load settings: %w", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(rawSettings, &settings); err != nil {
		return nil, fmt.Errorf("parse settings: %w", err)
	}

	return settings, nil
}

func (s *SettingsService) Update(ctx context.Context, updates map[string]interface{}) error {
	current, err := s.Get(ctx)
	if err != nil {
		return err
	}

	for k, v := range updates {
		current[k] = v
	}

	merged, err := json.Marshal(current)
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}

	if err := s.db.Exec(ctx, "UPDATE platform_settings SET settings = $1::jsonb, updated_at = NOW() WHERE id = 1", string(merged)); err != nil {
		return fmt.Errorf("save settings: %w", err)
	}

	return nil
}
