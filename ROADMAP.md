# Nexbic Platform — Roadmap

**Current version:** 0.1.0 (dashboard) / 1.0.0 (API)

---

## Phase 1: Foundation ✅ (Completed)

The core PostgreSQL administration dashboard is fully built and functional.

- User authentication & role-based access control
- Full CRUD for schemas, tables, columns, constraints, indexes, sequences
- Table data browser with pagination, sorting, search, row-level CRUD
- SQL workspace with EXPLAIN, history, saved queries
- Schema explorer (tables, views, functions, procedures, triggers, indexes, constraints, sequences, materialized views, extensions)
- PostgreSQL role management with privilege grants/revokes
- Extension installer
- Database monitoring (sessions, slow queries, locks, cache, stats)
- Backup/restore via pg_dump/pg_restore
- PostgreSQL log viewer with filtering
- Audit trail for admin actions
- SvelteKit SPA served from Go binary
- Docker Compose with PostgreSQL
- CI pipeline (build, lint, test)

---

## Phase 2: Hardening & Quality 🔄 (In Progress)

Making the project production-ready through testing, docs, and observability.

### 2.1 — Documentation Refresh
- [ ] Rewrite README.md to match current PostgreSQL admin dashboard (remove old BaaS architecture references)
- [ ] Regenerate OpenAPI spec (`docs/openapi.json`) to match actual `/v1` endpoints
- [ ] Consolidate `/audit-logs` and `/audit` frontend routes into one

### 2.2 — Testing
- [ ] Go unit tests for all service layers
- [ ] Go integration tests against real PostgreSQL (via CI)
- [ ] Frontend component tests (SvelteKit)
- [ ] Establish minimum coverage threshold

### 2.3 — Observability
- [ ] Expose Prometheus metrics at `/metrics`
- [ ] Wire OpenTelemetry tracing (OTLP exporter)
- [ ] Add structured request logging with request IDs

### 2.4 — Infrastructure
- [ ] Docker Compose monitoring stack (Prometheus + Grafana + Loki)
- [ ] Redis service for distributed rate limiting
- [ ] Health check improvements (detailed component status)

---

## Phase 3: Production Deployment 🚀 (Planned)

Preparing for multi-instance deployments and operational tooling.

- [ ] Database migration versioning tool (replace raw SQL apply)
- [ ] Distributed rate limiting (Redis-backed)
- [ ] Server-side pagination for all list endpoints
- [ ] Backup scheduling (cron-based)
- [ ] Alerting rules (connection spikes, slow query thresholds)
- [ ] Kubernetes manifests (Helm chart)
- [ ] Session management UI improvements
- [ ] Rate limit headers (`X-RateLimit-*`)

---

## Phase 4: Advanced Features 💡 (Future)

Enhancements beyond the core admin dashboard.

- [ ] Multi-tenant app provisioning (original BaaS concept)
- [ ] PostgREST auto-generated REST APIs per schema
- [ ] Custom domain management with auto-HTTPS (Caddy)
- [ ] Visual query builder
- [ ] Table relationship diagram / ERD viewer
- [ ] CSV/JSON data import
- [ ] WebSocket-based live query notifications
- [ ] API key management for programmatic access
- [ ] Audit log retention policies
- [ ] Role-based dashboard customization

---

## Milestone Timeline

| Milestone | Target | Status |
|-----------|--------|--------|
| v0.1.0 — Core Dashboard MVP | Completed | ✅ |
| v0.2.0 — Testing + Docs Refresh | Next | 🔄 |
| v0.3.0 — Observability + Infrastructure | Near-term | ⬜ |
| v1.0.0 — Production Ready | Future | ⬜ |
| v2.0.0 — Advanced Features | Future | ⬜ |
