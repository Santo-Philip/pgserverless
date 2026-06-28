package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/nexbic/platform/internal/extensions/models"
	"github.com/nexbic/platform/pkg/database"
)

type ExtensionsService struct {
	db *database.DB
}

func NewExtensionsService(db *database.DB) *ExtensionsService {
	return &ExtensionsService{db: db}
}

func (s *ExtensionsService) ListExtensions(ctx context.Context) ([]models.ExtensionInfo, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT
			e.name,
			COALESCE(e.default_version, ''),
			COALESCE(e.comment, ''),
			CASE WHEN pg.extname IS NOT NULL THEN true ELSE false END,
			COALESCE(pg.extversion, '')
		FROM pg_available_extensions e
		LEFT JOIN pg_extension pg ON e.name = pg.extname
		ORDER BY e.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var extensions []models.ExtensionInfo
	for rows.Next() {
		var ext models.ExtensionInfo
		if err := rows.Scan(&ext.Name, &ext.Version, &ext.Description, &ext.Installed, &ext.InstalledVersion); err != nil {
			return nil, err
		}
		extensions = append(extensions, ext)
	}

	if extensions == nil {
		extensions = []models.ExtensionInfo{}
	}

	return extensions, nil
}

func (s *ExtensionsService) InstallExtension(ctx context.Context, name, version string) error {
	query := fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS %s", quoteIdent(name))
	if version != "" {
		query += fmt.Sprintf(" VERSION %s", quoteLiteral(version))
	}
	_, err := s.db.Pool.Exec(ctx, query)
	return err
}

func (s *ExtensionsService) UninstallExtension(ctx context.Context, name string) error {
	_, err := s.db.Pool.Exec(ctx, fmt.Sprintf("DROP EXTENSION IF EXISTS %s", quoteIdent(name)))
	return err
}

func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func quoteLiteral(val string) string {
	return "'" + strings.ReplaceAll(val, "'", "''") + "'"
}
