<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';
  import type { DashboardOverview, SchemaInfo } from '$lib/types';
  import StatCard from '$lib/components/StatCard.svelte';
  import Card from '$lib/components/Card.svelte';
  import LoadingCard from '$lib/components/LoadingCard.svelte';
  import Skeleton from '$lib/components/Skeleton.svelte';

  let loading = $state(true);
  let error = $state('');
  let overview = $state<DashboardOverview | null>(null);
  let schemas = $state<SchemaInfo[]>([]);

  onMount(async () => {
    try {
      const [ov, sc] = await Promise.all([
        api.getOverview(),
        api.getSchemas(),
      ]);
      overview = ov;
      schemas = sc;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load dashboard data';
    } finally {
      loading = false;
    }
  });

  function formatBytes(bytes: number): string {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  }

  function formatUptime(seconds: number): string {
    const d = Math.floor(seconds / 86400);
    const h = Math.floor((seconds % 86400) / 3600);
    return `${d}d ${h}h`;
  }
</script>

<div class="max-w-6xl mx-auto">
  <h1 class="text-2xl font-bold mb-6" style="color: var(--text-primary);">Dashboard</h1>

  {#if error}
    <div class="card p-6 text-center">
      <div class="text-lg mb-2" style="color: var(--danger);">Failed to load dashboard</div>
      <p class="text-sm mb-4" style="color: var(--text-secondary);">{error}</p>
      <button onclick={() => window.location.reload()} class="btn btn-primary">Retry</button>
    </div>
  {:else if loading}
    <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3 mb-8">
      {#each [1,2,3,4,5,6] as _}
        <LoadingCard />
      {/each}
    </div>
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <Card title="Recent Activity"><Skeleton rows={4} /></Card>
      <Card title="Database Stats"><Skeleton rows={4} /></Card>
    </div>
  {:else}
    <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3 mb-8">
      <StatCard title="PG Version" value={overview?.pg_version || '-'} icon="◉" subtitle={overview?.databases?.length + ' databases'} />
      <StatCard title="Database Size" value={formatBytes(overview?.db_size || 0)} icon="▤" />
      <StatCard title="Connections" value={overview?.active_connections || 0} icon="◎" subtitle={'max ' + (overview as any)?.max_connections || '-'} />
      <StatCard title="Uptime" value={formatUptime((overview as any)?.uptime_seconds || 0)} icon="◈" subtitle={overview?.uptime || ''} />
      <StatCard title="Cache Hit Ratio" value={(overview?.cache_hit_ratio || 0).toFixed(1) + '%'} icon="◇" trend={overview && overview.cache_hit_ratio > 95 ? { value: 'Good', positive: true } : undefined} />
      <StatCard title="TPS" value={overview?.tps || 0} icon="▶" subtitle="Transactions/sec" />
    </div>

    <div class="flex items-center gap-2 mb-4">
      <div class="flex items-center gap-1.5">
        <span class="w-2 h-2 rounded-full" style="background-color: {overview?.replication_status === 'healthy' ? 'var(--success)' : overview?.replication_status === 'degraded' ? 'var(--warning)' : 'var(--danger)'};"></span>
        <span class="text-xs font-medium" style="color: var(--text-secondary);">Replication: {overview?.replication_status || 'unknown'}</span>
      </div>
      <span class="text-xs" style="color: var(--text-tertiary);">•</span>
      <span class="text-xs" style="color: var(--text-tertiary);">{overview?.table_count || 0} tables across {schemas.length} schemas</span>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <Card title="Recent Schemas">
        {#if schemas.length === 0}
          <p class="text-sm" style="color: var(--text-secondary);">No schemas found.</p>
        {:else}
          <div class="space-y-2">
            {#each schemas.slice(0, 8) as schema}
              <div class="flex items-center justify-between p-3 rounded-lg" style="background-color: var(--bg-hover);">
                <div>
                  <div class="text-sm font-medium" style="color: var(--text-primary);">{schema.schema_name}</div>
                  <div class="text-xs" style="color: var(--text-secondary);">Owner: {schema.owner}</div>
                </div>
                <div class="text-right">
                  <div class="text-sm font-mono" style="color: var(--text-secondary);">{formatBytes(schema.size_bytes)}</div>
                  <div class="text-xs" style="color: var(--text-tertiary);">{schema.table_count} tables</div>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </Card>

      <Card title="Quick Actions">
        <div class="space-y-2">
          <a href="/explorer" class="flex items-center gap-3 p-3 rounded-lg transition-colors no-underline" style="background-color: var(--bg-hover);">
            <span class="text-lg">▦</span>
            <div>
              <div class="text-sm font-medium" style="color: var(--text-primary);">Database Explorer</div>
              <div class="text-xs" style="color: var(--text-secondary);">Browse schemas, tables, views and more</div>
            </div>
          </a>
          <a href="/sql" class="flex items-center gap-3 p-3 rounded-lg transition-colors no-underline" style="background-color: var(--bg-hover);">
            <span class="text-lg">▶</span>
            <div>
              <div class="text-sm font-medium" style="color: var(--text-primary);">SQL Workspace</div>
              <div class="text-xs" style="color: var(--text-secondary);">Run queries and explore results</div>
            </div>
          </a>
          <a href="/monitoring" class="flex items-center gap-3 p-3 rounded-lg transition-colors no-underline" style="background-color: var(--bg-hover);">
            <span class="text-lg">◈</span>
            <div>
              <div class="text-sm font-medium" style="color: var(--text-primary);">Monitoring</div>
              <div class="text-xs" style="color: var(--text-secondary);">Active sessions, slow queries, locks</div>
            </div>
          </a>
          <a href="/backups" class="flex items-center gap-3 p-3 rounded-lg transition-colors no-underline" style="background-color: var(--bg-hover);">
            <span class="text-lg">☆</span>
            <div>
              <div class="text-sm font-medium" style="color: var(--text-primary);">Backups</div>
              <div class="text-xs" style="color: var(--text-secondary);">Create and manage database backups</div>
            </div>
          </a>
        </div>
      </Card>
    </div>
  {/if}
</div>