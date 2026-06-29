package service

import (
	"context"
	"fmt"
	"math"

	"github.com/nexbic/platform/internal/database/monitoring/models"
	"github.com/nexbic/platform/pkg/database"
)

type MonitoringService struct {
	db *database.DB
}

func NewMonitoringService(db *database.DB) *MonitoringService {
	return &MonitoringService{db: db}
}

func (s *MonitoringService) GetActiveSessions(ctx context.Context) ([]models.ActiveSession, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT
			pid,
			COALESCE(usename, '') AS usename,
			COALESCE(datname, '') AS datname,
			COALESCE(state, 'unknown') AS state,
			COALESCE(query, '') AS query,
			wait_event,
			to_char(query_start AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
			COALESCE(client_addr::text, ''),
			to_char(backend_start AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		FROM pg_stat_activity
		WHERE state != 'idle'
		  AND backend_type = 'client backend'
		ORDER BY query_start DESC NULLS LAST`)
	if err != nil {
		return nil, fmt.Errorf("get active sessions: %w", err)
	}
	defer rows.Close()

	var sessions []models.ActiveSession
	for rows.Next() {
		var s models.ActiveSession
		var clientAddr string
		if err := rows.Scan(&s.PID, &s.User, &s.Database, &s.State, &s.Query,
			&s.WaitEvent, &s.QueryStart, &clientAddr, &s.BackendStart); err != nil {
			return nil, fmt.Errorf("scan active session: %w", err)
		}
		if clientAddr != "" {
			s.ClientAddr = &clientAddr
		}
		sessions = append(sessions, s)
	}

	if sessions == nil {
		sessions = []models.ActiveSession{}
	}

	return sessions, nil
}

func (s *MonitoringService) GetSlowQueries(ctx context.Context, minSeconds float64) ([]models.SlowQuery, error) {
	interval := fmt.Sprintf("%.0f seconds", minSeconds)

	rows, err := s.db.Pool.Query(ctx, `
		SELECT
			pid,
			COALESCE(usename, '') AS usename,
			COALESCE(datname, '') AS datname,
			COALESCE(query, '') AS query,
			EXTRACT(EPOCH FROM (NOW() - query_start))::float8 AS duration,
			COALESCE(state, 'unknown') AS state,
			to_char(query_start AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
			wait_event
		FROM pg_stat_activity
		WHERE state = 'active'
		  AND backend_type = 'client backend'
		  AND query_start IS NOT NULL
		  AND (NOW() - query_start) > $1::interval
		ORDER BY duration DESC`, interval)
	if err != nil {
		return nil, fmt.Errorf("get slow queries: %w", err)
	}
	defer rows.Close()

	var queries []models.SlowQuery
	for rows.Next() {
		var q models.SlowQuery
		if err := rows.Scan(&q.PID, &q.User, &q.Database, &q.Query, &q.Duration,
			&q.State, &q.QueryStart, &q.WaitEvent); err != nil {
			return nil, fmt.Errorf("scan slow query: %w", err)
		}
		// Round to 2 decimal places
		q.Duration = math.Round(q.Duration*100) / 100
		queries = append(queries, q)
	}

	if queries == nil {
		queries = []models.SlowQuery{}
	}

	return queries, nil
}

func (s *MonitoringService) GetLocks(ctx context.Context) ([]models.LockInfo, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT
			l.pid,
			l.locktype,
			l.mode,
			l.granted,
			COALESCE(c.relname::text, '') AS relation,
			l.relation::int4 AS relation_id,
			COALESCE(a.usename, '') AS usename,
			COALESCE(a.query, '') AS query
		FROM pg_locks l
		LEFT JOIN pg_class c ON l.relation = c.oid
		LEFT JOIN pg_stat_activity a ON l.pid = a.pid
		WHERE a.backend_type = 'client backend'
		   OR a.backend_type IS NULL
		ORDER BY l.pid, l.granted DESC, l.mode`)
	if err != nil {
		return nil, fmt.Errorf("get locks: %w", err)
	}
	defer rows.Close()

	var locks []models.LockInfo
	for rows.Next() {
		var l models.LockInfo
		var relation string
		var relationID *int32
		if err := rows.Scan(&l.PID, &l.LockType, &l.Mode, &l.Granted,
			&relation, &relationID, &l.User, &l.Query); err != nil {
			return nil, fmt.Errorf("scan lock: %w", err)
		}
		if relation != "" {
			l.Relation = &relation
		}
		if relationID != nil && *relationID > 0 {
			l.RelationID = relationID
		}
		locks = append(locks, l)
	}

	if locks == nil {
		locks = []models.LockInfo{}
	}

	return locks, nil
}

func (s *MonitoringService) GetWaitingQueries(ctx context.Context) ([]models.WaitingQuery, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT
			pid,
			COALESCE(usename, '') AS usename,
			COALESCE(datname, '') AS datname,
			COALESCE(query, '') AS query,
			EXTRACT(EPOCH FROM (NOW() - query_start))::float8 AS duration,
			COALESCE(wait_event, 'unknown') AS wait_event,
			COALESCE(state, 'unknown') AS state
		FROM pg_stat_activity
		WHERE wait_event IS NOT NULL
		  AND backend_type = 'client backend'
		  AND state = 'active'
		ORDER BY duration DESC NULLS LAST`)
	if err != nil {
		return nil, fmt.Errorf("get waiting queries: %w", err)
	}
	defer rows.Close()

	var queries []models.WaitingQuery
	for rows.Next() {
		var q models.WaitingQuery
		if err := rows.Scan(&q.PID, &q.User, &q.Database, &q.Query, &q.Duration,
			&q.WaitEvent, &q.State); err != nil {
			return nil, fmt.Errorf("scan waiting query: %w", err)
		}
		q.Duration = math.Round(q.Duration*100) / 100
		queries = append(queries, q)
	}

	if queries == nil {
		queries = []models.WaitingQuery{}
	}

	return queries, nil
}

func (s *MonitoringService) GetQueryStats(ctx context.Context, limit int) ([]models.QueryStats, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	// Try pg_stat_statements first
	rows, err := s.db.Pool.Query(ctx, `
		SELECT
			query,
			calls,
			total_exec_time / 1000.0 AS total_time_ms,
			mean_exec_time / 1000.0 AS mean_time_ms,
			rows,
			shared_blks_hit,
			shared_blks_read
		FROM pg_stat_statements
		WHERE query NOT LIKE '%pg_stat_statements%'
		ORDER BY total_exec_time DESC
		LIMIT $1`, limit)
	if err != nil {
		// Fallback to pg_stat_activity
		return s.getQueryStatsFallback(ctx, limit)
	}
	defer rows.Close()

	var stats []models.QueryStats
	for rows.Next() {
		var s models.QueryStats
		if err := rows.Scan(&s.Query, &s.Calls, &s.TotalTime, &s.MeanTime,
			&s.Rows, &s.SharedBlksHit, &s.SharedBlksRead); err != nil {
			return nil, fmt.Errorf("scan query stats: %w", err)
		}
		stats = append(stats, s)
	}

	if stats == nil {
		stats = []models.QueryStats{}
	}

	return stats, nil
}

func (s *MonitoringService) getQueryStatsFallback(ctx context.Context, limit int) ([]models.QueryStats, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT
			query,
			null::int8 AS calls,
			null::float8 AS total_time,
			null::float8 AS mean_time,
			null::int8 AS rows,
			null::int8 AS blks_hit,
			null::int8 AS blks_read
		FROM pg_stat_activity
		WHERE state = 'active'
		  AND backend_type = 'client backend'
		  AND query IS NOT NULL
		  AND query != ''
		LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("get query stats fallback: %w", err)
	}
	defer rows.Close()

	var stats []models.QueryStats
	for rows.Next() {
		var s models.QueryStats
		if err := rows.Scan(&s.Query, &s.Calls, &s.TotalTime, &s.MeanTime,
			&s.Rows, &s.SharedBlksHit, &s.SharedBlksRead); err != nil {
			return nil, fmt.Errorf("scan fallback query stats: %w", err)
		}
		stats = append(stats, s)
	}

	if stats == nil {
		stats = []models.QueryStats{}
	}

	return stats, nil
}

func (s *MonitoringService) GetConnectionStats(ctx context.Context) (*models.ConnectionStats, error) {
	stats := &models.ConnectionStats{}

	err := s.db.Pool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN state = 'active' THEN 1 ELSE 0 END), 0) AS active,
			COALESCE(SUM(CASE WHEN state = 'idle' THEN 1 ELSE 0 END), 0) AS idle,
			COALESCE(SUM(CASE WHEN state = 'idle in transaction' THEN 1 ELSE 0 END), 0) AS idle_in_xact,
			COALESCE(SUM(CASE WHEN wait_event IS NOT NULL AND state = 'active' THEN 1 ELSE 0 END), 0) AS waiting,
			COUNT(*) AS total
		FROM pg_stat_activity
		WHERE backend_type = 'client backend'`).Scan(
		&stats.Active, &stats.Idle, &stats.IdleInTransaction, &stats.Waiting, &stats.Total)
	if err != nil {
		return nil, fmt.Errorf("get connection counts: %w", err)
	}

	err = s.db.Pool.QueryRow(ctx, `
		SELECT COALESCE(setting::int, 100) FROM pg_settings WHERE name = 'max_connections'`).Scan(&stats.MaxConnections)
	if err != nil {
		stats.MaxConnections = 100
	}

	return stats, nil
}

func (s *MonitoringService) GetCacheStats(ctx context.Context) (*models.CacheStats, error) {
	stats := &models.CacheStats{}

	err := s.db.Pool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(blks_hit), 0) AS hits,
			COALESCE(SUM(blks_read), 0) AS reads
		FROM pg_stat_database
		WHERE datname IS NOT NULL`).Scan(&stats.Hits, &stats.Reads)
	if err != nil {
		return nil, fmt.Errorf("get cache stats: %w", err)
	}

	total := stats.Hits + stats.Reads
	if total > 0 {
		stats.HitRate = math.Round(float64(stats.Hits)/float64(total)*10000) / 100
	}

	return stats, nil
}

func (s *MonitoringService) GetDatabaseStats(ctx context.Context) ([]models.DatabaseStat, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT
			COALESCE(d.datname, '') AS datname,
			d.numbackends,
			d.xact_commit,
			d.xact_rollback,
			d.blks_read,
			d.blks_hit,
			CASE WHEN (d.blks_hit + d.blks_read) > 0
				THEN ROUND((d.blks_hit::numeric / (d.blks_hit + d.blks_read) * 100), 2)
				ELSE 0
			END AS cache_hit_ratio,
			pg_database_size(d.datname) AS size_bytes
		FROM pg_stat_database d
		WHERE d.datname IS NOT NULL
		  AND d.datname NOT IN ('template0', 'template1')
		ORDER BY d.datname`)
	if err != nil {
		return nil, fmt.Errorf("get database stats: %w", err)
	}
	defer rows.Close()

	var dbs []models.DatabaseStat
	for rows.Next() {
		var d models.DatabaseStat
		if err := rows.Scan(&d.Datname, &d.NumBackends, &d.XactCommit, &d.XactRollback,
			&d.BlksRead, &d.BlksHit, &d.CacheHitRatio, &d.SizeBytes); err != nil {
			return nil, fmt.Errorf("scan database stat: %w", err)
		}
		dbs = append(dbs, d)
	}

	if dbs == nil {
		dbs = []models.DatabaseStat{}
	}

	return dbs, nil
}

