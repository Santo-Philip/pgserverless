<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { api } from '$lib/api/client';
  import type { ActiveSession, SlowQuery, LockInfo, WaitingQuery, ConnectionStats, CacheStats, DatabaseStat, TableStat, IndexStat } from '$lib/types';
  import Card from '$lib/components/Card.svelte';
  import StatCard from '$lib/components/StatCard.svelte';
  import LoadingCard from '$lib/components/LoadingCard.svelte';
  import Skeleton from '$lib/components/Skeleton.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';

  const tabs = ['Sessions', 'Slow Queries', 'Locks', 'Waiting', 'Connections', 'Cache', 'Databases', 'Tables', 'Indexes'] as const;
  type Tab = typeof tabs[number];

  let loading = $state(true);
  let error = $state('');
  let activeTab = $state<Tab>('Sessions');
  let autoRefresh = $state(false);
  let refreshInterval: ReturnType<typeof setInterval> | null = null;

  let sessions = $state<ActiveSession[]>([]);
  let slowQueries = $state<SlowQuery[]>([]);
  let locks = $state<LockInfo[]>([]);
  let waiting = $state<WaitingQuery[]>([]);
  let connStats = $state<ConnectionStats | null>(null);
  let cacheStats = $state<CacheStats | null>(null);
  let dbStats = $state<DatabaseStat[]>([]);
  let tableStats = $state<TableStat[]>([]);
  let indexStats = $state<IndexStat[]>([]);

  let terminatePid = $state<number | null>(null);
  let cancelPid = $state<number | null>(null);

  onMount(() => { loadAll(); });

  onDestroy(() => {
    if (refreshInterval) clearInterval(refreshInterval);
  });

  $effect(() => {
    if (autoRefresh && !refreshInterval) {
      refreshInterval = setInterval(loadAll, 5000);
    } else if (!autoRefresh && refreshInterval) {
      clearInterval(refreshInterval);
      refreshInterval = null;
    }
  });

  async function loadAll() {
    try {
      const results = await Promise.all([
        api.getActiveSessions(),
        api.getSlowQueries(1000).catch(() => [] as SlowQuery[]),
        api.getLocks().catch(() => [] as LockInfo[]),
        api.getWaitingQueries().catch(() => [] as WaitingQuery[]),
        api.getConnectionStats().catch(() => null),
        api.getCacheStats().catch(() => null),
        api.getDatabaseStats().catch(() => [] as DatabaseStat[]),
        api.getTableStats().catch(() => [] as TableStat[]),
        api.getIndexStats().catch(() => [] as IndexStat[]),
      ]);
      sessions = results[0];
      slowQueries = results[1] as SlowQuery[];
      locks = results[2] as LockInfo[];
      waiting = results[3] as WaitingQuery[];
      connStats = results[4] as ConnectionStats | null;
      cacheStats = results[5] as CacheStats | null;
      dbStats = results[6] as DatabaseStat[];
      tableStats = results[7] as TableStat[];
      indexStats = results[8] as IndexStat[];
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load monitoring data';
    } finally {
      loading = false;
    }
  }

  async function handleTerminate() {
    if (terminatePid === null) return;
    try {
      await api.terminateSession(terminatePid);
      terminatePid = null;
      loadAll();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to terminate session';
    }
  }

  async function handleCancel() {
    if (cancelPid === null) return;
    try {
      await api.cancelQuery(cancelPid);
      cancelPid = null;
      loadAll();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to cancel query';
    }
  }

  function formatBytes(bytes: number): string {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  }

  function truncate(str: string, len: number): string {
    return str.length > len ? str.slice(0, len) + '...' : str;
  }
</script>

