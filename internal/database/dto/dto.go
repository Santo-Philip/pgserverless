package dto

import "github.com/nexbic/platform/internal/database/models"

type CreateDatabaseRequest struct {
	Name       string `json:"name"`
	ProjectID  string `json:"project_id"`
}

type CreateDatabaseUserRequest struct {
	Name     string `json:"name"`
}

type CreateTableRequest struct {
	Name    string                `json:"name"`
	Columns []models.TableColumn  `json:"columns"`
}

type AddColumnRequest struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Nullable     bool   `json:"nullable"`
	DefaultValue string `json:"default_value,omitempty"`
}

type InsertRowRequest struct {
	Values map[string]any `json:"values"`
}

type UpdateRowRequest struct {
	Values map[string]any `json:"values"`
	Where  map[string]any `json:"where"`
}

type DeleteRowRequest struct {
	Where map[string]any `json:"where"`
}

type RunSQLRequest struct {
	Query string `json:"query"`
}

type ToggleExtensionRequest struct {
	Name    string `json:"name"`
	Install bool   `json:"install"`
}
