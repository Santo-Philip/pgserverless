# cURL Examples

## Prerequisites

```bash
# Copy and modify environment
cp .env.example .env
# Edit .env with your values, then:
docker compose up -d
```

## Health Check

```bash
curl http://localhost:8080/health
```

## OpenAPI Schema

```bash
curl http://localhost:8080/
```

## Anonymous Access (Read-Only)

```bash
# List users (public data)
curl http://localhost:8080/users -H "Accept: application/json"

# List organizations
curl http://localhost:8080/organizations

# List roles
curl http://localhost:8080/roles
```

## Registration & Authentication

```bash
# Register a new user
curl -X POST http://localhost:8080/rpc/register \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{"p_email": "user@example.com", "p_password": "securepass123", "p_name": "John Doe"}'

# Login
curl -X POST http://localhost:8080/rpc/login \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{"p_email": "user@example.com", "p_password": "securepass123"}'
```

Save the token from the login response for authenticated requests.

## Authenticated Requests

```bash
TOKEN="your-jwt-token-here"

# Get current user profile
curl http://localhost:8080/rpc/get_current_user \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{"p_user_id": "your-user-uuid"}'
```

## CRUD Operations

```bash
TOKEN="your-jwt-token-here"

# CREATE - Add an organization
curl -X POST http://localhost:8080/organizations \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{"name": "My Organization", "slug": "my-org", "description": "Our first org"}'

# READ - List with filtering
curl "http://localhost:8080/organizations?status=eq.active" \
  -H "Authorization: Bearer $TOKEN"

# READ - Pagination
curl "http://localhost:8080/users?limit=10&offset=0" \
  -H "Authorization: Bearer $TOKEN"

# READ - Ordering
curl "http://localhost:8080/users?order=created_at.desc" \
  -H "Authorization: Bearer $TOKEN"

# READ - Get single item by ID
curl "http://localhost:8080/users?id=eq.your-user-uuid" \
  -H "Authorization: Bearer $TOKEN"

# READ - Count
curl "http://localhost:8080/users?select=count" \
  -H "Authorization: Bearer $TOKEN"

# UPDATE - Patch
curl -X PATCH "http://localhost:8080/users?id=eq.your-user-uuid" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{"name": "Updated Name"}'

# DELETE
curl -X DELETE "http://localhost:8080/users?id=eq.your-user-uuid" \
  -H "Authorization: Bearer $TOKEN"
```

## API Key Authentication

```bash
API_KEY="your-generated-api-key"

# Using API key in header
curl http://localhost:8080/users \
  -H "X-API-Key: $API_KEY" \
  -H "Accept: application/json"
```

## Embedded Relationships

```bash
# Include related role in user response
curl "http://localhost:8080/users?select=id,email,name,role:role_id(id,name)" \
  -H "Authorization: Bearer $TOKEN"

# Include organization details
curl "http://localhost:8080/users?select=id,email,name,organization:organization_id(id,name,slug)" \
  -H "Authorization: Bearer $TOKEN"
```

## RPC Functions

```bash
# Health check
curl http://localhost:8080/rpc/healthcheck

# Get stats (admin only)
curl http://localhost:8080/rpc/get_stats \
  -H "Authorization: Bearer $TOKEN"

# Change password
curl -X POST http://localhost:8080/rpc/change_password \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{"p_user_id": "your-user-uuid", "p_old_password": "oldpass", "p_new_password": "newpass"}'

# Generate API key
curl -X POST http://localhost:8080/rpc/generate_api_key \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{"p_user_id": "your-user-uuid", "p_name": "My App Key", "p_permissions": ["users:read", "organizations:read"]}'
```

## Headers Reference

| Header | Purpose | Example |
|--------|---------|---------|
| `Authorization: Bearer <token>` | JWT Auth | `Bearer eyJhbGciOiJIUzI1NiIs...` |
| `X-API-Key: <key>` | API Key Auth | `X-API-Key: a1b2c3d4e5f6g7h8...` |
| `Accept: application/json` | Response format | Default |
| `Content-Type: application/json` | Request format | For POST/PATCH |
| `Prefer: count=exact` | Include row count | Returns `Content-Range` header |
| `Range: 0-9` | Range request | Alternative pagination |

## Status Codes

- `200` - Success
- `201` - Created (POST)
- `204` - No Content (DELETE, PATCH with no body)
- `401` - Unauthorized (missing/invalid token)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found
- `429` - Rate Limited
- `500` - Server Error
