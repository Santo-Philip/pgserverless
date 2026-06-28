<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';
  import type { AuditLog } from '$lib/types';
  import Card from '$lib/components/Card.svelte';
  import Pagination from '$lib/components/Pagination.svelte';
  import Badge from '$lib/components/Badge.svelte';
  import LoadingCard from '$lib/components/LoadingCard.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';

  let loading = $state(false);
  let error = $state('');
  let logs = $state<AuditLog[]>([]);
  let total = $state(0);
  let limit = $state(50);
  let offset = $state(0);

  let actionFilter = $state('');
  let resourceFilter = $state('');
  let userFilter = $state('');
  let dateFrom = $state('');
  let dateTo = $state('');

  let expandedId = $state<string | null>(null);

  function groupByDate(items: AuditLog[]): { date: string; entries: AuditLog[] }[] {
    const groups: Record<string, AuditLog[]> = {};
    items.forEach(item => {
      const date = new Date(item.created_at).toLocaleDateString();
      if (!groups[date]) groups[date] = [];
      groups[date].push(item);
    });
    return Object.entries(groups).map(([date, entries]) => ({ date, entries }));
  }

  let grouped = $derived(groupByDate(logs));

  onMount(() => { loadLogs(); });

  async function loadLogs() {
    loading = true;
    error = '';
    try {
      const result = await api.listAuditLogs(limit, offset);
      logs = result || [];
      total = (result as any)?.total || logs.length;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load audit logs';
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

  function toggleExpand(id: string) {
    expandedId = expandedId === id ? null : id;
  }

  function actionColor(action: string): string {
    if (action.includes('DELETE') || action.includes('DROP')) return 'var(--danger)';
    if (action.includes('CREATE') || action.includes('INSERT')) return 'var(--success)';
    if (action.includes('UPDATE') || action.includes('ALTER')) return 'var(--warning)';
    return 'var(--accent)';
  }
</script>

<div class="max-w-6xl mx-auto">
  <h1 class="text-2xl font-bold mb-6" style="color: var(--text-primary);">Audit Logs</h1>

  <div class="card p-4 mb-4">
    <div class="flex flex-wrap items-end gap-3">
      <div>
        <label class="block text-xs font-medium mb-1" style="color: var(--text-tertiary);">Action</label>
        <input type="text" bind:value={actionFilter} class="input text-sm" placeholder="Filter by action" />
      </div>
      <div>
        <label class="block text-xs font-medium mb-1" style="color: var(--text-tertiary);">Resource</label>
        <input type="text" bind:value={resourceFilter} class="input text-sm" placeholder="Filter by resource" />
      </div>
      <div>
        <label class="block text-xs font-medium mb-1" style="color: var(--text-tertiary);">User</label>
        <input type="text" bind:value={userFilter} class="input text-sm" placeholder="Filter by user" />
      </div>
      <div>
        <label class="block text-xs font-medium mb-1" style="color: var(--text-tertiary);">From</label>
        <input type="date" bind:value={dateFrom} class="input text-sm" />
      </div>
      <div>
        <label class="block text-xs font-medium mb-1" style="color: var(--text-tertiary);">To</label>
        <input type="date" bind:value={dateTo} class="input text-sm" />
      </div>
      <button onclick={handleSearch} class="btn btn-primary btn-sm">Search</button>
    </div>
  </div>

  {#if error}
    <div class="card p-4 mb-4" style="border-color: rgba(239,68,68,0.3);">
      <p class="text-sm" style="color: var(--danger);">{error}</p>
    </div>
  {/if}

  {#if loading}
    <LoadingCard />
  {:else if logs.length === 0}
    <EmptyState title="No audit log entries" description="Audit events will appear here" />
  {:else}
    <div class="space-y-6">
      {#each grouped as group}
        <div>
          <h3 class="text-sm font-semibold mb-3" style="color: var(--text-secondary);">{group.date}</h3>
          <div class="space-y-2">
            {#each group.entries as log}
              <div class="card card-hover overflow-hidden">
                <button
                  onclick={() => toggleExpand(log.id)}
                  class="w-full text-left p-4 flex items-start justify-between gap-4"
                >
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center gap-2 mb-1">
                      <span
                        class="badge text-xs"
                        style="background-color: {actionColor(log.action)}22; color: {actionColor(log.action)};"
                      >{log.action}</span>
                      <span class="text-xs font-medium" style="color: var(--text-primary);">{log.resource}</span>
                    </div>
                    <div class="text-xs" style="color: var(--text-secondary);">
                      by <span class="font-medium">{log.actor_name || log.actor_id}</span>
                      <span class="mx-1">•</span>
                      {new Date(log.created_at).toLocaleTimeString()}
                    </div>
                  </div>
                  <div class="flex items-center gap-2 flex-shrink-0">
                    <span class="text-xs" style="color: var(--text-tertiary);">{log.ip_address}</span>
                    <span class="text-xs transition-transform" style="color: var(--text-tertiary); transform: {expandedId === log.id ? 'rotate(180deg)' : 'none'};">▼</span>
                  </div>
                </button>

                {#if expandedId === log.id}
                  <div class="px-4 pb-4 border-t" style="border-color: var(--border-primary);">
                    <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 mt-3 text-xs">
                      <div>
                        <span class="block font-medium mb-0.5" style="color: var(--text-tertiary);">Resource ID</span>
                        <span style="color: var(--text-primary);">{log.resource_id}</span>
                      </div>
                      <div>
                        <span class="block font-medium mb-0.5" style="color: var(--text-tertiary);">User Agent</span>
                        <span style="color: var(--text-primary);" class="truncate block">{log.user_agent}</span>
                      </div>
                    </div>
                    {#if log.metadata && Object.keys(log.metadata).length > 0}
                      <div class="mt-3">
                        <span class="block text-xs font-medium mb-1" style="color: var(--text-tertiary);">Metadata</span>
                        <pre class="text-xs font-mono rounded-lg p-3 overflow-x-auto" style="background-color: var(--bg-tertiary); color: var(--text-secondary);">{JSON.stringify(log.metadata, null, 2)}</pre>
                      </div>
                    {/if}
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        </div>
      {/each}
    </div>

    <div class="mt-4">
      <Pagination {total} {limit} {offset} onpage={handlePage} />
    </div>
  {/if}
</div>
