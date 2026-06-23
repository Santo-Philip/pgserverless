# Postgres API Stack

Expose PostgreSQL entirely through a secure REST API. No direct database connections required by clients.

## Quick Start

```bash
cp .env.example .env
# Edit .env with secure passwords
docker compose up -d
```

Your API is live at `http://localhost:8080`.

```bash
curl http://localhost:8080/health
curl http://localhost:8080/
```

## Architecture

```
Client (Worker/App/Frontend)
        |
        | HTTP/HTTPS :8080
        v
    +--------+
    |  Nginx |  Reverse proxy, rate limiting, security headers, compression
    +--------+
        |
        | HTTP :3000 (internal)
        v
  +----------+
  | PostgREST|  REST-to-SQL translation, JWT auth, OpenAPI
  +----------+
        |
        | TCP :5432 (internal)
        v
  +----------+
  | PostgreSQL|  Persistent storage, extensions, roles
  +----------+
```

## Services

| Service | Image | Port | Description |
|---------|-------|------|-------------|
| PostgreSQL | postgres:17-alpine | internal | Database with UUID, pgcrypto, UTC |
| PostgREST | postgrest/postgrest | internal | REST API layer over PostgreSQL |
| Nginx | nginx:1.27-alpine | 8080 | Reverse proxy with security features |

Only port 8080 is publicly exposed. Everything else runs on an internal Docker network.

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `POSTGRES_DB` | `postgres_api` | Database name |
| `POSTGRES_USER` | `api_admin` | Database user |
| `POSTGRES_PASSWORD` | *(required)* | Database password |
| `JWT_SECRET` | *(required)* | JWT signing secret (32+ chars) |
| `JWT_AUD` | `postgres-api-stack` | JWT audience claim |
| `API_PORT` | `8080` | Public API port |

### Production Security Checklist

- [ ] Change all default passwords in `.env`
- [ ] Generate a strong JWT secret: `openssl rand -base64 48`
- [ ] Set `cors-origin` in `postgrest/postgrest.conf` to your domain
- [ ] Configure HTTPS (see below)
- [ ] Enable rate limiting in `nginx/nginx.conf`
- [ ] Restrict `API_PORT` to your Cloudflare IP range via firewall

## Database Schema

### Tables

| Table | Description |
|-------|-------------|
| `users` | User accounts with email, password hash, profile |
| `organizations` | Organizations/tenants |
| `roles` | Role definitions (anon, authenticated, admin) |
| `permissions` | Resource-action permission definitions |
| `role_permissions` | Many-to-many role-permission assignments |
| `sessions` | User sessions with tokens and expiry |
| `api_keys` | API key management with scoped permissions |
| `logs` | API request audit log |

### Extensions

- `uuid-ossp` - UUID generation
- `pgcrypto` - Password hashing, encryption functions

### Indexes

All tables have indexes on foreign keys, status fields, and frequently queried columns.

## Authentication

### JWT Flow

1. **Register** - `POST /rpc/register` creates a user account
2. **Login** - `POST /rpc/login` validates credentials, returns JWT
3. **Authenticate** - Pass JWT in `Authorization: Bearer <token>` header
4. **PostgREST** verifies JWT and applies role-based permissions

### API Key Authentication

1. Generate an API key via `POST /rpc/generate_api_key`
2. Pass it in the `X-API-Key` header
3. Keys can be scoped to specific permissions

### Roles

| Role | Access | Description |
|------|--------|-------------|
| `anon` | Read-only | Public access, no auth required |
| `authenticated` | Read + own write | Logged-in users |
| `admin` | Full access | All CRUD operations |

## REST API

### CRUD Operations

```bash
# List (GET)
GET /users
GET /users?limit=10&offset=0
GET /users?order=created_at.desc

# Create (POST)
POST /users
Content-Type: application/json
{"email": "user@example.com", "name": "John"}

# Update (PATCH)
PATCH /users?id=eq.1
Content-Type: application/json
{"name": "Jane"}

# Delete (DELETE)
DELETE /users?id=eq.1
```

### Filtering

```bash
# Equality
GET /users?status=eq.active

# Negation
GET /users?status=neq.suspended

# Greater/Less than
GET /users?created_at=gt.2024-01-01

# In list
GET /users?status=in.(active,inactive)

# Like
GET /users?email=like.*@example.com

# IS NULL
GET /users?organization_id=is.null
```

### Pagination

```bash
# Using limit/offset
GET /users?limit=25&offset=0

# Using Range header
GET /users
Range: 0-24
# Response includes Content-Range: 0-24/100
```

### Ordering

```bash
GET /users?order=name.asc
GET /users?order=created_at.desc
GET /users?order=organization_id.desc,email.asc
```

### Counting

```bash
GET /users?select=count
GET /users?select=count&status=eq.active
```

### Embedded Relationships

```bash
# Include related role
GET /users?select=id,email,name,role:role_id(id,name)

# Include organization
GET /users?select=id,email,name,organization:organization_id(id,name,slug)

# Nested embedding
GET /organizations?select=id,name,users:users(id,email,role:role_id(name))
```

### RPC Functions

