<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { api } from '$lib/api/client';
  import type { ColumnInfo, TableRowResponse } from '$lib/types';
  import Card from '$lib/components/Card.svelte';
  import LoadingCard from '$lib/components/LoadingCard.svelte';
  import Skeleton from '$lib/components/Skeleton.svelte';
  import Pagination from '$lib/components/Pagination.svelte';
  import Modal from '$lib/components/Modal.svelte';
  import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
  import Badge from '$lib/components/Badge.svelte';
  import Breadcrumbs from '$lib/components/Breadcrumbs.svelte';

  let schema = $derived($page.params.schema!);
  let table = $derived($page.params.table!);

  let loading = $state(true);
  let error = $state('');
  let columns = $state<ColumnInfo[]>([]);
  let rows = $state<Record<string, unknown>[]>([]);
  let total = $state(0);
  let limit = $state(50);
  let offset = $state(0);
  let sortCol = $state('');
  let sortOrder = $state<'asc' | 'desc'>('asc');
  let searchQuery = $state('');
  let activeTab = $state<'Browse' | 'Structure' | 'Info'>('Browse');

  let insertModal = $state(false);
  let editModal = $state(false);
  let deleteModal = $state(false);
  let selectedRow = $state<Record<string, unknown> | null>(null);
  let editValues = $state<Record<string, string>>({});
  let insertValues = $state<Record<string, string>>({});
  let bulkDeleteIds = $state<unknown[]>([]);
  let bulkMode = $state(false);
  let saveLoading = $state(false);

  let tableInfo = $state<Record<string, unknown> | null>(null);

  onMount(() => { loadData(); });

  async function loadData() {
    loading = true;
    error = '';
    try {
      const [details, data] = await Promise.all([
        api.getTableDetails(schema, table),
        api.queryTable(schema, table, limit, offset, sortCol || undefined, sortOrder),
      ]);
      columns = details.columns;
      tableInfo = details.info as unknown as Record<string, unknown>;
      rows = data.rows;
      total = data.total;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load table data';
    } finally {
      loading = false;
    }
  }

  async function loadRows() {
    try {
      const data = await api.queryTable(schema, table, limit, offset, sortCol || undefined, sortOrder);
      rows = data.rows;
      total = data.total;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load rows';
    }
  }

  function handleSort(col: string) {
    if (sortCol === col) {
      sortOrder = sortOrder === 'asc' ? 'desc' : 'asc';
    } else {
      sortCol = col;
      sortOrder = 'asc';
    }
    offset = 0;
    loadRows();
  }

  function handlePage(newOffset: number) {
    offset = newOffset;
    loadRows();
  }

  function handleSearch() {
    offset = 0;
    loadRows();
  }

  function openEdit(row: Record<string, unknown>) {
    selectedRow = row;
    editValues = {};
    columns.forEach(c => { editValues[c.name] = String(row[c.name] ?? ''); });
    editModal = true;
  }

  function openDelete(row: Record<string, unknown>) {
    selectedRow = row;
    deleteModal = true;
  }

  function openInsert() {
    insertValues = {};
    columns.forEach(c => { insertValues[c.name] = ''; });
    insertModal = true;
  }

  async function handleInsert() {
    saveLoading = true;
    try {
      await api.insertRow(schema, table, insertValues as unknown as Record<string, unknown>);
      insertModal = false;
      loadRows();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Insert failed';
    } finally {
      saveLoading = false;
    }
  }

  async function handleEdit() {
    if (!selectedRow) return;
    saveLoading = true;
    try {
      const where: Record<string, unknown> = {};
      const pkCol = columns.find(c => c.is_pk);
      if (pkCol) where[pkCol.name] = selectedRow[pkCol.name];
      await api.updateRow(schema, table, editValues as unknown as Record<string, unknown>, where);
      editModal = false;
      loadRows();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Update failed';
    } finally {
      saveLoading = false;
    }
  }

  async function handleDelete() {
    if (!selectedRow) return;
    saveLoading = true;
    try {
      const where: Record<string, unknown> = {};
      const pkCol = columns.find(c => c.is_pk);
      if (pkCol) where[pkCol.name] = selectedRow[pkCol.name];
      await api.deleteRow(schema, table, where);
      deleteModal = false;
      loadRows();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Delete failed';
    } finally {
      saveLoading = false;
    }
  }

  function toggleBulkSelect(id: unknown) {
    if (bulkDeleteIds.includes(id)) {
      bulkDeleteIds = bulkDeleteIds.filter(i => i !== id);
    } else {
      bulkDeleteIds = [...bulkDeleteIds, id];
    }
  }

  async function handleBulkDelete() {
    if (bulkDeleteIds.length === 0) return;
    saveLoading = true;
    try {
      await api.bulkDelete(schema, table, bulkDeleteIds);
      bulkDeleteIds = [];
      bulkMode = false;
      loadRows();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Bulk delete failed';
    } finally {
      saveLoading = false;
    }
  }

  function formatBytes(bytes: number): string {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  }