func (s *MonitoringService) GetTableStats(ctx context.Context, schema string, limit int) ([]models.TableStat, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	query := `
		SELECT
			schemaname,
			tablename,
			seq_scan,
			seq_tup_read,
			idx_scan,
			idx_tup_fetch,
			n_tup_ins,
			n_tup_upd,
			n_tup_del,
			n_live_tup,
			n_dead_tup
		FROM pg_stat_user_tables
		WHERE ($1 = '' OR schemaname = $1)
		ORDER BY (seq_scan + idx_scan) DESC
		LIMIT $2`

	rows, err := s.db.Pool.Query(ctx, query, schema, limit)
	if err != nil {
		return nil, fmt.Errorf("get table stats: %w", err)
	}
	defer rows.Close()

	var tables []models.TableStat
	for rows.Next() {
		var t models.TableStat
		var liveTup, deadTup *int64
		if err := rows.Scan(&t.Schemaname, &t.Tablename, &t.SeqScan, &t.SeqTupRead,
			&t.IdxScan, &t.IdxTupFetch, &t.NTupIns, &t.NTupUpd, &t.NTupDel,
			&liveTup, &deadTup); err != nil {
			return nil, fmt.Errorf("scan table stat: %w", err)
		}
		if liveTup != nil && *liveTup > 0 {
			t.LiveTup = liveTup
		}
		if deadTup != nil && *deadTup > 0 {
			t.DeadTup = deadTup
		}
		tables = append(tables, t)
	}

	if tables == nil {
		tables = []models.TableStat{}
	}

	return tables, nil
}