<div class="max-w-6xl mx-auto">
  <div class="flex items-center justify-between mb-6">
    <h1 class="text-2xl font-bold" style="color: var(--text-primary);">Monitoring</h1>
    <div class="flex items-center gap-2">
      <button onclick={() => loadAll()} class="btn btn-secondary btn-sm" disabled={loading}>Refresh</button>
      <label class="flex items-center gap-2 text-xs cursor-pointer" style="color: var(--text-secondary);">
        <input type="checkbox" bind:checked={autoRefresh} />
        Auto-refresh (5s)
      </label>
    </div>
  </div>

  <div class="flex gap-0 border-b mb-4 overflow-x-auto" style="border-color: var(--border-primary);">
    {#each tabs as tab}
      <button onclick={() => activeTab = tab} class="tab whitespace-nowrap" class:active={activeTab === tab}>{tab}</button>
    {/each}
  </div>

  {#if error}
    <div class="card p-4 mb-4" style="border-color: rgba(239,68,68,0.3);">
      <p class="text-sm" style="color: var(--danger);">{error}</p>
    </div>
  {/if}

  {#if loading}
    <LoadingCard />
    <Skeleton rows={6} />
  {:else if activeTab === 'Sessions'}
    <div class="card overflow-hidden">
      <table class="table-wrap w-full">
        <thead><tr><th>PID</th><th>Database</th><th>User</th><th>Application</th><th>State</th><th>Query</th><th>Wait</th><th>Actions</th></tr></thead>
        <tbody>
          {#each sessions as s}
            <tr>
              <td class="font-mono text-xs">{s.pid}</td>
              <td>{s.database}</td>
              <td>{s.user}</td>
              <td class="text-xs">{s.application_name}</td>
              <td><span class="badge" style="background-color: {s.state === 'active' ? 'rgba(34,197,94,0.1)' : 'rgba(107,114,128,0.1)'}; color: {s.state === 'active' ? 'var(--success)' : 'var(--text-tertiary)'};">{s.state}</span></td>
              <td class="font-mono text-xs max-w-xs truncate">{truncate(s.query, 80)}</td>
              <td class="text-xs">{s.wait_event || '-'}</td>
              <td>
                <div class="flex gap-1">
                  <button onclick={() => cancelPid = s.pid} class="btn btn-ghost btn-sm" title="Cancel Query">✕</button>
                  <button onclick={() => terminatePid = s.pid} class="btn btn-ghost btn-sm" style="color: var(--danger);" title="Terminate">⊘</button>
                </div>
              </td>
            </tr>
          {:else}
            <tr><td colspan="8" class="text-center py-8" style="color: var(--text-tertiary);">No active sessions</td></tr>
          {/each}
        </tbody>
      </table>
    </div>

  {:else if activeTab === 'Slow Queries'}
    <div class="card overflow-hidden">
      <table class="table-wrap w-full">
        <thead><tr><th>PID</th><th>User</th><th>Database</th><th>Duration (ms)</th><th>State</th><th>Query</th><th>Started</th></tr></thead>
        <tbody>
          {#each slowQueries as q}
            <tr>
              <td class="font-mono text-xs">{q.pid}</td>
              <td>{q.user}</td>
              <td>{q.database}</td>
              <td class="font-mono text-xs" style="color: {q.duration_ms > 5000 ? 'var(--danger)' : 'var(--warning)'};">{q.duration_ms.toLocaleString()}</td>
              <td><span class="badge">{q.state}</span></td>
              <td class="font-mono text-xs max-w-xs truncate">{truncate(q.query, 80)}</td>
              <td class="text-xs">{new Date(q.query_start).toLocaleString()}</td>
            </tr>
          {:else}
            <tr><td colspan="7" class="text-center py-8" style="color: var(--text-tertiary);">No slow queries</td></tr>
          {/each}
        </tbody>
      </table>
    </div>

  {:else if activeTab === 'Locks'}
    <div class="card overflow-hidden">
      <table class="table-wrap w-full">
        <thead><tr><th>PID</th><th>Database</th><th>Relation</th><th>Type</th><th>Mode</th><th>Granted</th><th>Blocked By</th></tr></thead>
        <tbody>
          {#each locks as l}
            <tr>
              <td class="font-mono text-xs">{l.pid}</td>
              <td>{l.database}</td>
              <td>{l.relation}</td>
              <td>{l.lock_type}</td>
              <td>{l.lock_mode}</td>
              <td>
                <span style="color: {l.granted ? 'var(--success)' : 'var(--danger)'};">{l.granted ? 'Yes' : 'No'}</span>
              </td>
              <td>{l.blocked_by?.join(', ') || '-'}</td>
            </tr>
          {:else}
            <tr><td colspan="7" class="text-center py-8" style="color: var(--text-tertiary);">No locks</td></tr>
          {/each}
        </tbody>
      </table>
    </div>

  {:else if activeTab === 'Waiting'}
    <div class="card overflow-hidden">
      <table class="table-wrap w-full">
        <thead><tr><th>PID</th><th>User</th><th>Database</th><th>Blocked By</th><th>Wait (ms)</th><th>State</th><th>Query</th></tr></thead>
        <tbody>
          {#each waiting as w}
            <tr>
              <td class="font-mono text-xs">{w.pid}</td>
              <td>{w.user}</td>
              <td>{w.database}</td>
              <td class="font-mono text-xs">{w.blocked_by_pid}</td>
              <td class="font-mono text-xs">{w.wait_duration_ms?.toLocaleString()}</td>
              <td><span class="badge">{w.state}</span></td>
              <td class="font-mono text-xs max-w-xs truncate">{truncate(w.query, 80)}</td>
            </tr>
          {:else}
            <tr><td colspan="7" class="text-center py-8" style="color: var(--text-tertiary);">No waiting queries</td></tr>
          {/each}
        </tbody>
      </table>
    </div>

  {:else if activeTab === 'Connections'}
    <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3 mb-6">
      <StatCard title="Total" value={connStats?.total || 0} icon="◎" subtitle={'max ' + connStats?.max_connections} />
      <StatCard title="Active" value={connStats?.active || 0} icon="▶" trend={connStats && connStats.active > connStats.max_connections * 0.8 ? { value: 'High', positive: false } : undefined} />
      <StatCard title="Idle" value={connStats?.idle || 0} icon="◈" />
      <StatCard title="Idle in Transaction" value={connStats?.idle_in_transaction || 0} icon="◉" />
      <StatCard title="Waiting" value={connStats?.waiting || 0} icon="◈" />
      <StatCard title="Max" value={connStats?.max_connections || 0} icon="☆" />
    </div>
    {#if connStats?.by_database}
      <Card title="By Database">
        <div class="space-y-2">
          {#each connStats.by_database as db}
            <div class="flex items-center justify-between p-2 rounded" style="background-color: var(--bg-hover);">
              <span class="text-sm">{db.database}</span>
              <span class="text-sm font-mono">{db.count}</span>
            </div>
          {/each}
        </div>
      </Card>
    {/if}

  {:else if activeTab === 'Cache'}
    {#if cacheStats}
      <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-6">
        <StatCard title="Hit Ratio" value={(cacheStats.hit_ratio * 100).toFixed(1) + '%'} icon="◇" trend={{ value: cacheStats.hit_ratio > 0.95 ? 'Good' : 'Low', positive: cacheStats.hit_ratio > 0.95 }} />
        <StatCard title="Shared Hit" value={cacheStats.shared_hit?.toLocaleString() || '0'} icon="◉" subtitle="blocks" />
        <StatCard title="Shared Read" value={cacheStats.shared_read?.toLocaleString() || '0'} icon="◉" subtitle="blocks" />
        <StatCard title="WAL Hit Ratio" value={((cacheStats.hit_ratio_wal || 0) * 100).toFixed(1) + '%'} icon="☆" />
      </div>
    {/if}

  {:else if activeTab === 'Databases'}
    <div class="card overflow-hidden">
      <table class="table-wrap w-full">
        <thead><tr><th>Database</th><th>Size</th><th>Connections</th><th>Committed</th><th>Rolled Back</th><th>Hit Ratio</th></tr></thead>
        <tbody>
          {#each dbStats as db}
            <tr>
              <td class="font-medium">{db.database}</td>
              <td class="font-mono text-xs">{formatBytes(db.size_bytes)}</td>
              <td>{db.connections}</td>
              <td class="font-mono text-xs">{db.transactions_committed?.toLocaleString() || '-'}</td>
              <td class="font-mono text-xs">{db.transactions_rolled_back?.toLocaleString() || '-'}</td>
              <td>{(db.hit_ratio * 100).toFixed(1)}%</td>
            </tr>
          {:else}
            <tr><td colspan="6" class="text-center py-8" style="color: var(--text-tertiary);">No database stats</td></tr>
          {/each}
        </tbody>
      </table>
    </div>

  {:else if activeTab === 'Tables'}
    <div class="card overflow-hidden">
      <table class="table-wrap w-full">
        <thead><tr><th>Schema</th><th>Table</th><th>Seq Scan</th><th>Idx Scan</th><th>Inserts</th><th>Updates</th><th>Deletes</th><th>Live</th><th>Dead</th></tr></thead>
        <tbody>
          {#each tableStats as t}
            <tr>
              <td>{t.schema_name}</td>
              <td class="font-medium">{t.table_name}</td>
              <td class="font-mono text-xs">{t.seq_scan?.toLocaleString() || '-'}</td>
              <td class="font-mono text-xs">{t.idx_scan?.toLocaleString() || '-'}</td>
              <td class="font-mono text-xs">{t.n_tup_ins?.toLocaleString() || '-'}</td>
              <td class="font-mono text-xs">{t.n_tup_upd?.toLocaleString() || '-'}</td>
              <td class="font-mono text-xs">{t.n_tup_del?.toLocaleString() || '-'}</td>
              <td class="font-mono text-xs">{t.n_live_tup?.toLocaleString() || '-'}</td>
              <td class="font-mono text-xs">{t.n_dead_tup?.toLocaleString() || '-'}</td>
            </tr>
          {:else}
            <tr><td colspan="9" class="text-center py-8" style="color: var(--text-tertiary);">No table stats</td></tr>
          {/each}
        </tbody>
      </table>
    </div>

  {:else if activeTab === 'Indexes'}
    <div class="card overflow-hidden">
      <table class="table-wrap w-full">
        <thead><tr><th>Schema</th><th>Table</th><th>Index</th><th>Scans</th><th>Tuples Read</th><th>Size</th><th>Unique</th></tr></thead>
        <tbody>
          {#each indexStats as idx}
            <tr>
              <td>{idx.schema_name}</td>
              <td>{idx.table_name}</td>
              <td class="font-medium">{idx.index_name}</td>
              <td class="font-mono text-xs">{idx.idx_scan?.toLocaleString() || '-'}</td>
              <td class="font-mono text-xs">{idx.idx_tup_read?.toLocaleString() || '-'}</td>
              <td class="font-mono text-xs">{formatBytes(idx.size_bytes)}</td>
              <td>{idx.unique ? '✓' : '✗'}</td>
            </tr>
          {:else}
            <tr><td colspan="7" class="text-center py-8" style="color: var(--text-tertiary);">No index stats</td></tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<ConfirmDialog
  open={terminatePid !== null}
  title="Terminate Session"
  description={'Are you sure you want to terminate session PID ' + terminatePid + '?'}
  confirmLabel="Terminate"
  variant="danger"
  onconfirm={handleTerminate}
  oncancel={() => terminatePid = null}
/>

<ConfirmDialog
  open={cancelPid !== null}
  title="Cancel Query"
  description={'Cancel query for PID ' + cancelPid + '?'}
  confirmLabel="Cancel"
  variant="primary"
  onconfirm={handleCancel}
  oncancel={() => cancelPid = null}
/>
