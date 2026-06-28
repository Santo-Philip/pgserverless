<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';
  import type { SchemaInfo, TableInfo, ViewInfo, FunctionInfo, ProcedureInfo, TriggerInfo, IndexInfo, ConstraintInfo, SequenceInfo, MaterializedViewInfo } from '$lib/types';
  import { goto } from '$app/navigation';
  import Card from '$lib/components/Card.svelte';
  import LoadingCard from '$lib/components/LoadingCard.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';

  const tabs = ['Tables', 'Views', 'Functions', 'Procedures', 'Triggers', 'Indexes', 'Constraints', 'Sequences', 'Materialized Views'] as const;
  type Tab = typeof tabs[number];

  let loading = $state(true);
  let error = $state('');
  let schemas = $state<SchemaInfo[]>([]);
  let search = $state('');
  let selectedSchema = $state<string | null>(null);
  let activeTab = $state<Tab>('Tables');
  let schemaLoading = $state(false);

  let tables = $state<TableInfo[]>([]);
  let views = $state<ViewInfo[]>([]);
  let functions = $state<FunctionInfo[]>([]);
  let procedures = $state<ProcedureInfo[]>([]);
  let triggers = $state<TriggerInfo[]>([]);
  let indexes = $state<IndexInfo[]>([]);
  let constraints = $state<ConstraintInfo[]>([]);
  let sequences = $state<SequenceInfo[]>([]);
  let materializedViews = $state<MaterializedViewInfo[]>([]);

  let filteredSchemas = $derived(
    schemas.filter(s => s.schema_name.toLowerCase().includes(search.toLowerCase()))
  );

  function formatBytes(bytes: number): string {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  }

  onMount(async () => {
    try {
      schemas = await api.listSchemas();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load schemas';
    } finally {
      loading = false;
    }
  });

  async function selectSchema(name: string) {
    selectedSchema = name;
    schemaLoading = true;
    activeTab = 'Tables';
    try {
      const results = await Promise.all([
        api.listTables(name),
        api.listViews(name),
        api.listFunctions(name),
        api.listProcedures(name),
        api.listTriggers(name),
        api.listIndexes(name),
        api.listConstraints(name),
        api.listSequences(name),
        api.listMaterializedViews(name),
      ]);
      tables = results[0] || [];
      views = results[1] || [];
      functions = results[2] || [];
      procedures = results[3] || [];
      triggers = results[4] || [];
      indexes = results[5] || [];
      constraints = results[6] || [];
      sequences = results[7] || [];
      materializedViews = results[8] || [];
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load schema details';
    } finally {
      schemaLoading = false;
    }
  }

  function currentItems() {
    switch (activeTab) {
      case 'Tables': return tables;
      case 'Views': return views;
      case 'Functions': return functions;
      case 'Procedures': return procedures;
      case 'Triggers': return triggers;
      case 'Indexes': return indexes;
      case 'Constraints': return constraints;
      case 'Sequences': return sequences;
      case 'Materialized Views': return materializedViews;
    }
  }

  function handleItemClick(item: unknown) {
    if (activeTab === 'Tables') {
      const t = item as unknown as TableInfo;
      goto(`/tables/${t.schema_name}/${t.table_name}`);
    }
  }
</script>

