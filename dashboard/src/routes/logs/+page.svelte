<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';
  import type { LogEntry } from '$lib/types';
  import Card from '$lib/components/Card.svelte';
  import Pagination from '$lib/components/Pagination.svelte';
  import Badge from '$lib/components/Badge.svelte';
  import LoadingCard from '$lib/components/LoadingCard.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';

  const tabs = ['All', 'Query', 'Error', 'Auth', 'Connection'] as const;
  type Tab = typeof tabs[number];

  let loading = $state(false);
  let error = $state('');
  let activeTab = $state<Tab>('All');
  let logs = $state<LogEntry[]>([]);
  let total = $state(0);
  let limit = $state(50);
  let offset = $state(0);

  let severityFilter = $state('ALL');
  let dbFilter = $state('');
  let userFilter = $state('');
  let dateFrom = $state('');
  let dateTo = $state('');
  let autoRefresh = $state(false);
  let refreshInterval: ReturnType<typeof setInterval> | null = null;

  onMount(() => { loadLogs(); });

  $effect(() => {
    if (autoRefresh && !refreshInterval) {
      refreshInterval = setInterval(loadLogs, 5000);
    } else if (!autoRefresh && refreshInterval) {
      clearInterval(refreshInterval);
      refreshInterval = null;
    }
  });

  async function loadLogs() {
    loading = true;
    error = '';
    try {
      const result = await api.getLogs(limit, offset, severityFilter !== 'ALL' ? severityFilter : undefined);
      logs = result || [];
      total = 0;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load logs';
    } finally {
      loading = false;
    }
  }

  function handleSearch() {
    offset = 0;
    loadLogs();
  }

  function handlePage(newOffset: number) {
    offset = newOffset;
    loadLogs();
  }

  function severityColor(level: string): string {
    switch (level.toUpperCase()) {
      case 'ERROR': return 'var(--danger)';
      case 'WARNING': return 'var(--warning)';
      case 'INFO': return 'var(--accent)';
      case 'LOG': return 'var(--text-secondary)';
      default: return 'var(--text-tertiary)';
    }
  }

  function severityBg(level: string): string {
    switch (level.toUpperCase()) {
      case 'ERROR': return 'rgba(239,68,68,0.1)';
      case 'WARNING': return 'rgba(245,158,11,0.1)';
      case 'INFO': return 'rgba(59,130,246,0.1)';
      default: return 'rgba(107,114,128,0.1)';
    }
  }

  function formatTimestamp(ts: string): string {
    return new Date(ts).toLocaleString();
  }
</script>

<div class="max-w-6xl mx-auto">
  <div class="flex items-center justify-between mb-6">
    <h1 class="text-2xl font-bold" style="color: var(--text-primary);">Logs</h1>
    <label class="flex items-center gap-2 text-xs cursor-pointer" style="color: var(--text-secondary);">
      <input type="checkbox" bind:checked={autoRefresh} />
      Auto-refresh
    </label>
  </div>

  <div class="card p-4 mb-4">
    <div class="flex flex-wrap items-end gap-3">
      <div>
        <label class="block text-xs font-medium mb-1" style="color: var(--text-tertiary);">Severity</label>
        <select bind:value={severityFilter} class="input text-sm">
          <option>ALL</option><option>ERROR</option><option>WARNING</option><option>INFO</option><option>LOG</option>
        </select>
      </div>
      <div>
        <label class="block text-xs font-medium mb-1" style="color: var(--text-tertiary);">Database</label>
        <input type="text" bind:value={dbFilter} class="input text-sm" placeholder="Filter by database" />
      </div>
      <div>
        <label class="block text-xs font-medium mb-1" style="color: var(--text-tertiary);">User</label>
        <input type="text" bind:value={userFilter} class="input text-sm" placeholder="Filter by user" />
      </div>
      <div>
        <label class="block text-xs font-medium mb-1" style="color: var(--text-tertiary);">From</label>
        <input type="datetime-local" bind:value={dateFrom} class="input text-sm" />
      </div>
      <div>
        <label class="block text-xs font-medium mb-1" style="color: var(--text-tertiary);">To</label>
        <input type="datetime-local" bind:value={dateTo} class="input text-sm" />
      </div>
      <button onclick={handleSearch} class="btn btn-primary btn-sm">Search</button>
    </div>
  </div>

  <div class="flex gap-0 border-b mb-4 overflow-x-auto" style="border-color: var(--border-primary);">
    {#each tabs as tab}
      <button
        onclick={() => { activeTab = tab; offset = 0; loadLogs(); }}
        class="tab"
        class:active={activeTab === tab}
      >{tab}</button>
    {/each}
  </div>

  {#if error}
    <div class="card p-4 mb-4" style="border-color: rgba(239,68,68,0.3);">
      <p class="text-sm" style="color: var(--danger);">{error}</p>
    </div>
  {/if}

  {#if loading}
    <LoadingCard />
  {:else if logs.length === 0}
    <EmptyState title="No log entries" description="Try adjusting your filters" />
  {:else}
    <div class="card overflow-hidden">
      <table class="table-wrap w-full">
        <thead>
          <tr>
            <th style="width: 80px;">Level</th>
            <th style="width: 180px;">Timestamp</th>
            <th>Message</th>
            <th style="width: 100px;">Database</th>
            <th style="width: 100px;">User</th>
            <th style="width: 60px;">PID</th>
          </tr>
        </thead>
        <tbody>
          {#each logs as log}
            <tr>
              <td>
                <span class="badge" style="background-color: {severityBg(log.level)}; color: {severityColor(log.level)};">
                  {log.level}
                </span>
              </td>
              <td class="text-xs whitespace-nowrap">{formatTimestamp(log.timestamp)}</td>
              <td class="text-xs max-w-md truncate font-mono">{log.message}</td>
              <td class="text-xs">{log.database || '-'}</td>
              <td class="text-xs">{log.user || '-'}</td>
              <td class="font-mono text-xs">{log.pid || '-'}</td>
            </tr>
          {/each}
        </tbody>
      </table>
      <Pagination {total} {limit} {offset} onpage={handlePage} />
    </div>
  {/if}
</div>