func (s *MonitoringService) GetIndexStats(ctx context.Context, schema string, limit int) ([]models.IndexStat, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	query := `
		SELECT
			schemaname,
			indexrelname,
			relname,
			idx_scan,
			idx_tup_read,
			idx_tup_fetch
		FROM pg_stat_user_indexes
		WHERE ($1 = '' OR schemaname = $1)
		ORDER BY idx_scan DESC
		LIMIT $2`

	rows, err := s.db.Pool.Query(ctx, query, schema, limit)
	if err != nil {
		return nil, fmt.Errorf("get index stats: %w", err)
	}
	defer rows.Close()

	var indexes []models.IndexStat
	for rows.Next() {
		var idx models.IndexStat
		if err := rows.Scan(&idx.Schemaname, &idx.Indexname, &idx.Tablename,
			&idx.IdxScan, &idx.IdxTupRead, &idx.IdxTupFetch); err != nil {
			return nil, fmt.Errorf("scan index stat: %w", err)
		}
		indexes = append(indexes, idx)
	}

	if indexes == nil {
		indexes = []models.IndexStat{}
	}

	return indexes, nil
}

func (s *MonitoringService) TerminateSession(ctx context.Context, pid int32) (*models.TerminateResponse, error) {
	var result bool
	err := s.db.Pool.QueryRow(ctx, `SELECT pg_terminate_backend($1)`, pid).Scan(&result)
	if err != nil {
		return &models.TerminateResponse{
			PID:     pid,
			Success: false,
			Message: err.Error(),
		}, nil
	}

	msg := "Session terminated successfully"
	if !result {
		msg = "No session found with the given PID"
	}

	return &models.TerminateResponse{
		PID:     pid,
		Success: result,
		Message: msg,
	}, nil
}

func (s *MonitoringService) CancelQuery(ctx context.Context, pid int32) (*models.TerminateResponse, error) {
	var result bool
	err := s.db.Pool.QueryRow(ctx, `SELECT pg_cancel_backend($1)`, pid).Scan(&result)
	if err != nil {
		return &models.TerminateResponse{
			PID:     pid,
			Success: false,
			Message: err.Error(),
		}, nil
	}

	msg := "Query cancelled successfully"
	if !result {
		msg = "No active query found for the given PID"
	}

	return &models.TerminateResponse{
		PID:     pid,
		Success: result,
		Message: msg,
	}, nil
}
