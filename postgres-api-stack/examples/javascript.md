# JavaScript Examples

## Using Fetch (Browser / Node 18+)

```javascript
const API = 'http://localhost:8080';
let token = null;

// Login
async function login(email, password) {
  const res = await fetch(`${API}/rpc/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    },
    body: JSON.stringify({ p_email: email, p_password: password }),
  });

  const data = await res.json();
  token = data.token;
  return data;
}

// Authenticated GET
async function getUsers() {
  const res = await fetch(`${API}/users?order=created_at.desc&limit=10`, {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Accept': 'application/json',
    },
  });

  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  const count = res.headers.get('Content-Range');
  const users = await res.json();
  return { users, count };
}

// Create organization
async function createOrganization(name, slug) {
  const res = await fetch(`${API}/organizations`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
      'Accept': 'application/json',
      'Prefer': 'return=representation',
    },
    body: JSON.stringify({ name, slug }),
  });

  return res.json();
}

// Update user
async function updateUser(id, updates) {
  const res = await fetch(`${API}/users?id=eq.${id}`, {
    method: 'PATCH',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    },
    body: JSON.stringify(updates),
  });

  return res.status === 204 ? 'Updated' : res.json();
}

// Delete user
async function deleteUser(id) {
  const res = await fetch(`${API}/users?id=eq.${id}`, {
    method: 'DELETE',
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });

  return res.status === 204 ? 'Deleted' : res.json();
}
```

## Using Axios

```javascript
import axios from 'axios';

const api = axios.create({
  baseURL: 'http://localhost:8080',
  headers: { 'Accept': 'application/json' },
});

let token = null;

// Set auth token
api.interceptors.request.use((config) => {
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Register
async function register(email, password, name) {
  const { data } = await api.post('/rpc/register', {
    p_email: email,
    p_password: password,
    p_name: name,
  });
  return data;
}

// Login
async function login(email, password) {
  const { data } = await api.post('/rpc/login', {
    p_email: email,
    p_password: password,
  });
  token = data.token;
  return data;
}

// CRUD
async function getUsers() {
  const { data, headers } = await api.get('/users', {
    params: { order: 'created_at.desc', limit: 20 },
  });
  return { users: data, total: headers['content-range'] };
}

async function createOrg(payload) {
  const { data } = await api.post('/organizations', payload, {
    headers: { 'Prefer': 'return=representation' },
  });
  return data;
}

async function updateUser(id, payload) {
  await api.patch(`/users?id=eq.${id}`, payload);
}

async function deleteUser(id) {
  await api.delete(`/users?id=eq.${id}`);
}

// Embedded relationships
async function getUsersWithRoles() {
  const { data } = await api.get('/users', {
    params: {
      select: 'id,email,name,role:role_id(id,name)',
    },
  });
  return data;
}

// RPC call
async function healthcheck() {
  const { data } = await api.get('/rpc/healthcheck');
  return data;
}
```

## Error Handling

```javascript
// Generic fetch wrapper
async function apiRequest(method, path, options = {}) {
  const { body, params, token: customToken } = options;

  const url = new URL(path, API);
  if (params) {
    Object.entries(params).forEach(([k, v]) => url.searchParams.set(k, v));
  }

  const headers = {
    'Accept': 'application/json',
    ...(token && { 'Authorization': `Bearer ${token || customToken}` }),
    ...(body && { 'Content-Type': 'application/json' }),
  };

  const res = await fetch(url, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  if (!res.ok) {
    const error = await res.json().catch(() => ({}));
    throw new Error(error.message || `HTTP ${res.status}: ${res.statusText}`);
  }

  if (res.status === 204) return null;
  return res.json();
}
```