</script>

<div class="max-w-6xl mx-auto">
  <Breadcrumbs items={[
    { label: 'Explorer', href: '/explorer' },
    { label: schema, href: '/explorer' },
    { label: table },
  ]} />

  {#if error}
    <div class="card p-6 text-center">
      <p style="color: var(--danger);">{error}</p>
      <button onclick={loadData} class="btn btn-primary mt-3">Retry</button>
    </div>
  {:else if loading}
    <LoadingCard />
    <Skeleton rows={8} />
  {:else}
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold" style="color: var(--text-primary);">{table}</h1>
        <p class="text-sm mt-0.5" style="color: var(--text-secondary);">{schema}.{table} • {total.toLocaleString()} rows • {columns.length} columns</p>
      </div>
      <div class="flex items-center gap-2">
        <button onclick={openInsert} class="btn btn-primary btn-sm">+ Insert</button>
        <button onclick={() => bulkMode = !bulkMode} class="btn btn-secondary btn-sm">Bulk</button>
      </div>
    </div>

    <div class="flex gap-0 border-b mb-4" style="border-color: var(--border-primary);">
      {#each ['Browse', 'Structure', 'Info'] as tab}
        <button
          onclick={() => activeTab = tab as typeof activeTab}
          class="tab"
          class:active={activeTab === tab}
        >{tab}</button>
      {/each}
    </div>

    {#if activeTab === 'Browse'}
      <div class="card overflow-hidden">
        <div class="flex items-center gap-3 p-4 border-b" style="border-color: var(--border-primary);">
          <input
            type="text"
            bind:value={searchQuery}
            placeholder="Search within table..."
            class="input flex-1"
          />
          <button onclick={handleSearch} class="btn btn-primary btn-sm">Search</button>
        </div>

        {#if bulkMode && bulkDeleteIds.length > 0}
          <div class="flex items-center gap-3 px-4 py-2" style="background-color: rgba(239,68,68,0.1);">
            <span class="text-sm">{bulkDeleteIds.length} selected</span>
            <button onclick={handleBulkDelete} class="btn btn-danger btn-sm">Delete Selected</button>
            <button onclick={() => { bulkDeleteIds = []; bulkMode = false; }} class="btn btn-ghost btn-sm">Cancel</button>
          </div>
        {/if}

        <div class="overflow-x-auto">
          <table class="table-wrap w-full">
            <thead>
              <tr>
                {#if bulkMode}<th style="width: 40px;"></th>{/if}
                {#each columns as col}
                  <th
                    onclick={() => handleSort(col.name)}
                    class="cursor-pointer select-none"
                  >
                    {col.name}
                    {#if sortCol === col.name}
                      <span class="ml-1">{sortOrder === 'asc' ? '↑' : '↓'}</span>
                    {/if}
                  </th>
                {/each}
                <th style="width: 100px;">Actions</th>
              </tr>
            </thead>
            <tbody>
              {#each rows as row}
                <tr>
                  {#if bulkMode}
                    <td>
                      <input
                        type="checkbox"
                        checked={bulkDeleteIds.includes(row[columns.find(c => c.is_pk)?.name || ''])}
                        onchange={() => toggleBulkSelect(row[columns.find(c => c.is_pk)?.name || ''])}
                      />
                    </td>
                  {/if}
                  {#each columns as col}
                    <td class="truncate max-w-xs font-mono text-xs">{String(row[col.name] ?? '')}</td>
                  {/each}
                  <td>
                    <div class="flex gap-1">
                      <button onclick={() => openEdit(row)} class="btn btn-ghost btn-sm" title="Edit">✎</button>
                      <button onclick={() => openDelete(row)} class="btn btn-ghost btn-sm" style="color: var(--danger);" title="Delete">✕</button>
                    </div>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>

        <Pagination {total} {limit} {offset} onpage={handlePage} />
      </div>

    {:else if activeTab === 'Structure'}
      <div class="card overflow-hidden">
        <table class="table-wrap w-full">
          <thead>
            <tr>
              <th>Column</th>
              <th>Type</th>
              <th>Nullable</th>
              <th>Default</th>
              <th>PK</th>
            </tr>
          </thead>
          <tbody>
            {#each columns as col}
              <tr>
                <td class="font-medium">{col.name}</td>
                <td class="font-mono text-xs">{col.data_type}</td>
                <td>{col.is_nullable ? 'Yes' : 'No'}</td>
                <td class="font-mono text-xs">{col.default_value || '-'}</td>
                <td>{col.is_pk ? '✓' : ''}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>

    {:else if activeTab === 'Info'}
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card title="Row Count"><p class="text-2xl font-bold">{total?.toLocaleString() || '-'}</p></Card>
        <Card title="Columns"><p class="text-2xl font-bold">{columns.length}</p></Card>
        <Card title="Size"><p class="text-2xl font-bold">{formatBytes((tableInfo as any)?.size_bytes || 0)}</p></Card>
        <Card title="Has PK"><p class="text-2xl font-bold">{(tableInfo as any)?.has_pk ? 'Yes' : 'No'}</p></Card>
      </div>
    {/if}
  {/if}
</div>

<Modal title="Insert Row" open={insertModal} onclose={() => insertModal = false}>
  {#each columns as col}
    <div class="mb-3">
      <label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">{col.name}</label>
      <input type="text" bind:value={insertValues[col.name]} class="input" placeholder={col.data_type} />
    </div>
  {/each}
  <div class="flex justify-end gap-3 mt-4">
    <button onclick={() => insertModal = false} class="btn btn-secondary">Cancel</button>
    <button onclick={handleInsert} disabled={saveLoading} class="btn btn-primary">{saveLoading ? 'Inserting...' : 'Insert'}</button>
  </div>
</Modal>

<Modal title="Edit Row" open={editModal} onclose={() => editModal = false}>
  {#each columns as col}
    <div class="mb-3">
      <label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">{col.name}</label>
      <input type="text" bind:value={editValues[col.name]} class="input" disabled={col.is_pk} />
    </div>
  {/each}
  <div class="flex justify-end gap-3 mt-4">
    <button onclick={() => editModal = false} class="btn btn-secondary">Cancel</button>
    <button onclick={handleEdit} disabled={saveLoading} class="btn btn-primary">{saveLoading ? 'Saving...' : 'Save'}</button>
  </div>
</Modal>

<ConfirmDialog
  open={deleteModal}
  title="Delete Row"
  description="Are you sure you want to delete this row? This action cannot be undone."
  confirmLabel="Delete"
  variant="danger"
  onconfirm={handleDelete}
  oncancel={() => deleteModal = false}
  loading={saveLoading}
/>
