<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';
  import type { ExecuteResponse, SavedQuery, QueryHistory } from '$lib/types';
  import Card from '$lib/components/Card.svelte';
  import Modal from '$lib/components/Modal.svelte';
  import LoadingCard from '$lib/components/LoadingCard.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import CopyButton from '$lib/components/CopyButton.svelte';

  let query = $state('');
  let loading = $state(false);
  let error = $state('');
  let result = $state<ExecuteResponse | null>(null);
  let explainResult = $state<Record<string, unknown>[] | null>(null);
  let explainModal = $state(false);
  let duration = $state(0);
  let history = $state<QueryHistory[]>([]);
  let savedQueries = $state<SavedQuery[]>([]);
  let historyOffset = $state(0);
  let activeSidebar = $state<'history' | 'saved'>('history');
  let saveModal = $state(false);
  let saveName = $state('');
  let saveLoading = $state(false);

  const limit = 50;

  onMount(async () => {
    try {
      const [h, s] = await Promise.all([
        api.getQueryHistory(limit, 0),
        api.getSavedQueries(),
      ]);
      history = h;
      savedQueries = s;
    } catch {}
  });

  async function execute() {
    if (!query.trim()) return;
    loading = true;
    error = '';
    result = null;
    explainResult = null;
    const start = performance.now();
    try {
      const res = await api.executeSQL({ query });
      result = res;
      duration = res.duration_ms;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Query execution failed';
    } finally {
      loading = false;
    }
  }

  async function handleExplain() {
    if (!query.trim()) return;
    loading = true;
    error = '';
    explainResult = null;
    try {
      const res = await api.explainQuery({ query });
      explainResult = [res.plan];
      explainModal = true;
    } catch (e) {
      error = e instanceof Error ? e.message : 'EXPLAIN failed';
    } finally {
      loading = false;
    }
  }

  async function handleCancel() {
    if (!result?.columns) return;
    loading = true;
    try {
      await api.cancelQuery(0);
    } catch {}
    loading = false;
  }

  async function handleSave() {
    if (!saveName.trim() || !query.trim()) return;
    saveLoading = true;
    try {
      await api.saveQuery({ name: saveName, query, database: '' });
      saveModal = false;
      saveName = '';
      savedQueries = await api.getSavedQueries();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to save query';
    } finally {
      saveLoading = false;
    }
  }

  function loadSaved(q: SavedQuery) {
    query = q.query;
  }

  function loadHistory(h: QueryHistory) {
    query = h.query;
  }

  async function deleteSaved(id: string) {
    try {
      await api.deleteSavedQuery(id);
      savedQueries = savedQueries.filter(s => s.id !== id);
    } catch {}
  }

  function exportCSV() {
    if (!result || !result.columns) return;
    const headers = result.columns.map(c => c.name);
    const csvRows = [headers.join(',')];
    result.rows.forEach(row => {
      csvRows.push(headers.map(h => {
        const val = row[h];
        const s = String(val ?? '');
        return s.includes(',') || s.includes('"') ? `"${s.replace(/"/g, '""')}"` : s;
      }).join(','));
    });
    const blob = new Blob([csvRows.join('\n')], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'query_results.csv';
    a.click();
    URL.revokeObjectURL(url);
  }

  function handleKeydown(e: KeyboardEvent) {
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      e.preventDefault();
      execute();
    }
  }
</script>

