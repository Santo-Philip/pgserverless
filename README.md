# Nexbic Platform

Multi-tenant PostgreSQL management platform with a public REST API and an internal admin dashboard.

## Architecture

```
┌──────────────────────────────────────────────┐
│  cmd/api/  (Public API on :2121)             │
│  ┌──────┐ ┌────────┐ ┌──────┐ ┌───────┐    │
│  │ Auth │ │Projects│ │Wallet│ │ Files │    │
│  └──────┘ └────────┘ └──────┘ └───────┘    │
└──────────────┬───────────────────────────────┘
               │
┌──────────────┴───────────────────────────────┐
│  cmd/dashboard/  (Dashboard on :2122)        │
│  ┌──────┐ ┌──────────┐ ┌─────┐ ┌────────┐  │
│  │ Auth │ │ Explorer │ │ SQL │ │ Schema │  │
│  │ Audit│ │ Dashboard│ │……   │ │ Storage │  │
│  └──────┘ └──────────┘ └─────┘ └────────┘  │
│  ┌─────────────────────────────────────────┐ │
│  │  SvelteKit Frontend                     │ │
│  └─────────────────────────────────────────┘ │
└──────────────┬───────────────────────────────┘
               │
               ▼
        ┌──────────────┐
        │  PostgreSQL   │
        └──────────────┘
```

## Components

| Component | Description | Port (default) |
|-----------|-------------|---------------|
| **API** | Public-facing REST API — auth, projects, wallet, file uploads | 2121 |
| **Dashboard** | Internal admin API + SvelteKit dashboard UI | 2122 |
| **Docs** | Static API documentation served at `/docs/` | (mounted on API server) |

## Project Structure

```
cmd/
├── api/          — Public API server entrypoint
├── dashboard/    — Dashboard server entrypoint
└── server/       — Legacy monolith entrypoint (preserved)

internal/
├── app/          — Shared initialization for both servers
├── audit/        — Audit log module (dashboard)
├── auth/         — Authentication & authorization
├── backups/      — Database backup management
├── dashboard/    — Dashboard aggregation queries
├── explorer/     — Schema browsing (tables, views, routines)
├── extensions/   — PostgreSQL extension management
├── files/        — Public file upload/download (API)
├── logs/         — Query log viewer
├── middleware/    — Auth, CORS, rate-limit, request ID, audit
├── monitoring/   — Database health & metrics
├── pgrest/       — Auto-generated REST API per schema/table
├── pgroles/      — PostgreSQL role management
├── projects/     — Project CRUD (meta API)
├── schema/       — Schema diff & migration tooling
├── sql/          — SQL query executor
├── storage/      — Internal storage providers/buckets/files
├── tables/       — Table metadata browsing
└── wallet/       — User credit/debit wallet

migrations/
├── 001-schema.sql
├── 002-seed.sql
├── 003-auth-storage.sql
├── 004-projects.sql
└── 005-wallet.sql

pkg/
├── database/     — PostgreSQL connection pool
├── helpers/      — UUID parsing, pagination, etc.
└── response/     — JSON response helpers
```

## API Endpoints

All endpoints are prefixed with `/v1`.

### Public API (`cmd/api/`)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/auth/register` | Register a new user |
| POST | `/auth/login` | Login |
| POST | `/auth/refresh` | Refresh JWT |
| POST | `/auth/forgot-password` | Request password reset |
| POST | `/auth/reset-password` | Reset password |
| POST | `/auth/verify-email` | Verify email with token |
| GET | `/auth/me` | Current user profile |
| PATCH | `/auth/password` | Change password |
| POST | `/auth/verify-email/send` | Send verification email |
| POST | `/auth/totp/enable` | Enable 2FA TOTP |
| POST | `/auth/totp/verify` | Verify TOTP setup |
| POST | `/auth/totp/disable` | Disable TOTP |
| GET | `/auth/devices` | List trusted devices |
| DELETE | `/auth/devices/:id` | Remove device |
| GET | `/auth/security-events` | List security events |
| GET | `/auth/api-keys` | List API keys |
| POST | `/auth/api-keys` | Create API key |
| DELETE | `/auth/api-keys/:id` | Revoke API key |
| GET | `/admin/users` | List users (super_admin) |
| GET | `/admin/users/:id` | Get user (super_admin) |
| POST | `/admin/users` | Create user (super_admin) |
| PATCH | `/admin/users/:id` | Update user (super_admin) |
| PATCH | `/admin/users/:id/password` | Update user password (super_admin) |
| DELETE | `/admin/users/:id` | Delete user (super_admin) |
| GET | `/projects` | List projects |
| POST | `/projects` | Create project |
| GET | `/projects/:projectId` | Get project |
| PATCH | `/projects/:projectId` | Update project |
| DELETE | `/projects/:projectId` | Delete project |
| GET | `/wallet/balance` | Get wallet balance |
| POST | `/wallet/transactions` | Create credit/debit transaction |
| GET | `/wallet/transactions` | List transactions |
| GET | `/files` | List user files |
| POST | `/files/upload` | Upload a file |
| GET | `/files/:id/download` | Download a file |
| DELETE | `/files/:id` | Delete a file |