<div class="max-w-6xl mx-auto">
  <h1 class="text-2xl font-bold mb-6" style="color: var(--text-primary);">Database Explorer</h1>

  {#if error}
    <div class="card p-6 text-center mb-4">
      <p style="color: var(--danger);">{error}</p>
      <button onclick={() => window.location.reload()} class="btn btn-primary mt-3">Retry</button>
    </div>
  {/if}

  <div class="flex flex-col lg:flex-row gap-6">
    <div class="w-full lg:w-72 flex-shrink-0">
      <div class="card p-4">
        <input type="text" bind:value={search} placeholder="Search schemas..." class="input mb-3" />
        {#if loading}
          {#each [1,2,3] as _}
            <div class="skeleton h-12 w-full mb-2 rounded-lg"></div>
          {/each}
        {:else if filteredSchemas.length === 0}
          <EmptyState title="No schemas found" />
        {:else}
          <div class="space-y-1">
            {#each filteredSchemas as schema}
              <button
                onclick={() => selectSchema(schema.schema_name)}
                class="w-full text-left p-3 rounded-lg text-sm transition-colors"
                style="background-color: {selectedSchema === schema.schema_name ? 'var(--accent-muted)' : 'transparent'}; color: {selectedSchema === schema.schema_name ? 'var(--accent)' : 'var(--text-secondary)'};"
              >
                <div class="font-medium">{schema.schema_name}</div>
                <div class="text-xs mt-0.5" style="color: var(--text-tertiary);">{formatBytes(schema.size_bytes)} • {schema.table_count} tables</div>
              </button>
            {/each}
          </div>
        {/if}
      </div>
    </div>

    <div class="flex-1 min-w-0">
      {#if !selectedSchema}
        <div class="card p-12 text-center">
          <div class="text-4xl mb-3">▦</div>
          <h3 class="text-base font-semibold" style="color: var(--text-secondary);">Select a Schema</h3>
          <p class="text-sm mt-1" style="color: var(--text-tertiary);">Choose a schema from the left to explore its objects</p>
        </div>
      {:else}
        <div class="card">
          <div class="px-5 py-4 border-b" style="border-color: var(--border-primary);">
            <h2 class="text-lg font-semibold" style="color: var(--text-primary);">{selectedSchema}</h2>
          </div>

          <div class="flex gap-0 border-b overflow-x-auto" style="border-color: var(--border-primary);">
            {#each tabs as tab}
              <button
                onclick={() => activeTab = tab}
                class="tab whitespace-nowrap"
                class:active={activeTab === tab}
              >{tab}</button>
            {/each}
          </div>

          <div class="p-5">
            {#if schemaLoading}
              <LoadingCard />
            {:else}
              {#if currentItems().length === 0}
                <EmptyState title={'No ' + activeTab.toLowerCase()} description={'This schema has no ' + activeTab.toLowerCase() + ' defined'} />
              {:else}
                <div class="overflow-x-auto">
                  <table class="table-wrap w-full">
                    <thead>
                      <tr>
                        {#if activeTab === 'Tables'}
                          <th>Name</th><th>Rows</th><th>Columns</th><th>Size</th><th>PK</th>
                        {:else if activeTab === 'Views'}
                          <th>Name</th><th>Materialized</th>
                        {:else if activeTab === 'Functions'}
                          <th>Name</th><th>Return Type</th><th>Language</th>
                        {:else if activeTab === 'Procedures'}
                          <th>Name</th><th>Language</th>
                        {:else if activeTab === 'Triggers'}
                          <th>Name</th><th>Table</th><th>Event</th><th>Timing</th>
                        {:else if activeTab === 'Indexes'}
                          <th>Name</th><th>Table</th><th>Type</th><th>Unique</th><th>Size</th>
                        {:else if activeTab === 'Constraints'}
                          <th>Name</th><th>Table</th><th>Type</th><th>Columns</th>
                        {:else if activeTab === 'Sequences'}
                          <th>Name</th><th>Type</th><th>Start</th><th>Increment</th>
                        {:else if activeTab === 'Materialized Views'}
                          <th>Name</th><th>Rows</th><th>Size</th><th>Last Refresh</th>
                        {/if}
                      </tr>
                    </thead>
                    <tbody>
                      {#each currentItems() as item}
                        <tr
                          onclick={() => handleItemClick(item)}
                          onkeydown={(e) => e.key === 'Enter' && handleItemClick(item)}
                          tabindex={activeTab === 'Tables' ? 0 : -1}
                          class={activeTab === 'Tables' ? 'cursor-pointer' : ''}
                        >
                          {#if activeTab === 'Tables'}
                            {@const t = item as unknown as TableInfo}
                            <td class="font-medium" style="color: var(--accent);">{t.table_name}</td>
                            <td>{t.row_estimate?.toLocaleString() || '-'}</td>
                            <td>{t.column_count}</td>
                            <td>{formatBytes(t.size_bytes)}</td>
                            <td>{t.has_pk ? '✓' : '✗'}</td>
                          {:else if activeTab === 'Views'}
                            {@const v = item as unknown as ViewInfo}
                            <td class="font-medium">{v.view_name}</td>
                            <td>{v.is_materialized ? 'Yes' : 'No'}</td>
                          {:else if activeTab === 'Functions'}
                            {@const f = item as unknown as FunctionInfo}
                            <td class="font-medium">{f.function_name}</td>
                            <td class="font-mono text-xs">{f.return_type}</td>
                            <td>{f.language}</td>
                          {:else if activeTab === 'Procedures'}
                            {@const p = item as unknown as ProcedureInfo}
                            <td class="font-medium">{p.procedure_name}</td>
                            <td>{p.language}</td>
                          {:else if activeTab === 'Triggers'}
                            {@const t = item as unknown as TriggerInfo}
                            <td class="font-medium">{t.trigger_name}</td>
                            <td>{t.table_name}</td>
                            <td>{t.event}</td>
                            <td>{t.timing}</td>
                          {:else if activeTab === 'Indexes'}
                            {@const i = item as unknown as IndexInfo}
                            <td class="font-medium">{i.index_name}</td>
                            <td>{i.table_name}</td>
                            <td>{i.index_type}</td>
                            <td>{i.is_unique ? '✓' : '✗'}</td>
                            <td>{formatBytes(i.size_bytes)}</td>
                          {:else if activeTab === 'Constraints'}
                            {@const c = item as unknown as ConstraintInfo}
                            <td class="font-medium">{c.constraint_name}</td>
                            <td>{c.table_name}</td>
                            <td>{c.constraint_type}</td>
                            <td>{c.columns?.join(', ')}</td>
                          {:else if activeTab === 'Sequences'}
                            {@const s = item as unknown as SequenceInfo}
                            <td class="font-medium">{s.sequence_name}</td>
                            <td>{s.data_type}</td>
                            <td>{s.start_value}</td>
                            <td>{s.increment_by}</td>
                          {:else if activeTab === 'Materialized Views'}
                            {@const m = item as unknown as MaterializedViewInfo}
                            <td class="font-medium">{m.view_name}</td>
                            <td>{m.row_count?.toLocaleString() || '-'}</td>
                            <td>{formatBytes(m.size_bytes)}</td>
                            <td>{m.last_refresh || 'Never'}</td>
                          {/if}
                        </tr>
                      {/each}
                    </tbody>
                  </table>
                </div>
              {/if}
            {/if}
          </div>
        </div>
      {/if}
    </div>
  </div>
</div>
