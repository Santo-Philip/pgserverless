package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/nexbic/platform/shared/database"
)

type ExtensionService struct {
	db *database.DB
}

func NewExtensionService(db *database.DB) *ExtensionService {
	return &ExtensionService{db: db}
}

type Extension struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Installed   bool   `json:"installed"`
}

func (s *ExtensionService) ListExtensions(ctx context.Context) ([]Extension, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT
			e.name,
			COALESCE(e.default_version, '') AS version,
			COALESCE(c.comment, e.name) AS description,
			(e.name IN (SELECT extname FROM pg_extension)) AS installed
		FROM pg_available_extensions e
		LEFT JOIN pg_available_extension_versions c ON e.name = c.name AND e.default_version = c.version
		WHERE e.name NOT IN ('plpgsql', 'plpython3u', 'plperlu', 'pltclu')
		ORDER BY
			(e.name IN (SELECT extname FROM pg_extension)) DESC,
			e.name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query extensions: %w", err)
	}
	defer rows.Close()

	var extensions []Extension
	for rows.Next() {
		var ext Extension
		if err := rows.Scan(&ext.Name, &ext.Version, &ext.Description, &ext.Installed); err != nil {
			return nil, fmt.Errorf("scan extension: %w", err)
		}
		extensions = append(extensions, ext)
	}
	return extensions, nil
}

func (s *ExtensionService) ToggleExtension(ctx context.Context, name string, install bool) error {
	if err := validateExtensionName(name); err != nil {
		return err
	}

	if install {
		_, err := s.db.Pool.Exec(ctx, fmt.Sprintf(`CREATE EXTENSION IF NOT EXISTS %s`, quoteIdent(name)))
		if err != nil {
			return fmt.Errorf("install extension %s: %w", name, err)
		}
	} else {
		_, err := s.db.Pool.Exec(ctx, fmt.Sprintf(`DROP EXTENSION IF EXISTS %s`, quoteIdent(name)))
		if err != nil {
			return fmt.Errorf("uninstall extension %s: %w", name, err)
		}
	}
	return nil
}

var blockedExtensions = map[string]bool{
	"plpgsql": true, "plpython3u": true, "plperlu": true, "pltclu": true,
	"amcheck": true, "pageinspect": true, "pgbuffercache": true,
	"pgrowlocks": true, "pgstattuple": true, "auto_explain": true,
	"pg_prewarm": true, "old_snapshot": true, "pg_surgery": true,
	"adminpack": true, "pg_freespacemap": true, "pg_visibility": true,
	"earthdistance": true, "cube": true,
}

func validateExtensionName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("extension name cannot be empty")
	}
	if len(name) > 63 {
		return fmt.Errorf("extension name too long")
	}
	if blockedExtensions[name] {
		return fmt.Errorf("extension %q cannot be toggled by users", name)
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-') {
			return fmt.Errorf("extension name contains invalid character: %c", c)
		}
	}
	return nil
}

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}
