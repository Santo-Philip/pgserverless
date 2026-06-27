package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nexbic/platform/config"
	"github.com/nexbic/platform/pkg/database"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()

	db, err := database.New(ctx, cfg.Database)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := runMigrations(ctx, db); err != nil {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}

	slog.Info("migrations complete")
}

func runMigrations(ctx context.Context, db *database.DB) error {
	dir := "migrations"
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, file := range files {
		path := filepath.Join(dir, file)
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", file, err)
		}

		slog.Info("running migration", "file", file)

		sql := string(content)

		statements := splitSQL(sql)
		for i, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := db.Pool.Exec(ctx, stmt); err != nil {
				return fmt.Errorf("migration %s statement %d: %w", file, i+1, err)
			}
		}
	}

	return nil
}

func splitSQL(sql string) []string {
	var statements []string
	current := strings.Builder{}
	for _, line := range strings.Split(sql, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "--") {
			continue
		}

		if trimmed == "" {
			continue
		}

		current.WriteString(line)
		current.WriteString("\n")

		if strings.HasSuffix(trimmed, ";") {
			statements = append(statements, current.String())
			current.Reset()
		}
	}

	if current.Len() > 0 {
		remaining := strings.TrimSpace(current.String())
		if remaining != "" {
			statements = append(statements, remaining)
		}
	}

	return statements
}
