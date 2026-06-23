# Cloudflare Workers Example

## Basic Worker

```javascript
// wrangler.toml
// name = "my-api-client"
// main = "src/index.js"
//
// [vars]
// API_BASE_URL = "https://api.yourdomain.com"
// API_KEY = "your-api-key"

// src/index.js
export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    const path = url.pathname;
    const method = request.method;

    // Route requests to our PostgREST API
    const apiUrl = `${env.API_BASE_URL}${path}${url.search}`;

    const headers = new Headers(request.headers);
    headers.set('X-API-Key', env.API_KEY);
    headers.set('Accept', 'application/json');

    // Forward the request to PostgREST
    const apiResponse = await fetch(apiUrl, {
      method,
      headers,
      body: method !== 'GET' && method !== 'HEAD' ? request.body : undefined,
    });

    // Add CORS headers
    const response = new Response(apiResponse.body, apiResponse);
    response.headers.set('Access-Control-Allow-Origin', '*');
    response.headers.set('Access-Control-Allow-Methods', 'GET, POST, PATCH, DELETE, OPTIONS');
    response.headers.set('Access-Control-Allow-Headers', 'Authorization, Content-Type');

    return response;
  },
};
```

## User Authentication Worker

```javascript
// src/auth.js
export default {
  async fetch(request, env, ctx) {
    if (request.method === 'OPTIONS') {
      return new Response(null, {
        headers: {
          'Access-Control-Allow-Origin': '*',
          'Access-Control-Allow-Methods': 'GET, POST, PATCH, DELETE, OPTIONS',
          'Access-Control-Allow-Headers': 'Authorization, Content-Type',
          'Access-Control-Max-Age': '86400',
        },
      });
    }

    const url = new URL(request.url);
    const { pathname } = url;

    try {
      // Public endpoints
      if (pathname.startsWith('/auth/')) {
        return handleAuth(request, env);
      }

      // Protected endpoints - verify JWT
      const token = extractToken(request);
      if (!token) {
        return jsonResponse({ error: 'Unauthorized' }, 401);
      }

      // Verify JWT and forward with user context
      const payload = await verifyJWT(token, env.JWT_SECRET);

      // Forward to PostgREST with Authorization header
      const apiUrl = `${env.API_BASE_URL}${pathname}${url.search}`;
      const headers = new Headers(request.headers);
      headers.set('Authorization', `Bearer ${token}`);
      headers.set('X-User-Id', payload.sub);
      headers.set('X-User-Role', payload.role || 'authenticated');

      const apiResponse = await fetch(apiUrl, {
        method: request.method,
        headers,
        body: ['GET', 'HEAD'].includes(request.method) ? null : request.body,
      });

      return apiResponse;
    } catch (err) {
      return jsonResponse({ error: err.message }, 500);
    }
  },
};

async function handleAuth(request, env) {
  const url = new URL(request.url);
  const action = url.pathname.replace('/auth/', '');
  const body = await request.json();

  switch (action) {
    case 'login': {
      const res = await fetch(`${env.API_BASE_URL}/rpc/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
        body: JSON.stringify({
          p_email: body.email,
          p_password: body.password,
        }),
      });
      return res;
    }

    case 'register': {
      const res = await fetch(`${env.API_BASE_URL}/rpc/register`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
        body: JSON.stringify({
          p_email: body.email,
          p_password: body.password,
          p_name: body.name || null,
        }),
      });
      return res;
    }

    default:
      return jsonResponse({ error: 'Unknown action' }, 404);
  }
}

function extractToken(request) {
  const auth = request.headers.get('Authorization');
  if (auth?.startsWith('Bearer ')) {
    return auth.slice(7);
  }
  return request.headers.get('X-API-Key') || null;
}

async function verifyJWT(token, secret) {
  // In production, use a JWT library like jose
  // import * as jose from 'jose';
  // const { payload } = await jose.jwtVerify(token, new TextEncoder().encode(secret));
  // return payload;
  return { sub: 'user-id', role: 'authenticated' };
}

function jsonResponse(data, status = 200) {
  return new Response(JSON.stringify(data), {
    status,
    headers: {
      'Content-Type': 'application/json',
      'Access-Control-Allow-Origin': '*',
    },
  });
}
```

## Durable Objects Session Store (Advanced)

```javascript
// src/session.js
export class SessionStore {
  constructor(state, env) {
    this.state = state;
  }

  async fetch(request) {
    const url = new URL(request.url);
    const key = url.searchParams.get('key');

    if (request.method === 'GET' && key) {
      const value = await this.state.storage.get(key);
      return new Response(JSON.stringify({ value }), {
        headers: { 'Content-Type': 'application/json' },
      });
    }

    if (request.method === 'PUT' && key) {
      const body = await request.json();
      await this.state.storage.put(key, body.value, {
        expirationTtl: body.ttl || 86400,
      });
      return new Response(JSON.stringify({ ok: true }), {
        headers: { 'Content-Type': 'application/json' },
      });
    }

    if (request.method === 'DELETE' && key) {
      await this.state.storage.delete(key);
      return new Response(JSON.stringify({ ok: true }), {
        headers: { 'Content-Type': 'application/json' },
      });
    }

    return new Response('Not found', { status: 404 });
  }
}
```

## KV Cache Layer

```javascript
// src/cache.js
// Cache API responses in KV for faster reads

const CACHE_TTL = 60; // 60 seconds

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    const cacheKey = `api:${url.pathname}${url.search}`;

    // Only cache GET requests
    if (request.method === 'GET') {
      const cached = await env.API_CACHE.get(cacheKey, { type: 'json' });
      if (cached) {
        return new Response(JSON.stringify(cached), {
          headers: {
            'Content-Type': 'application/json',
            'X-Cache': 'HIT',
          },
        });
      }
    }

    // Forward to API
    const apiUrl = `${env.API_BASE_URL}${url.pathname}${url.search}`;
    const headers = new Headers(request.headers);
    headers.set('X-API-Key', env.API_KEY);

    const apiResponse = await fetch(apiUrl, {
      method: request.method,
      headers,
      body: ['GET', 'HEAD'].includes(request.method) ? null : request.body,
    });

    // Cache successful GET responses
    if (request.method === 'GET' && apiResponse.ok) {
      const data = await apiResponse.json();
      ctx.waitUntil(
        env.API_CACHE.put(cacheKey, JSON.stringify(data), {
          expirationTtl: CACHE_TTL,
        })
      );
      return new Response(JSON.stringify(data), {
        headers: {
          'Content-Type': 'application/json',
          'X-Cache': 'MISS',
        },
      });
    }

    return apiResponse;
  },
};
```

## Environment Variables (wrangler.toml)

```toml
name = "postgres-api-client"
main = "src/index.js"
compatibility_date = "2024-01-01"

[vars]
API_BASE_URL = "https://api.yourdomain.com"
API_KEY = "your-api-key"
JWT_SECRET = "your-jwt-secret"

[[kv_namespaces]]
binding = "API_CACHE"
id = "your-kv-namespace-id"
```