### Dashboard API (`cmd/dashboard/`)

Same auth endpoints (minus public registration / password reset), plus:

| Method | Path | Description |
|--------|------|-------------|
| GET | `/audit` | List audit logs |
| POST | `/projects/:projectId/dashboard/stats` | Dashboard aggregate stats |
| GET | `/projects/:projectId/explorer/schemas` | List schemas |
| GET | `/projects/:projectId/explorer/tables` | List tables |
| GET | `/projects/:projectId/explorer/views` | List views |
| GET | `/projects/:projectId/explorer/routines` | List routines |
| GET | `/projects/:projectId/explorer/extensions` | List installed extensions |
| POST | `/projects/:projectId/sql/query` | Execute SQL query |
| POST | `/projects/:projectId/schema/diff` | Schema diff |
| POST | `/projects/:projectId/schema/migrate` | Apply migration |
| GET/POST | `/projects/:projectId/pg-roles/*` | PostgreSQL role management |
| GET/POST | `/projects/:projectId/extensions/*` | Extension management |
| GET | `/projects/:projectId/monitoring/*` | Database health & metrics |
| GET/POST/DELETE | `/projects/:projectId/backups/*` | Backup & restore |
| GET | `/projects/:projectId/logs` | Query execution logs |
| GET/POST/PATCH/DELETE | `/projects/:projectId/storage/*` | Storage providers, buckets, files |

### Auto-generated REST API (`pgREST`)

Under `/v1/projects/:projectId/r/:schema/:table`, the platform automatically exposes CRUD endpoints for every table discovered in the project's dedicated schema, supporting filtering (`eq`, `neq`, `gt`, `gte`, `lt`, `lte`, `like`, `ilike`, `in`, `is`, `isnot`), ordering, pagination, and selecting specific columns.

## Quick Start

### 1. Environment

```bash
cp .env.example .env
```

### 2. Database

```bash
for f in migrations/*.sql; do
  psql "$DATABASE_URL" -f "$f"
done
```

### 3. Run

```bash
# Terminal 1: Public API
go run ./cmd/api

# Terminal 2: Dashboard (internal + frontend)
go run ./cmd/dashboard
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | — | PostgreSQL connection string |
| `JWT_SECRET` | — | HMAC-SHA256 JWT signing key |
| `JWT_ACCESS_TTL` | `15m` | Access token lifetime |
| `JWT_REFRESH_TTL` | `168h` | Refresh token lifetime |
| `SERVER_PORT` | `2121` | HTTP listen port |
| `SUPER_ADMIN_EMAIL` | — | Auto-seeded super admin email |
| `SUPER_ADMIN_PASSWORD` | — | Super admin password |
| `CORS_ORIGINS` | `http://localhost:5173` | Allowed origins |
| `FILES_DIR` | `/data/user_files` | Public file storage root |
| `BACKUP_DIR` | `/data/backups` | Database backup directory |

## Security

- Passwords hashed with Argon2id
- JWT tokens with configurable TTL
- API keys stored as SHA-256 hashes
- Rate limiting per-IP (200 req/min)
- CORS with explicit origin allowlist
- Audit logging for all admin actions (dashboard)
- Project-scoped access via project ownership or super_admin role

## Testing

```bash
go test ./...
go vet ./...
```

## License

Proprietary. All rights reserved.
