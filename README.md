# Nexbic Platform

Multi-tenant backend-as-a-service platform that provisions isolated PostgreSQL schemas per tenant ("app"), provides an auto-generated REST API via PostgREST, and offers a management API for CRUD on apps, users, API keys, domains, database tables, and PostgreSQL extensions.

## Architecture

```
┌──────────┐  ┌───────────┐  ┌──────────────┐
│ Dashboard│  │  Caddy    │  │  PostgREST   │
│ (Svelte) │  │ (Reverse  │  │ (per-tenant  │
│          │  │  Proxy)   │  │  REST API)   │
└────┬─────┘  └─────┬─────┘  └──────┬───────┘
     │              │               │
     │     ┌────────┴────────┐      │
     │     │    Gateway      │      │
     │     │  (Fiber Proxy)  │      │
     │     └────────┬────────┘      │
     │              │               │
     │     ┌────────┴────────┐      │
     └─────┤ Management API  ├──────┘
           │  (Fiber CRUD)   │
           └────────┬────────┘
                    │
           ┌────────┴────────┐
           │    Worker       │
           │  (Asynq Tasks)  │
           └────────┬────────┘
                    │
           ┌────────┴────────┐
           │   PostgreSQL    │
           │  + PgBouncer    │
           └─────────────────┘
```

### Components

| Component | Description | Port |
|-----------|-------------|------|
| **Gateway** | Fiber-based reverse proxy. Routes API requests to PostgREST per tenant. Handles host-based and slug-based routing. | 8080 |
| **Management API** | CRUD REST API for managing tenants (apps), users, API keys, domains, DB schemas/tables, extensions, platform settings. | 8081 |
| **Worker** | Asynq-based background task processor. Handles PostgREST schema cache refreshes after app creation. | - |
| **Dashboard** | SvelteKit SPA for managing the platform via the Management API. | 5173 |
| **PostgREST** | Auto-generates REST APIs per PostgreSQL schema. Each app gets its own schema/role. | 3000 / 3001 |
| **PgBouncer** | Connection pooler for PostgreSQL. | 6432 |
| **Caddy** | Reverse proxy with automatic HTTPS, security headers, subdomain routing. | 2121 / 8443 |

### Infrastructure

| Service | Purpose |
|---------|---------|
| PostgreSQL 17 | Primary database with per-tenant schemas |
| Redis 7 | Caching, distributed rate limiting, task queue backend |
| Prometheus | Metrics collection (exposed at `/metrics`) |
| Loki | Log aggregation |
| Grafana | Metrics/logs visualization |

## Prerequisites

- Go 1.25+
- Node.js 22+
- Docker & Docker Compose (for full stack)
- PostgreSQL 17 (for local dev)

## Quick Start

### 1. Environment

```bash
cp .env.example .env
# Edit .env with your settings
```

### 2. Database

```bash
# Apply migrations in order:
psql -U api_admin -d postgres_api -f postgres/init/001-schema.sql
psql -U api_admin -d postgres_api -f postgres/init/002-functions.sql
psql -U api_admin -d postgres_api -f postgres/init/003-roles.sql
psql -U api_admin -d postgres_api -f postgres/init/004-platform.sql
psql -U api_admin -d postgres_api -f postgres/init/005-domains.sql
psql -U api_admin -d postgres_api -f postgres/init/006-super-admin.sql
psql -U api_admin -d postgres_api -f postgres/init/007-extensions.sql
```

### 3. Run Services

```bash
# Terminal 1: Management API
go run ./management-api

# Terminal 2: Gateway
go run ./gateway

# Terminal 3: Worker
go run ./worker

# Terminal 4: Dashboard
cd dashboard && npm run dev
```

### 4. Docker Compose (Full Stack)

```bash
docker compose up -d
```

## API Documentation

Full OpenAPI spec: `docs/openapi.json`

### Management API

Base URL: `http://localhost:8081/api/v1/platform`

| Method | Path | Description |
|--------|------|-------------|
| POST | `/auth/register` | Register new admin user |
| POST | `/auth/login` | Login |
| POST | `/auth/refresh` | Refresh JWT token |
| GET | `/me` | Current user info |
| POST | `/apps` | Create app (tenant) |
| GET | `/apps` | List apps |
| GET | `/apps/:id` | Get app details |
| DELETE | `/apps/:id` | Delete app |
| POST | `/apps/:id/apikey` | Create API key |
| GET | `/apps/:id/apikey` | List API keys |
| DELETE | `/apps/:id/apikey/:keyId` | Deactivate API key |
| GET/POST/DELETE | `/apps/:id/domains` | Manage custom domains |
| POST | `/apps/:id/domains/:domainId/verify` | Verify domain ownership |
| GET/POST | `/apps/:id/extensions` | Manage PG extensions |
| GET/POST | `/apps/:id/tables` | Manage DB tables |
| GET | `/apps/:id/tables/:table` | Query table data |
| POST/PATCH/DELETE | `/apps/:id/tables/:table/rows` | CRUD rows |
| POST | `/apps/:id/tables/:table/columns` | Add column |
| GET/PATCH | `/settings` | Platform settings |
| GET/POST | `/users` | Manage users |
| GET | `/users/:userId` | Get user |
| POST | `/users/:userId/suspend` | Suspend user |
| POST | `/users/:userId/activate` | Activate user |

## Monitoring

- **Metrics**: All Go services expose Prometheus metrics at `/metrics`
- **Health**: All Go services expose `/health` (liveness) and `/ready` (readiness)
- **Logs**: Structured JSON logging via `log/slog`
- **Tracing**: OpenTelemetry integration (configurable OTLP exporter)

### Grafana Dashboards

When running with Docker Compose, Grafana is available at `http://localhost:3000` (default: admin/admin).

Pre-configured datasources:
- Prometheus (metrics)
- Loki (logs)

## Production Deployment

### Requirements

- PostgreSQL 17+ with replication for HA
- Redis 7+ with sentinel/cluster for HA
- At least 2 gateway instances behind a load balancer
- Caddy or other reverse proxy for TLS termination

### Environment Variables

See `.env.example` for all configurable variables. Key ones:

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | - | PostgreSQL connection string |
| `REDIS_URL` | - | Redis connection string |
| `JWT_SECRET` | `change-me` | HMAC-SHA256 signing key |
| `CORS_ORIGINS` | `http://localhost:5173` | Allowed CORS origins |
| `MONITORING_ENABLED` | `true` | Enable Prometheus metrics |
| `OTLP_ENDPOINT` | - | OpenTelemetry collector endpoint |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |

### Scalability

- **Gateway**: Stateless, scales horizontally. Requires shared Redis for distributed rate limiting.
- **Management API**: Stateless, scales horizontally.
- **Worker**: Single-active pattern recommended for task ordering.

## Security

- Passwords hashed with Argon2id (64MB memory, 3 iterations)
- JWT tokens with configurable TTL (15m access, 7d refresh)
- API keys stored as SHA-256 hashes
- Per-app PostgreSQL roles with scoped schema privileges
- Rate limiting (per-IP, distributed via Redis)
- CORS with explicit origin allowlist
- Audit logging for admin actions
- All containers run with `no-new-privileges:true`

## Testing

```bash
# Run all tests
go test ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Dashboard tests
cd dashboard && npm run test
```

## CI/CD

The project uses GitHub Actions (`.github/workflows/ci.yml`):
- Backend: `go build`, `go vet`, `go test` against a real PostgreSQL
- Dashboard: `npm ci`, `npm run build`

## License

Proprietary. All rights reserved.