| Function | Method | Auth | Description |
|----------|--------|------|-------------|
| `/rpc/login` | POST | No | Authenticate user |
| `/rpc/register` | POST | No | Create account |
| `/rpc/change_password` | POST | Yes | Change password |
| `/rpc/healthcheck` | GET | No | API health status |
| `/rpc/get_stats` | GET | Admin | System statistics |
| `/rpc/get_current_user` | POST | Yes | Current user profile |
| `/rpc/generate_api_key` | POST | Yes | Create API key |
| `/rpc/validate_api_key` | POST | No | Validate API key |

## Examples

See the `examples/` directory:

- [cURL](examples/curl.md) - Command-line examples
- [JavaScript](examples/javascript.md) - Fetch and Axios
- [Go](examples/golang.md) - Go client library
- [Cloudflare Workers](examples/cloudflare-workers.md) - Worker integration

## Adding New Tables

1. Create a migration in `postgres/init/` (e.g., `004-orders.sql`)

```sql
-- 004-orders.sql
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    total DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_user ON orders(user_id);

-- Grant access
GRANT ALL ON orders TO admin;
GRANT SELECT, INSERT ON orders TO authenticated;
GRANT SELECT ON orders TO anon;
```

2. Restart PostgREST: `docker compose restart postgrest`

The new table is immediately available at `GET /orders`, `POST /orders`, etc.

## Permissions

### Adding a Permission

```sql
INSERT INTO permissions (resource, action) VALUES ('orders', 'read');
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'authenticated' AND p.resource = 'orders' AND p.action = 'read';
```

### Row-Level Security (PostgREST)

PostgREST uses PostgreSQL row-level security when configured. Example:

```sql
ALTER TABLE orders ENABLE ROW LEVEL SECURITY;

CREATE POLICY user_orders ON orders
    FOR ALL
    USING (user_id = current_setting('request.jwt.claim.sub')::UUID);
```

## Backups

### Manual Backup

```bash
chmod +x scripts/backup.sh
./scripts/backup.sh
```

Backups are stored in `backups/` as compressed SQL files. Old backups are automatically cleaned after 30 days.

### Manual Restore

```bash
./scripts/restore.sh backups/postgres_api_20240101_120000.sql.gz
```

### Scheduled Backups (Cron)

```bash
# Run daily at 2 AM
0 2 * * * /path/to/postgres-api-stack/scripts/backup.sh
```

## Production Deployment

### HTTPS with Let's Encrypt

```nginx
# nginx/nginx.conf
server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/api.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.yourdomain.com/privkey.pem;

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # ... rest of config
}
```

### Cloudflare Integration

1. Set DNS to proxy through Cloudflare (orange cloud)
2. Configure SSL/TLS to Full (strict)
3. Create a Cloudflare API Token with Zone.DNS access
4. Add a Worker using examples from `examples/cloudflare-workers.md`

### Docker Resource Limits

```yaml
# docker-compose.yml
services:
  postgres:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '0.5'
          memory: 512M
```

## Extending into a Supabase-Like Backend

This stack is designed to be the foundation of a Supabase-like platform:

- **Realtime** - Add `supabase/realtime` container for WebSocket subscriptions
- **Storage** - Add `minio` or `supabase/storage-api` for S3-compatible file storage
- **Auth** - Extend JWT with custom claims, add OAuth providers
- **Edge Functions** - Add Deno/Kong for serverless function execution
- **GraphQL** - Add `pg_graphql` extension or Hasura
- **Search** - Add `typesense` or `meilisearch` for full-text search
- **Queue** - Add `graphile-worker` for background job processing

## Development

### Applying Schema Changes

1. Add migration files to `postgres/init/`
2. Rebuild the database container: `docker compose up -d --force-recreate postgres`
3. For existing data, use `psql` directly (requires exec into container):

```bash
docker compose exec postgres psql -U api_admin -d postgres_api -f /path/to/migration.sql
```

### Viewing Logs

```bash
docker compose logs -f nginx
docker compose logs -f postgrest
docker compose logs -f postgres
```

### Restarting Services

```bash
docker compose restart postgrest   # Apply config changes
docker compose restart nginx       # Apply nginx changes
```

## Testing

```bash
curl -s http://localhost:8080/rpc/healthcheck | jq .
```

## Project Structure

```
postgres-api-stack/
├── docker-compose.yml           # Service orchestration
├── .env.example                 # Environment template
├── .gitignore                   # Git ignore rules
├── README.md                    # This file
├── postgres/
│   ├── init/                    # DB initialization scripts
│   │   ├── 001-schema.sql       # Tables, indexes, triggers
│   │   ├── 002-functions.sql    # RPC functions (login, register, etc.)
│   │   └── 003-roles.sql        # Roles, permissions, seed data
│   └── data/                    # Persistent data (gitignored)
├── postgrest/
│   └── postgrest.conf           # PostgREST configuration
├── nginx/
│   └── nginx.conf               # Reverse proxy configuration
├── scripts/
│   ├── backup.sh                # Database backup
│   └── restore.sh               # Database restore
└── examples/
    ├── curl.md                  # cURL examples
    ├── javascript.md            # JavaScript/Fetch/Axios
    ├── golang.md                # Go client
    └── cloudflare-workers.md    # CF Workers integration
```