<div class="max-w-6xl mx-auto">
  <div class="flex items-center justify-between mb-6">
    <h1 class="text-2xl font-bold" style="color: var(--text-primary);">SQL Workspace</h1>
    <div class="flex items-center gap-2">
      <button onclick={() => saveModal = true} class="btn btn-secondary btn-sm" disabled={!query.trim()}>Save</button>
      <button onclick={() => activeSidebar = activeSidebar === 'history' ? 'saved' : 'history'} class="btn btn-ghost btn-sm">History</button>
    </div>
  </div>

  <div class="flex flex-col lg:flex-row gap-6">
    <div class="flex-1 min-w-0">
      <div class="card p-4 mb-4">
        <textarea
          bind:value={query}
          onkeydown={handleKeydown}
          class="w-full bg-transparent text-sm font-mono resize-none"
          style="color: var(--text-primary); min-height: 160px; border: none; outline: none;"
          placeholder="Enter SQL query... (Cmd/Ctrl + Enter to execute)"
          spellcheck="false"
        ></textarea>
      </div>

      <div class="flex items-center justify-between mb-4">
        <div class="flex items-center gap-2">
          <button onclick={execute} disabled={loading || !query.trim()} class="btn btn-primary">
            {loading ? 'Running...' : 'Execute'}
          </button>
          <button onclick={handleExplain} disabled={loading || !query.trim()} class="btn btn-secondary">Explain</button>
          {#if loading}
            <button onclick={handleCancel} class="btn btn-ghost" style="color: var(--danger);">Cancel</button>
          {/if}
        </div>
        {#if result}
          <div class="flex items-center gap-3 text-xs" style="color: var(--text-tertiary);">
            <span>{result.row_count} rows</span>
            <span>•</span>
            <span>{duration}ms</span>
            <button onclick={exportCSV} class="btn btn-ghost btn-sm">Export CSV</button>
          </div>
        {/if}
      </div>

      {#if error}
        <div class="card p-4 mb-4" style="border-color: rgba(239,68,68,0.3);">
          <p class="text-sm font-mono" style="color: var(--danger);">{error}</p>
        </div>
      {/if}

      {#if loading}
        <LoadingCard />
      {/if}

      {#if result && result.columns}
        <div class="card overflow-hidden">
          <div class="overflow-x-auto">
            <table class="table-wrap w-full">
              <thead>
                <tr>
                  {#each result.columns as col}
                    <th>{col.name}</th>
                  {/each}
                </tr>
              </thead>
              <tbody>
                {#each result.rows as row}
                  <tr>
                    {#each result.columns as col}
                      <td class="font-mono text-xs truncate max-w-xs">{String(row[col.name] ?? '')}</td>
                    {/each}
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </div>
      {/if}
    </div>

    {#if activeSidebar}
      <div class="w-full lg:w-72 flex-shrink-0">
        <div class="card">
          <div class="flex border-b" style="border-color: var(--border-primary);">
            <button
              onclick={() => activeSidebar = 'history'}
              class="tab flex-1 text-center"
              class:active={activeSidebar === 'history'}
            >History</button>
            <button
              onclick={() => activeSidebar = 'saved'}
              class="tab flex-1 text-center"
              class:active={activeSidebar === 'saved'}
            >Saved</button>
          </div>
          <div class="p-3 max-h-[400px] overflow-y-auto">
            {#if activeSidebar === 'history'}
              {#if history.length === 0}
                <EmptyState title="No history" description="Run queries to see them here" />
              {:else}
                {#each history as h}
                  <button onclick={() => loadHistory(h)} class="w-full text-left p-2 rounded-lg text-xs mb-1" style="background-color: var(--bg-hover);">
                    <div class="truncate font-mono" style="color: var(--text-primary);">{h.query}</div>
                    <div class="mt-0.5" style="color: var(--text-tertiary);">{h.duration_ms}ms • {h.row_count} rows</div>
                  </button>
                {/each}
              {/if}
            {:else}
              {#if savedQueries.length === 0}
                <EmptyState title="No saved queries" description="Save your queries for later" />
              {:else}
                {#each savedQueries as sq}
                  <div class="flex items-center gap-2 p-2 rounded-lg mb-1" style="background-color: var(--bg-hover);">
                    <button onclick={() => loadSaved(sq)} class="flex-1 text-left text-xs">
                      <div class="font-medium" style="color: var(--text-primary);">{sq.name}</div>
                      <div class="truncate font-mono" style="color: var(--text-tertiary);">{sq.query}</div>
                    </button>
                    <button onclick={() => deleteSaved(sq.id)} class="text-xs" style="color: var(--danger);">✕</button>
                  </div>
                {/each}
              {/if}
            {/if}
          </div>
        </div>
      </div>
    {/if}
  </div>
</div>

<Modal title="Save Query" open={saveModal} onclose={() => saveModal = false}>
  <div class="mb-4">
    <label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">Name</label>
    <input type="text" bind:value={saveName} class="input" placeholder="My Query" />
  </div>
  <div class="flex justify-end gap-3">
    <button onclick={() => saveModal = false} class="btn btn-secondary">Cancel</button>
    <button onclick={handleSave} disabled={saveLoading || !saveName.trim()} class="btn btn-primary">{saveLoading ? 'Saving...' : 'Save'}</button>
  </div>
</Modal>

<Modal title="Explain Plan" open={explainModal} onclose={() => explainModal = false}>
  {#if explainResult}
    <pre class="text-xs font-mono whitespace-pre-wrap rounded-lg p-4 overflow-x-auto" style="background-color: var(--bg-tertiary); color: var(--text-secondary); max-height: 400px;">{JSON.stringify(explainResult, null, 2)}</pre>
  {/if}
  <div class="flex justify-end mt-4">
    <button onclick={() => explainModal = false} class="btn btn-secondary">Close</button>
  </div>
</Modal>
