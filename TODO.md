# Nexbic Platform — TODO

## Legend
- ✅ Done
- 🔄 In Progress
- ⬜ Not Started

---

## Core Features

### Authentication & User Management
- ✅ JWT-based login with access + refresh tokens
- ✅ Argon2id password hashing
- ✅ Role-based access control (super_admin, dba, developer, read_only)
- ✅ User CRUD (admin panel)
- ✅ User profile, sessions, login history
- ✅ Password change / reset

### Dashboard
- ✅ Database overview stats (version, size, connections, uptime, cache ratio, TPS)
- ✅ Schema list with table counts
- ✅ Replication status indicator

### Schema Explorer
- ✅ Schema listing
- ✅ Tables, Views, Functions, Procedures, Triggers, Indexes, Constraints, Sequences, Materialized Views
- ✅ Extension listing
- ✅ Table detail navigation

### Table Data Management
- ✅ Browse rows with pagination, sorting, search
- ✅ Insert / Edit / Delete rows
- ✅ Bulk insert / delete

### SQL Workspace
- ✅ Execute arbitrary SQL
- ✅ EXPLAIN (ANALYZE, BUFFERS, JSON)
- ✅ Cancel running query
- ✅ Query history
- ✅ Saved queries
- ✅ CSV export

### Schema Management (DDL)
- ✅ Create / Drop schema
- ✅ Create / Drop table
- ✅ Add / Alter / Drop column
- ✅ Add / Drop constraint
- ✅ Create / Drop index
- ✅ Create / Drop / Alter sequence
- ✅ Get table DDL

### PostgreSQL Role Management
- ✅ List / Create / Alter / Drop roles
- ✅ Set role password
- ✅ Grant / Revoke privileges (database, schema, table)
- ✅ Role membership management
- ✅ Superuser badge, role attributes

### Extension Management
- ✅ List installed extensions
- ✅ List available extensions
- ✅ Install / Uninstall

### Database Monitoring
- ✅ Active sessions (with terminate)
- ✅ Slow queries (with cancel)
- ✅ Active locks
- ✅ Waiting queries
- ✅ Query statistics
- ✅ Connection statistics
- ✅ Cache statistics
- ✅ Database statistics
- ✅ Table statistics
- ✅ Index statistics

### Backup Management
- ✅ Create backup (pg_dump)
- ✅ List / Get / Delete backups
- ✅ Restore backup (pg_restore)
- ✅ Verify backup
- ✅ Download backup

### Log Viewer
- ✅ PostgreSQL log parsing (CSV/stderr)
- ✅ Query logs
- ✅ Error logs
- ✅ Auth logs
- ✅ Connection logs
- ✅ Filter by severity, database, user, date range

### Audit Logging
- ✅ Admin action audit trail
- ✅ Audit log viewer with filters, date grouping, expandable details
- ⬜ `/audit-logs` route — simpler version, merge or remove in favor of `/audit`

### API Documentation
- ✅ Root endpoint listing all routes
- ⬜ OpenAPI spec — outdated (still references old BaaS endpoints)

---

## Infrastructure & Deployment

### Build & CI
- ✅ Go backend build
- ✅ SvelteKit frontend build
- ✅ Multi-stage Docker build
- ✅ Docker Compose (PostgreSQL + dashboard)
- ✅ GitHub Actions CI (Go build+test+lint, Node build)

### Configuration
- ✅ Environment-based config via `.env`
- ✅ Config struct with sensible defaults

### Security
- ✅ Argon2id password hashing
- ✅ JWT short-lived access tokens (15m) + long-lived refresh (7d)
- ✅ Rate limiting (in-memory, per-IP, 200 req/min)
- ✅ CORS with explicit origin allowlist
- ✅ API key hashing (SHA-256) — note: API keys are from old BaaS, not currently used

### Observability
- ✅ Structured JSON logging (slog)
- ✅ Health (`/health`) and readiness (`/ready`) endpoints
- ✅ Graceful shutdown
- ⬜ Prometheus metrics endpoint (`/metrics`)
- ⬜ OpenTelemetry tracing integration

---

## Testing
- ⬜ Go unit tests
- ⬜ Go integration tests (against real PG)
- ⬜ Frontend component / E2E tests
- ⬜ Tests directory is empty

---

## Known Issues & Technical Debt

### Documentation
- ⬜ **README.md** — completely outdated. Still describes old multi-tenant BaaS architecture (Gateway, PostgREST, Worker, Caddy, PgBouncer). Needs rewrite to match current PostgreSQL admin dashboard reality.
- ⬜ **docs/openapi.json** — outdated OpenAPI spec referencing old BaaS endpoints (`/apps`, `/domains`, `/apikey`). Needs to be regenerated.

### Frontend
- ⬜ `/audit-logs` route is a minimal version of `/audit` — consolidate into one
- ⬜ No search/filter on table browse page for individual columns (only global search)

### Backend
- ⬜ Rate limiter is in-memory only — not suitable for multi-instance deployments
- ⬜ No database migration versioning (migrations are raw SQL files applied manually)
- ⬜ Backup path configurable but no validation on directory existence
- ⬜ No pagination on audit logs endpoint server-side

### Infrastructure
- ⬜ Monitoring stack (Prometheus, Grafana, Loki) referenced in README but not in docker-compose
- ⬜ No Redis service in docker-compose (needed for distributed rate limiting)
- ⬜ `docker/` directory is empty
- ⬜ No production deployment manifests (Kubernetes, etc.)

---

## Future / Nice-to-Have

- ⬜ Multi-tenant app provisioning (original BaaS concept)
- ⬜ PostgREST auto-generated REST APIs per schema
- ⬜ Custom domain management with auto-HTTPS (Caddy)
- ⬜ WebSocket-based live query notifications
- ⬜ Visual query builder
- ⬜ Table relationship diagram / ERD view
- ⬜ Import data from CSV/JSON
- ⬜ Scheduled backup cron jobs
- ⬜ Alerting rules (e.g., connection threshold, slow query threshold)
- ⬜ Dark/light theme fully consistent (theme switcher exists, check consistency)
