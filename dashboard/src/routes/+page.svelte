<script lang="ts">
  import { isAuthenticated } from '$lib/stores/auth';
  import { goto } from '$app/navigation';
  import { getBaseUrl } from '$lib/api/client';
  import { APP_NAME } from '$lib/config/brand';
  import { onMount } from 'svelte';

  const baseUrl = getBaseUrl();

  onMount(() => {
    if ($isAuthenticated) {
      goto('/dashboard');
    }
  });
</script>

{#if !$isAuthenticated}
  <div class="max-w-4xl mx-auto space-y-8 pb-12">
    <div class="text-center pt-4">
      <h1 class="text-2xl font-bold">{APP_NAME}</h1>
      <p class="text-sm mt-1" style="color: var(--text-secondary);">REST API Guide — PostgreSQL Administration Backend</p>
    </div>

    <div class="card p-6">
      <h2 class="text-lg font-semibold mb-3">Base URL</h2>
      <p class="text-sm mb-2" style="color: var(--text-secondary);">All endpoints are served from the following base:</p>
      <div class="px-3 py-2 rounded-lg text-sm font-mono" style="background-color: var(--bg-hover); color: var(--accent);">{baseUrl}/v1</div>
    </div>

    <div class="card p-6">
      <h2 class="text-lg font-semibold mb-3">Authentication</h2>
      <p class="text-sm mb-3" style="color: var(--text-secondary);">Authenticate with your email and password to receive a JWT. Include the token in all subsequent requests.</p>

      <div class="mb-4">
        <h3 class="text-sm font-medium mb-2">Login</h3>
        <div class="px-3 py-2 rounded-lg text-xs font-mono mb-2" style="background-color: var(--bg-hover);">
          <span style="color: var(--accent);">POST</span> <span style="color: var(--text-primary);">/v1/auth/login</span>
        </div>
        <div class="px-3 py-2 rounded-lg text-xs font-mono" style="background-color: var(--bg-hover); color: var(--text-secondary);">
{`{
  "email": "admin@example.com",
  "password": "your-password"
}`}
        </div>
        <div class="px-3 py-2 mt-2 rounded-lg text-xs font-mono" style="background-color: var(--bg-hover); color: var(--text-secondary);">
{`{
  "access_token": "eyJhbGciOi...",
  "refresh_token": "dGhpcyBpcyBh...",
  "expires_at": "2026-06-29T00:00:00Z"
}`}
        </div>
      </div>

      <div class="mb-4">
        <h3 class="text-sm font-medium mb-2">Using the Token</h3>
        <p class="text-xs mb-2" style="color: var(--text-secondary);">Include the access token in the Authorization header for all authenticated endpoints:</p>
        <div class="px-3 py-2 rounded-lg text-xs font-mono" style="background-color: var(--bg-hover);">
          <span style="color: var(--text-secondary);">Authorization: Bearer &lt;access_token&gt;</span>
        </div>
      </div>

      <div class="mb-4">
        <h3 class="text-sm font-medium mb-2">Refresh Token</h3>
        <p class="text-xs mb-2" style="color: var(--text-secondary);">When the access token expires, use the refresh token to get a new one:</p>
        <div class="px-3 py-2 rounded-lg text-xs font-mono mb-2" style="background-color: var(--bg-hover);">
          <span style="color: var(--accent);">POST</span> <span style="color: var(--text-primary);">/v1/auth/refresh</span>
        </div>
        <div class="px-3 py-2 rounded-lg text-xs font-mono" style="background-color: var(--bg-hover); color: var(--text-secondary);">
{`{ "refresh_token": "dGhpcyBpcyBh..." }`}
        </div>
      </div>

      <div>
        <h3 class="text-sm font-medium mb-2">Current User</h3>
        <div class="px-3 py-2 rounded-lg text-xs font-mono mb-2" style="background-color: var(--bg-hover);">
          <span style="color: var(--success);">GET</span> <span style="color: var(--text-primary);">/v1/auth/me</span>
        </div>
        <p class="text-xs" style="color: var(--text-secondary);">Returns the authenticated user's profile including id, email, and role.</p>
      </div>
    </div>

    <div class="card p-6">
      <h2 class="text-lg font-semibold mb-3">Dashboard &amp; Explorer</h2>
      <p class="text-xs mb-3" style="color: var(--text-secondary);">Read-only views of database health, schema listings, and object metadata.</p>

      <div class="space-y-3">
        <div>
          <div class="px-3 py-2 rounded-lg text-xs font-mono mb-1" style="background-color: var(--bg-hover);">
            <span style="color: var(--success);">GET</span> <span style="color: var(--text-primary);">/v1/dashboard/overview</span>
          </div>
          <p class="text-xs" style="color: var(--text-secondary);">Database version, size, active connections, uptime, cache hit ratio, TPS, replication status.</p>
        </div>
        <div>
          <div class="px-3 py-2 rounded-lg text-xs font-mono mb-1" style="background-color: var(--bg-hover);">
            <span style="color: var(--success);">GET</span> <span style="color: var(--text-primary);">/v1/dashboard/stats</span>
          </div>
          <p class="text-xs" style="color: var(--text-secondary);">Same as overview but returns only the stats object.</p>
        </div>
        <div>
          <div class="px-3 py-2 rounded-lg text-xs font-mono mb-1" style="background-color: var(--bg-hover);">
            <span style="color: var(--success);">GET</span> <span style="color: var(--text-primary);">/v1/explorer/schemas</span>
          </div>
          <p class="text-xs" style="color: var(--text-secondary);">List all schemas with owner, size, and table count.</p>
        </div>
        <div>
          <div class="px-3 py-2 rounded-lg text-xs font-mono mb-1" style="background-color: var(--bg-hover);">
            <span style="color: var(--success);">GET</span> <span style="color: var(--text-primary);">/v1/explorer/schemas/:schema/tables</span>
          </div>
          <p class="text-xs" style="color: var(--text-secondary);">List tables in a schema with row counts and sizes.</p>
        </div>
      </div>
    </div>

    <div class="card p-6">
      <h2 class="text-lg font-semibold mb-3">SQL Workspace</h2>
      <p class="text-xs mb-3" style="color: var(--text-secondary);">Execute arbitrary SQL queries against the database.</p>

      <div class="space-y-3">
        <div>
          <div class="px-3 py-2 rounded-lg text-xs font-mono mb-1" style="background-color: var(--bg-hover);">
            <span style="color: var(--accent);">POST</span> <span style="color: var(--text-primary);">/v1/sql/execute</span>
          </div>
          <div class="px-3 py-2 rounded-lg text-xs font-mono mb-1" style="background-color: var(--bg-hover); color: var(--text-secondary);">
{`{ "query": "SELECT * FROM pg_stat_activity LIMIT 10" }`}
          </div>
          <p class="text-xs" style="color: var(--text-secondary);">Executes the query and returns rows, columns, row count, and execution time.</p>
        </div>
        <div>
          <div class="px-3 py-2 rounded-lg text-xs font-mono mb-1" style="background-color: var(--bg-hover);">
            <span style="color: var(--warning);">POST</span> <span style="color: var(--text-primary);">/v1/sql/explain</span>
          </div>
          <p class="text-xs" style="color: var(--text-secondary);">Returns the query execution plan without running the query.</p>
        </div>
        <div>
          <div class="px-3 py-2 rounded-lg text-xs font-mono mb-1" style="background-color: var(--bg-hover);">
            <span style="color: var(--accent);">POST</span> <span style="color: var(--text-primary);">/v1/sql/cancel</span>
          </div>
          <div class="px-3 py-2 rounded-lg text-xs font-mono mb-1" style="background-color: var(--bg-hover); color: var(--text-secondary);">
{`{ "pid": 12345 }`}
          </div>
          <p class="text-xs" style="color: var(--text-secondary);">Cancel a running query by its process ID.</p>
        </div>
      </div>
    </div>

    <div class="card p-6">
      <h2 class="text-lg font-semibold mb-3">Schema &amp; Table Management</h2>
      <p class="text-xs mb-3" style="color: var(--text-secondary);">Create, alter, and drop database objects.</p>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-3 text-xs">
        <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);">
          <span style="color: var(--accent);">POST</span> <span style="color: var(--text-primary);">/v1/schemas</span>
          <p class="mt-0.5" style="color: var(--text-secondary);">Create a new schema</p>
        </div>
        <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);">
          <span style="color: var(--danger);">DELETE</span> <span style="color: var(--text-primary);">/v1/schemas/:name</span>
          <p class="mt-0.5" style="color: var(--text-secondary);">Drop a schema (?cascade=true)</p>
        </div>
        <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);">
          <span style="color: var(--accent);">POST</span> <span style="color: var(--text-primary);">/v1/schemas/:schema/tables</span>
          <p class="mt-0.5" style="color: var(--text-secondary);">Create a table with columns</p>
        </div>
        <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);">
          <span style="color: var(--danger);">DELETE</span> <span style="color: var(--text-primary);">/v1/schemas/:schema/tables/:table</span>
          <p class="mt-0.5" style="color: var(--text-secondary);">Drop a table</p>
        </div>
        <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);">
          <span style="color: var(--accent);">POST</span> <span style="color: var(--text-primary);">/v1/schemas/:schema/tables/:table/columns</span>
          <p class="mt-0.5" style="color: var(--text-secondary);">Add a column</p>
        </div>
        <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);">
          <span style="color: var(--warning);">PATCH</span> <span style="color: var(--text-primary);">/v1/schemas/:schema/tables/:table/columns/:column</span>
          <p class="mt-0.5" style="color: var(--text-secondary);">Alter a column</p>
        </div>
      </div>
    </div>

    <div class="card p-6">
      <h2 class="text-lg font-semibold mb-3">Row Operations</h2>
      <p class="text-xs mb-3" style="color: var(--text-secondary);">CRUD operations on table data.</p>

      <div class="space-y-3">
        <div>
          <div class="px-3 py-2 rounded-lg text-xs font-mono mb-1" style="background-color: var(--bg-hover);">
            <span style="color: var(--success);">GET</span> <span style="color: var(--text-primary);">/v1/tables/:schema/:table?limit=50&amp;offset=0</span>
          </div>
          <p class="text-xs" style="color: var(--text-secondary);">Browse rows with pagination.</p>
        </div>
        <div>
          <div class="px-3 py-2 rounded-lg text-xs font-mono mb-1" style="background-color: var(--bg-hover);">
            <span style="color: var(--accent);">POST</span> <span style="color: var(--text-primary);">/v1/tables/:schema/:table/rows</span>
          </div>
          <div class="px-3 py-2 rounded-lg text-xs font-mono mb-1" style="background-color: var(--bg-hover); color: var(--text-secondary);">
{`{ "values": { "name": "example", "status": "active" } }`}
          </div>
          <p class="text-xs" style="color: var(--text-secondary);">Insert a single row.</p>
        </div>
        <div>
          <div class="px-3 py-2 rounded-lg text-xs font-mono mb-1" style="background-color: var(--bg-hover);">
            <span style="color: var(--warning);">PATCH</span> <span style="color: var(--text-primary);">/v1/tables/:schema/:table/rows</span>
          </div>
          <div class="px-3 py-2 rounded-lg text-xs font-mono mb-1" style="background-color: var(--bg-hover); color: var(--text-secondary);">
{`{ "values": { "status": "inactive" }, "where": { "id": 1 } }`}
          </div>
          <p class="text-xs" style="color: var(--text-secondary);">Update rows matching the where condition.</p>
        </div>
        <div>
          <div class="px-3 py-2 rounded-lg text-xs font-mono mb-1" style="background-color: var(--bg-hover);">
            <span style="color: var(--danger);">DELETE</span> <span style="color: var(--text-primary);">/v1/tables/:schema/:table/rows</span>
          </div>
          <div class="px-3 py-2 rounded-lg text-xs font-mono mb-1" style="background-color: var(--bg-hover); color: var(--text-secondary);">
{`{ "where": { "id": 1 } }`}
          </div>
          <p class="text-xs" style="color: var(--text-secondary);">Delete rows matching the where condition.</p>
        </div>
        <div>
          <div class="px-3 py-2 rounded-lg text-xs font-mono mb-1" style="background-color: var(--bg-hover);">
            <span style="color: var(--success);">GET</span> <span style="color: var(--text-primary);">/v1/tables/:schema/:table/search?q=term&amp;limit=20</span>
          </div>
          <p class="text-xs" style="color: var(--text-secondary);">Full-text search across all columns in a table.</p>
        </div>
      </div>
    </div>

    <div class="card p-6">
      <h2 class="text-lg font-semibold mb-3">PostgreSQL Roles &amp; Privileges</h2>
      <p class="text-xs mb-3" style="color: var(--text-secondary);">Manage database roles, memberships, and object-level privileges.</p>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-2 text-xs">
        <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--success);">GET</span> <span style="color: var(--text-primary);">/v1/roles</span></div>
        <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--accent);">POST</span> <span style="color: var(--text-primary);">/v1/roles</span></div>
        <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--warning);">PATCH</span> <span style="color: var(--text-primary);">/v1/roles/:name</span></div>
        <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--danger);">DELETE</span> <span style="color: var(--text-primary);">/v1/roles/:name</span></div>
        <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--accent);">POST</span> <span style="color: var(--text-primary);">/v1/roles/:name/password</span></div>
        <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--accent);">POST</span> <span style="color: var(--text-primary);">/v1/roles/:role/grant-database</span></div>
        <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--accent);">POST</span> <span style="color: var(--text-primary);">/v1/roles/:role/grant-schema</span></div>
        <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--accent);">POST</span> <span style="color: var(--text-primary);">/v1/roles/:role/grant-table</span></div>
        <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--accent);">POST</span> <span style="color: var(--text-primary);">/v1/roles/:role/revoke-database</span></div>
        <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--accent);">POST</span> <span style="color: var(--text-primary);">/v1/roles/:role/revoke-schema</span></div>
        <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--accent);">POST</span> <span style="color: var(--text-primary);">/v1/roles/:role/add-member</span></div>
        <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--accent);">POST</span> <span style="color: var(--text-primary);">/v1/roles/:role/remove-member</span></div>
        <div class="px-3 py-2 rounded-lg md:col-span-2" style="background-color: var(--bg-hover);"><span style="color: var(--success);">GET</span> <span style="color: var(--text-primary);">/v1/roles/privileges/databases?database=name</span></div>
        <div class="px-3 py-2 rounded-lg md:col-span-2" style="background-color: var(--bg-hover);"><span style="color: var(--success);">GET</span> <span style="color: var(--text-primary);">/v1/roles/:name/members</span></div>
      </div>
    </div>

    <div class="card p-6">
      <h2 class="text-lg font-semibold mb-3">Extensions &amp; Monitoring</h2>

      <div class="space-y-3">
        <div>
          <h3 class="text-sm font-medium mb-2">Extensions</h3>
          <div class="grid grid-cols-1 md:grid-cols-3 gap-2 text-xs">
            <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--success);">GET</span> <span style="color: var(--text-primary);">/v1/extensions</span></div>
            <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--accent);">POST</span> <span style="color: var(--text-primary);">/v1/extensions <span class="block mt-0.5" style="color: var(--text-secondary);">{'{'}"name":"pg_stat_statements","schema":"public"{'}'}</span></span></div>
            <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--danger);">DELETE</span> <span style="color: var(--text-primary);">/v1/extensions/:name</span></div>
          </div>
        </div>

        <div>
          <h3 class="text-sm font-medium mb-2">Monitoring</h3>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-2 text-xs">
            <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--success);">GET</span> <span style="color: var(--text-primary);">/v1/monitoring/sessions</span></div>
            <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--success);">GET</span> <span style="color: var(--text-primary);">/v1/monitoring/slow-queries?min_duration_ms=1000</span></div>
            <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--success);">GET</span> <span style="color: var(--text-primary);">/v1/monitoring/locks</span></div>
            <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--success);">GET</span> <span style="color: var(--text-primary);">/v1/monitoring/connections</span></div>
            <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--success);">GET</span> <span style="color: var(--text-primary);">/v1/monitoring/cache</span></div>
            <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--success);">GET</span> <span style="color: var(--text-primary);">/v1/monitoring/table-stats?schema=public</span></div>
            <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--accent);">POST</span> <span style="color: var(--text-primary);">/v1/monitoring/sessions/terminate</span></div>
          </div>
        </div>
      </div>
    </div>

    <div class="card p-6">
      <h2 class="text-lg font-semibold mb-3">Backups &amp; Logs</h2>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-2 text-xs">
        <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--success);">GET</span> <span style="color: var(--text-primary);">/v1/backups</span></div>
        <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--accent);">POST</span> <span style="color: var(--text-primary);">/v1/backups <span class="block mt-0.5" style="color: var(--text-secondary);">{'{'}"database":"mydb","type":"full"{'}'}</span></span></div>
        <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--accent);">POST</span> <span style="color: var(--text-primary);">/v1/backups/:id/restore</span></div>
        <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--accent);">POST</span> <span style="color: var(--text-primary);">/v1/backups/:id/verify</span></div>
        <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--success);">GET</span> <span style="color: var(--text-primary);">/v1/backups/:id/download</span></div>
        <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--danger);">DELETE</span> <span style="color: var(--text-primary);">/v1/backups/:id</span></div>
        <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--success);">GET</span> <span style="color: var(--text-primary);">/v1/logs?limit=50&amp;offset=0</span></div>
        <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--success);">GET</span> <span style="color: var(--text-primary);">/v1/logs/query</span></div>
        <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--success);">GET</span> <span style="color: var(--text-primary);">/v1/logs/auth</span></div>
        <div class="px-3 py-2 rounded-lg" style="background-color: var(--bg-hover);"><span style="color: var(--success);">GET</span> <span style="color: var(--text-primary);">/v1/audit-logs</span></div>
        <div class="px-3 py-2 rounded-lg md:col-span-2" style="background-color: var(--bg-hover);"><span style="color: var(--success);">GET</span> <span style="color: var(--text-primary);">/v1/audit-logs/:resource/:resource_id</span></div>
      </div>
    </div>

    <div class="card p-6">
      <h2 class="text-lg font-semibold mb-3">Quick Example (cURL)</h2>
      <div class="px-4 py-3 rounded-lg text-xs font-mono whitespace-pre-wrap" style="background-color: var(--bg-hover); color: var(--text-secondary);">
<span style="color: var(--text-secondary);"># 1. Login</span>
curl -s {baseUrl}/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{'{'}<span style="color: var(--text-primary);">"email":"admin@example.com","password":"your-password"</span>{'}'} \
  | jq .

<span style="color: var(--text-secondary);"># 2. Store the token</span>
TOKEN="eyJhbGciOi..."

<span style="color: var(--text-secondary);"># 3. Use the API</span>
curl -s {baseUrl}/v1/dashboard/overview \
  -H "Authorization: Bearer $TOKEN" \
  | jq .

<span style="color: var(--text-secondary);"># 4. Run a query</span>
curl -s {baseUrl}/v1/sql/execute \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{'{'}<span style="color: var(--text-primary);">"query":"SELECT version()"</span>{'}'} \
  | jq .</div>
    </div>

    <div class="text-center pt-2 pb-4">
    </div>
  </div>
{/if}