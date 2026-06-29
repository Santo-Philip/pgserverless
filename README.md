# Nexbic Core

The identity platform powering every Nexbic application.

Nexbic Core is a production-ready identity and platform backend that provides authentication, session management, wallet, projects, and file storage — backed by a PostgreSQL database.

---

## Architecture

```
account.nexbic.com
erp.nexbic.com
mehub.in
future apps
        │
        ▼

    Nexbic Core

────────────────────────────────────

Public Platform (cmd/api)

  /v1/auth  /v1/users  /v1/sessions
  /v1/wallet  /v1/projects  /v1/files

────────────────────────────────────

Internal Platform (cmd/dashboard)

  Identity Mgmt  |  Database Tooling
  Users          |  SQL Executor
  Sessions       |  Schema Explorer
  Wallet         |  Migrations
  Organizations  |  Monitoring
                 |  Backups

────────────────────────────────────

        PostgreSQL
```

---

## Public Platform

The public REST API provides everything a Nexbic application needs:

| Area | Endpoints |
|------|-----------|
| Auth | Login, Refresh, OAuth (Google, GitHub), Sessions, Devices, API Keys |
| Users | Profile, Password |
| Wallet | Balance, Transactions |
| Projects | CRUD, Members |
| Files | Upload, Download, List, Delete |

No database administration endpoints are exposed publicly.

---

## Internal Platform

The dashboard is an **internal** application for Nexbic development only. It directly invokes Go services — no HTTP requests to the public API.

Features:
- **Identity** — Users, Sessions, Wallet, Organizations
- **Projects** — Create, Delete, Members, Permissions
- **Database** — SQL, Explorer, Schemas, Migrations, Extensions, Roles, Monitoring, Logs
- **System** — Audit, Storage, Backups

---

## Shared Services

Business logic lives once in shared Go services and is consumed by both the public API and the dashboard:

```
internal/identity/auth/     — Authentication & OAuth
internal/identity/wallet/   — Wallet & transactions
internal/identity/oauth/    — OAuth providers
internal/projects/          — Project management
internal/files/             — File storage
internal/database/          — Database tooling
internal/audit/             — Audit logging
internal/middleware/        — Auth, CORS, rate limiting
```

---

## Repository Structure

```
cmd/
  api/            Public API server
  dashboard/      Internal dashboard server
  migrate/        Migration runner
  server/         Legacy monolith (preserved)

internal/
  identity/
    auth/         Authentication, users, sessions, API keys, OAuth
    wallet/       User wallet & transactions
    oauth/        OAuth provider stubs
    users/        User management
    sessions/     Session management
  projects/       Project CRUD
  files/          File upload/download
  dashboard/      Dashboard aggregation queries
  audit/          Audit logging
  middleware/     Auth, CORS, rate-limit, request ID, project guard
  app/            Shared initialization for both servers
  database/
    explorer/     Schema & table browsing
    sql/          SQL query execution
    schema/       Schema diff & migrations
    monitoring/   Database health & metrics
    backups/      Backup & restore
    extensions/   Extension management
    roles/        PostgreSQL role management
    logs/         Query execution logs
    storage/      Storage providers, buckets, files
    tables/       Table metadata

pkg/
  database/       PostgreSQL connection pool
  helpers/        UUID parsing, pagination, etc.
  password/       Argon2id hashing, API key generation
  response/       JSON response helpers
  validator/      Request validation
  totp/           TOTP utilities (for future use)

migrations/       SQL migration files
docs/             API documentation
```

---

## Quick Start

```bash
# 1. Configure environment
cp .env.example .env

# 2. Run migrations
for f in migrations/*.sql; do psql "$DATABASE_URL" -f "$f"; done

# 3. Start servers
go run ./cmd/api          # Public API on :2121
go run ./cmd/dashboard    # Dashboard on :2122
```

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | — | PostgreSQL connection string |
| `JWT_SECRET` | — | HMAC-SHA256 signing key |
| `JWT_ACCESS_TTL` | `15m` | Access token lifetime |
| `JWT_REFRESH_TTL` | `168h` | Refresh token lifetime |
| `SERVER_PORT` | `2121` | HTTP listen port |
| `SUPER_ADMIN_EMAIL` | — | Auto-seeded super admin |
| `SUPER_ADMIN_PASSWORD` | — | Super admin password |
| `OAUTH_GOOGLE_CLIENT_ID` | — | Google OAuth client ID |
| `OAUTH_GITHUB_CLIENT_ID` | — | GitHub OAuth client ID |
| `OAUTH_REDIRECT_URL` | — | OAuth callback URL |
| `FILES_DIR` | `/data/user_files` | File storage root |

---

## Authentication

- **OAuth 2.0** — Google and GitHub sign-in
- **JWT** — Signed access tokens with configurable TTL
- **Refresh Tokens** — Rotated on each use, supports revocation
- **Session Management** — Multiple devices, session listing & revocation
- **API Keys** — SHA-256 hashed, scoped to users
- **Security Events** — Login attempts, password changes, device changes

Email/password, TOTP, and OTP are not implemented in v1. The database schema supports them for future addition.

---

## Session Flow

```
Login → JWT (access_token + refresh_token)
        ↓
Access token in Authorization header
        ↓
Token expires → Use refresh_token
        ↓
Old refresh token revoked → New pair issued
        ↓
Logout → Refresh token deleted
```

---

## Wallet

- Credit/debit transactions scoped to user accounts
- Balance tracking per currency
- Transaction history with pagination
- Atomic balance updates
- No user-to-user transfers in v1

---

## Security

- Passwords hashed with Argon2id
- JWT with HMAC-SHA256
- API keys stored as SHA-256 hashes
- Rate limiting per IP (200 req/min)
- CORS with explicit origin allowlist
- Audit logging for admin actions
- Project-scoped access control
- Refresh token rotation

---

## Testing

```bash
go test ./...
go vet ./...
```

---

## License

Proprietary. All rights reserved.
