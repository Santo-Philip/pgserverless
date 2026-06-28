package service

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/nexbic/platform/internal/dashboard/models"
	"github.com/nexbic/platform/pkg/database"
)

type DashboardService struct {
	db *database.DB
}

func NewDashboardService(db *database.DB) *DashboardService {
	return &DashboardService{db: db}
}

func (s *DashboardService) GetOverview(ctx context.Context) (*models.DashboardOverview, error) {
	stats, err := s.getDBStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("get db stats: %w", err)
	}

	schemas, err := s.getSchemaInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("get schema info: %w", err)
	}

	tableCounts, err := s.getTableCounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("get table counts: %w", err)
	}

	return &models.DashboardOverview{
		Stats:       *stats,
		Schemas:     schemas,
		TableCounts: tableCounts,
	}, nil
}

func (s *DashboardService) getDBStats(ctx context.Context) (*models.DBStats, error) {
	var stats models.DBStats

	err := s.db.Pool.QueryRow(ctx, `SELECT current_setting('server_version')`).Scan(&stats.Version)
	if err != nil {
		return nil, err
	}

	err = s.db.Pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(blks_hit::numeric) / NULLIF(SUM(blks_hit::numeric + blks_read::numeric), 0), 0)
		FROM pg_stat_database
		WHERE datname IS NOT NULL AND datname NOT IN ('template0', 'template1')
	`).Scan(&stats.CacheHitRate)
	if err != nil {
		return nil, err
	}
	stats.CacheHitRate = math.Round(stats.CacheHitRate*10000) / 100

	err = s.db.Pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(xact_commit + xact_rollback), 0)
		FROM pg_stat_database
		WHERE datname IS NOT NULL AND datname NOT IN ('template0', 'template1')
	`).Scan(&stats.Transactions)
	if err != nil {
		return nil, err
	}

	err = s.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM pg_database WHERE datistemplate = false`).Scan(&stats.Databases)
	if err != nil {
		return nil, err
	}

	err = s.db.Pool.QueryRow(ctx, `
		SELECT COALESCE(COUNT(*), 0) FROM pg_stat_activity WHERE state = 'active'
	`).Scan(&stats.ActiveConns)
	if err != nil {
		return nil, err
	}

	err = s.db.Pool.QueryRow(ctx, `
		SELECT pg_size_pretty(COALESCE(SUM(pg_database_size(datname)), 0))
		FROM pg_database
		WHERE datistemplate = false
	`).Scan(&stats.TotalSize)
	if err != nil {
		return nil, err
	}

	var uptimeSeconds float64
	err = s.db.Pool.QueryRow(ctx, `
		SELECT EXTRACT(EPOCH FROM current_timestamp - pg_postmaster_start_time())
	`).Scan(&uptimeSeconds)
	if err != nil {
		return nil, err
	}
	stats.Uptime = formatUptime(uptimeSeconds)

	repl, err := s.getReplicationStatus(ctx)
	if err != nil {
		return nil, err
	}
	stats.Replication = repl

	return &stats, nil
}

func (s *DashboardService) getReplicationStatus(ctx context.Context) ([]models.ReplicationInfo, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT
			COALESCE(application_name, ''),
			COALESCE(state, ''),
			COALESCE(sync_state, ''),
			COALESCE(client_addr::text, ''),
			COALESCE(pg_size_pretty(pg_wal_lsn_diff(pg_current_wal_lsn(), replay_lsn)), '0')
		FROM pg_stat_replication
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.ReplicationInfo
	for rows.Next() {
		var r models.ReplicationInfo
		if err := rows.Scan(&r.AppName, &r.State, &r.SyncState, &r.ClientAddr, &r.LagBytes); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

func (s *DashboardService) getSchemaInfo(ctx context.Context) ([]models.SchemaInfo, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT
			n.nspname AS schema_name,
			COALESCE(u.usename, '') AS owner,
			COALESCE((
				SELECT SUM(pg_total_relation_size(c.oid))
				FROM pg_class c
				WHERE c.relnamespace = n.oid AND c.relkind IN ('r', 'm')
			), 0) AS size_bytes,
			(
				SELECT COUNT(*)
				FROM pg_class c
				WHERE c.relnamespace = n.oid AND c.relkind IN ('r', 'm')
			) AS table_count
		FROM pg_namespace n
		LEFT JOIN pg_user u ON n.nspowner = u.usesysid
		WHERE n.nspname NOT IN ('information_schema', 'pg_catalog', 'pg_toast')
			AND n.nspname NOT LIKE 'pg_temp_%'
			AND n.nspname NOT LIKE 'pg_toast_temp_%'
		ORDER BY size_bytes DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.SchemaInfo
	for rows.Next() {
		var s models.SchemaInfo
		if err := rows.Scan(&s.SchemaName, &s.Owner, &s.SizeBytes, &s.TableCount); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, nil
}

func (s *DashboardService) getTableCounts(ctx context.Context) ([]models.TableCountResult, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT
			n.nspname AS schema,
			COUNT(c.oid) AS count
		FROM pg_namespace n
		LEFT JOIN pg_class c ON c.relnamespace = n.oid AND c.relkind IN ('r', 'm')
		WHERE n.nspname NOT IN ('information_schema', 'pg_catalog', 'pg_toast')
			AND n.nspname NOT LIKE 'pg_temp_%'
			AND n.nspname NOT LIKE 'pg_toast_temp_%'
		GROUP BY n.nspname
		ORDER BY count DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.TableCountResult
	for rows.Next() {
		var t models.TableCountResult
		if err := rows.Scan(&t.Schema, &t.Count); err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, nil
}

func formatUptime(seconds float64) string {
	days := int(seconds) / 86400
	hours := (int(seconds) % 86400) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60

	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 || days > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 || hours > 0 || days > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	parts = append(parts, fmt.Sprintf("%ds", secs))

	return strings.Join(parts, " ")
}
