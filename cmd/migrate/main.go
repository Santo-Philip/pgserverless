package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nexbic/platform/shared/config"
)

func main() {
	cfg := config.Load()
	migrationsDir := "postgres/init"
	if len(os.Args) > 1 {
		migrationsDir = os.Args[1]
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User, cfg.Database.Password,
		cfg.Database.Host, cfg.Database.Port,
		cfg.Database.DBName, cfg.Database.SSLMode)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		slog.Error("failed to connect", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		slog.Error("failed to read migrations directory", "path", migrationsDir, "error", err)
		os.Exit(1)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, f := range files {
		path := filepath.Join(migrationsDir, f)
		slog.Info("applying migration", "file", f)

		sql, err := os.ReadFile(path)
		if err != nil {
			slog.Error("failed to read migration", "file", f, "error", err)
			os.Exit(1)
		}

		if _, err := pool.Exec(context.Background(), string(sql)); err != nil {
			slog.Error("failed to apply migration", "file", f, "error", err)
			os.Exit(1)
		}

		slog.Info("migration applied", "file", f)
	}

	slog.Info("all migrations applied successfully")
}
