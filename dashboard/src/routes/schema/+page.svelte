<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';
  import type { SchemaInfo, TableInfo, ColumnInfo, ConstraintInfo, IndexInfo, SequenceInfo } from '$lib/types';
  import Card from '$lib/components/Card.svelte';
  import Modal from '$lib/components/Modal.svelte';
  import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
  import LoadingCard from '$lib/components/LoadingCard.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import Alert from '$lib/components/Alert.svelte';

  let loading = $state(true);
  let error = $state('');
  let schemas = $state<SchemaInfo[]>([]);
  let selectedSchema = $state<string | null>(null);
  let schemaTables = $state<TableInfo[]>([]);
  let schemaColumns = $state<ColumnInfo[]>([]);
  let schemaSequences = $state<SequenceInfo[]>([]);
  let schemaLoading = $state(false);

  let createSchemaModal = $state(false);
  let createSchemaName = $state('');
  let dropSchemaName = $state('');
  let dropSchemaConfirm = $state(false);

  let createTableModal = $state(false);
  let createTableName = $state('');
  let createTableColumns = $state<{ name: string; type: string; nullable: boolean; is_pk: boolean; default: string }[]>([]);

  let addColumnModal = $state(false);
  let addColName = $state('');
  let addColType = $state('text');
  let addColNullable = $state(true);
  let addColDefault = $state('');

  let dropColName = $state('');
  let dropColConfirm = $state(false);

  let alterColModal = $state(false);
  let alterColOrigName = $state('');
  let alterColNewName = $state('');
  let alterColNewType = $state('');

  let addConstraintModal = $state(false);
  let constraintName = $state('');
  let constraintType = $state('PRIMARY KEY');
  let constraintCols = $state('');

  let dropConstraintName = $state('');
  let dropConstraintConfirm = $state(false);

  let createIndexModal = $state(false);
  let indexName = $state('');
  let indexCols = $state('');
  let indexUnique = $state(false);

  let dropIndexName = $state('');
  let dropIndexConfirm = $state(false);

  let createSeqModal = $state(false);
  let seqName = $state('');
  let seqStart = $state('1');
  let seqIncrement = $state('1');

  let dropSeqName = $state('');
  let dropSeqConfirm = $state(false);
  let ddlContent = $state('');
  let ddlModal = $state(false);
  let ddlLoading = $state(false);

  let loadingStates: Record<string, boolean> = $state({});

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

  async function loadSchemaDetails(name: string) {
    selectedSchema = name;
    schemaLoading = true;
    try {
      const [tables, cols, seqs] = await Promise.all([
        api.listTables(name),
        (async () => { try { const d = await api.getTableDetails(name, ''); return d.columns; } catch { return []; } })(),
        api.listSequences(name),
      ]);
      schemaTables = tables;
      schemaColumns = cols;
      schemaSequences = seqs;
    } catch {
      schemaTables = [];
      schemaColumns = [];
      schemaSequences = [];
    } finally {
      schemaLoading = false;
    }
  }

  async function handleCreateSchema() {
    loadingStates['createSchema'] = true;
    try {
      await api.createSchema(createSchemaName);
      createSchemaModal = false;
      createSchemaName = '';
      schemas = await api.listSchemas();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to create schema';
    } finally {
      loadingStates['createSchema'] = false;
    }
  }

  async function handleDropSchema() {
    loadingStates['dropSchema'] = true;
    try {
      await api.dropSchema(dropSchemaName);
      dropSchemaConfirm = false;
      dropSchemaName = '';
      selectedSchema = null;
      schemas = await api.listSchemas();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to drop schema';
    } finally {
      loadingStates['dropSchema'] = false;
    }
  }

  async function handleCreateTable() {
    if (!selectedSchema) return;
    loadingStates['createTable'] = true;
    try {
      const cols = createTableColumns.map(c => ({
        name: c.name,
        type: c.type,
        nullable: c.nullable,
        is_pk: c.is_pk,
        default: c.default || undefined,
      }));
      await api.createTable(selectedSchema, createTableName, cols);
      createTableModal = false;
      createTableName = '';
      createTableColumns = [];
      loadSchemaDetails(selectedSchema);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to create table';
    } finally {
      loadingStates['createTable'] = false;
    }
  }

  async function handleAddColumn() {
    if (!selectedSchema) return;
    loadingStates['addColumn'] = true;
    const table = schemaTables[0]?.table_name;
    if (!table) return;
    try {
      await api.addColumn(selectedSchema, table, { name: addColName, type: addColType, nullable: addColNullable, default: addColDefault || undefined });
      addColumnModal = false;
      addColName = '';
      addColType = 'text';
      addColNullable = true;
      addColDefault = '';
      if (selectedSchema) loadSchemaDetails(selectedSchema);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to add column';
    } finally {
      loadingStates['addColumn'] = false;
    }
  }

  async function handleDropColumn() {
    if (!selectedSchema) return;
    loadingStates['dropColumn'] = true;
    const table = schemaTables[0]?.table_name;
    if (!table) return;
    try {
      await api.dropColumn(selectedSchema, table, dropColName);
      dropColConfirm = false;
      dropColName = '';
      if (selectedSchema) loadSchemaDetails(selectedSchema);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to drop column';
    } finally {
      loadingStates['dropColumn'] = false;
    }
  }

  async function handleAlterColumn() {
    if (!selectedSchema) return;
    loadingStates['alterColumn'] = true;
    const table = schemaTables[0]?.table_name;
    if (!table) return;
    try {
      const changes: Record<string, unknown> = {};
      if (alterColNewName) changes.new_name = alterColNewName;
      if (alterColNewType) changes.new_type = alterColNewType;
      await api.alterColumn(selectedSchema, table, alterColOrigName, changes);
      alterColModal = false;
      alterColOrigName = '';
      alterColNewName = '';
      alterColNewType = '';
      if (selectedSchema) loadSchemaDetails(selectedSchema);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to alter column';
    } finally {
      loadingStates['alterColumn'] = false;
    }
  }

  async function handleAddConstraint() {
    if (!selectedSchema) return;
    loadingStates['addConstraint'] = true;
    const table = schemaTables[0]?.table_name;
    if (!table) return;
    try {
      await api.addConstraint(selectedSchema, table, {
        name: constraintName,
        type: constraintType,
        columns: constraintCols.split(',').map(c => c.trim()),
      });
      addConstraintModal = false;
      constraintName = '';
      constraintCols = '';
      if (selectedSchema) loadSchemaDetails(selectedSchema);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to add constraint';
    } finally {
      loadingStates['addConstraint'] = false;
    }
  }

  async function handleDropConstraint() {
    if (!selectedSchema) return;
    loadingStates['dropConstraint'] = true;
    const table = schemaTables[0]?.table_name;
    if (!table) return;
    try {
      await api.dropConstraint(selectedSchema, table, dropConstraintName);
      dropConstraintConfirm = false;
      dropConstraintName = '';
      if (selectedSchema) loadSchemaDetails(selectedSchema);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to drop constraint';
    } finally {
      loadingStates['dropConstraint'] = false;
    }
  }

  async function handleCreateIndex() {
    if (!selectedSchema) return;
    loadingStates['createIndex'] = true;
    const table = schemaTables[0]?.table_name;
    if (!table) return;
    try {
      await api.createIndex(selectedSchema, table, {
        name: indexName,
        columns: indexCols.split(',').map(c => c.trim()),
        unique: indexUnique,
      });
      createIndexModal = false;
      indexName = '';
      indexCols = '';
      indexUnique = false;
      if (selectedSchema) loadSchemaDetails(selectedSchema);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to create index';
    } finally {
      loadingStates['createIndex'] = false;
    }
  }

  async function handleDropIndex() {
    if (!selectedSchema) return;
    loadingStates['dropIndex'] = true;
    const table = schemaTables[0]?.table_name;
    if (!table) return;
    try {
      await api.dropIndex(selectedSchema, table, dropIndexName);
      dropIndexConfirm = false;
      dropIndexName = '';
      if (selectedSchema) loadSchemaDetails(selectedSchema);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to drop index';
    } finally {
      loadingStates['dropIndex'] = false;
    }
  }

  async function handleCreateSequence() {
    if (!selectedSchema) return;
    loadingStates['createSequence'] = true;
    try {
      await api.createSequence(selectedSchema, { name: seqName, start: parseInt(seqStart), increment: parseInt(seqIncrement) });
      createSeqModal = false;
      seqName = '';
      seqStart = '1';
      seqIncrement = '1';
      if (selectedSchema) loadSchemaDetails(selectedSchema);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to create sequence';
    } finally {
      loadingStates['createSequence'] = false;
    }
  }

  async function handleDropSequence() {
    if (!selectedSchema) return;
    loadingStates['dropSequence'] = true;
    try {
      await api.dropSequence(selectedSchema, dropSeqName);
      dropSeqConfirm = false;
      dropSeqName = '';
      if (selectedSchema) loadSchemaDetails(selectedSchema);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to drop sequence';
    } finally {
      loadingStates['dropSequence'] = false;
    }
  }

  async function handleGetDDL() {
    if (!selectedSchema) return;
    ddlLoading = true;
    ddlModal = true;
    try {
      ddlContent = await api.getTableDDL(selectedSchema, schemaTables[0]?.table_name || '');
    } catch (e) {
      ddlContent = 'Failed to load DDL: ' + (e instanceof Error ? e.message : 'Unknown error');
    } finally {
      ddlLoading = false;
    }
  }
</script>

<div class="max-w-6xl mx-auto">
  <div class="flex items-center justify-between mb-6">
    <h1 class="text-2xl font-bold" style="color: var(--text-primary);">Schema Designer</h1>
    <div class="flex items-center gap-2">
      <button onclick={() => { createSchemaName = ''; createSchemaModal = true; }} class="btn btn-primary btn-sm">+ Schema</button>
      {#if selectedSchema}
        <button onclick={handleGetDDL} class="btn btn-secondary btn-sm">Get DDL</button>
      {/if}
    </div>
  </div>

  <Alert message={error} type="error" />

  <div class="flex flex-col lg:flex-row gap-6">
    <div class="w-full lg:w-64 flex-shrink-0">
      <div class="card p-4">
        <h3 class="text-xs font-semibold uppercase tracking-wider mb-3" style="color: var(--text-tertiary);">Schemas</h3>
        {#if loading}
          {#each [1,2,3] as _}
            <div class="skeleton h-10 w-full mb-2 rounded-lg"></div>
          {/each}
        {:else if schemas.length === 0}
          <p class="text-sm" style="color: var(--text-secondary);">No schemas</p>
        {:else}
          <div class="space-y-1">
            {#each schemas as s}
              <div class="flex items-center justify-between p-2 rounded-lg text-sm" style="background-color: {selectedSchema === s.schema_name ? 'var(--accent-muted)' : 'transparent'};">
                <button onclick={() => loadSchemaDetails(s.schema_name)} class="flex-1 text-left" style="color: {selectedSchema === s.schema_name ? 'var(--accent)' : 'var(--text-secondary)'};">
                  {s.schema_name}
                </button>
                <button
                  onclick={() => { dropSchemaName = s.schema_name; dropSchemaConfirm = true; }}
                  class="text-xs p-1" style="color: var(--text-tertiary);"
                >✕</button>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </div>

    <div class="flex-1 min-w-0">
      {#if !selectedSchema}
        <div class="card p-12 text-center">
          <div class="text-4xl mb-3">▤</div>
          <h3 class="text-base font-semibold" style="color: var(--text-secondary);">Select a Schema</h3>
          <p class="text-sm mt-1" style="color: var(--text-tertiary);">Choose a schema to manage tables, columns, indexes and sequences</p>
        </div>
      {:else if schemaLoading}
        <LoadingCard />
      {:else}
        <div class="card p-5 mb-4">
          <div class="flex items-center justify-between mb-4">
            <h2 class="text-lg font-semibold" style="color: var(--text-primary);">{selectedSchema} <span class="text-sm font-normal" style="color: var(--text-tertiary);">({schemaTables.length} tables)</span></h2>
            <div class="flex gap-2">
              <button onclick={() => { createTableName = ''; createTableColumns = []; createTableModal = true; }} class="btn btn-primary btn-sm">+ Table</button>
              <button onclick={() => { seqName = ''; createSeqModal = true; }} class="btn btn-secondary btn-sm">+ Sequence</button>
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
            <div class="card p-4">
              <h4 class="text-xs font-semibold mb-2" style="color: var(--text-secondary);">Columns</h4>
              <div class="flex gap-2">
                <button onclick={() => { addColName = ''; addColumnModal = true; }} class="btn btn-primary btn-sm">+ Add</button>
                <button onclick={() => { dropColName = ''; dropColConfirm = true; }} class="btn btn-danger btn-sm">Drop</button>
                <button onclick={() => { alterColOrigName = ''; alterColNewName = ''; alterColNewType = ''; alterColModal = true; }} class="btn btn-secondary btn-sm">Alter</button>
              </div>
            </div>
            <div class="card p-4">
              <h4 class="text-xs font-semibold mb-2" style="color: var(--text-secondary);">Constraints</h4>
              <div class="flex gap-2">
                <button onclick={() => { constraintName = ''; constraintCols = ''; addConstraintModal = true; }} class="btn btn-primary btn-sm">+ Add</button>
                <button onclick={() => { dropConstraintName = ''; dropConstraintConfirm = true; }} class="btn btn-danger btn-sm">Drop</button>
              </div>
            </div>
            <div class="card p-4">
              <h4 class="text-xs font-semibold mb-2" style="color: var(--text-secondary);">Indexes</h4>
              <div class="flex gap-2">
                <button onclick={() => { indexName = ''; indexCols = ''; indexUnique = false; createIndexModal = true; }} class="btn btn-primary btn-sm">+ Create</button>
                <button onclick={() => { dropIndexName = ''; dropIndexConfirm = true; }} class="btn btn-danger btn-sm">Drop</button>
              </div>
            </div>
            <div class="card p-4">
              <h4 class="text-xs font-semibold mb-2" style="color: var(--text-secondary);">Sequences</h4>
              <div class="flex gap-2">
                <button onclick={() => { seqName = ''; createSeqModal = true; }} class="btn btn-primary btn-sm">+ Create</button>
                <button onclick={() => { dropSeqName = ''; dropSeqConfirm = true; }} class="btn btn-danger btn-sm">Drop</button>
              </div>
            </div>
          </div>
        </div>
      {/if}
    </div>
  </div>
</div>

<!-- Modals -->
<Modal title="Create Schema" open={createSchemaModal} onclose={() => createSchemaModal = false}>
  <div class="mb-4">
    <label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">Schema Name</label>
    <input type="text" bind:value={createSchemaName} class="input" placeholder="new_schema" />
  </div>
  <div class="flex justify-end gap-3">
    <button onclick={() => createSchemaModal = false} class="btn btn-secondary">Cancel</button>
    <button onclick={handleCreateSchema} disabled={loadingStates['createSchema'] || !createSchemaName.trim()} class="btn btn-primary">{loadingStates['createSchema'] ? '...' : 'Create'}</button>
  </div>
</Modal>

<ConfirmDialog
  open={dropSchemaConfirm}
  title="Drop Schema"
  description={'Are you sure you want to drop "' + dropSchemaName + '"? This cannot be undone.'}
  onconfirm={handleDropSchema}
  oncancel={() => dropSchemaConfirm = false}
  loading={loadingStates['dropSchema']}
/>

<Modal title="Create Table" open={createTableModal} onclose={() => createTableModal = false}>
  <div class="mb-4">
    <label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">Table Name</label>
    <input type="text" bind:value={createTableName} class="input" placeholder="new_table" />
  </div>
  <div class="mb-4">
    <div class="flex items-center justify-between mb-2">
      <span class="text-xs font-medium" style="color: var(--text-secondary);">Columns</span>
      <button onclick={() => { createTableColumns = [...createTableColumns, { name: '', type: 'text', nullable: true, is_pk: false, default: '' }]; }} class="btn btn-ghost btn-sm">+ Add Column</button>
    </div>
    {#each createTableColumns as col, i}
      <div class="flex gap-2 mb-2 items-center">
        <input type="text" bind:value={col.name} class="input flex-1" placeholder="name" />
        <input type="text" bind:value={col.type} class="input w-24" placeholder="type" />
        <label class="text-xs"><input type="checkbox" bind:checked={col.is_pk} /> PK</label>
        <button onclick={() => { createTableColumns = createTableColumns.filter((_, idx) => idx !== i); }} class="text-xs" style="color: var(--danger);">✕</button>
      </div>
    {/each}
  </div>
  <div class="flex justify-end gap-3">
    <button onclick={() => createTableModal = false} class="btn btn-secondary">Cancel</button>
    <button onclick={handleCreateTable} disabled={loadingStates['createTable'] || !createTableName.trim()} class="btn btn-primary">{loadingStates['createTable'] ? '...' : 'Create'}</button>
  </div>
</Modal>

<Modal title="Add Column" open={addColumnModal} onclose={() => addColumnModal = false}>
  <div class="mb-3"><label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">Name</label><input type="text" bind:value={addColName} class="input" /></div>
  <div class="mb-3"><label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">Type</label><input type="text" bind:value={addColType} class="input" placeholder="text" /></div>
  <div class="mb-3"><label class="flex items-center gap-2 text-xs"><input type="checkbox" bind:checked={addColNullable} /> Nullable</label></div>
  <div class="mb-3"><label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">Default</label><input type="text" bind:value={addColDefault} class="input" placeholder="optional" /></div>
  <div class="flex justify-end gap-3">
    <button onclick={() => addColumnModal = false} class="btn btn-secondary">Cancel</button>
    <button onclick={handleAddColumn} disabled={loadingStates['addColumn']} class="btn btn-primary">{loadingStates['addColumn'] ? '...' : 'Add'}</button>
  </div>
</Modal>

<ConfirmDialog
  open={dropColConfirm}
  title="Drop Column"
  description={'Enter column name to drop:'}
  onconfirm={handleDropColumn}
  oncancel={() => dropColConfirm = false}
  loading={loadingStates['dropColumn']}
>
  <div class="mb-4"><input type="text" bind:value={dropColName} class="input" placeholder="column_name" /></div>
</ConfirmDialog>

<Modal title="Alter Column" open={alterColModal} onclose={() => alterColModal = false}>
  <div class="mb-3"><label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">Column Name</label><input type="text" bind:value={alterColOrigName} class="input" placeholder="current_name" /></div>
  <div class="mb-3"><label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">New Name (optional)</label><input type="text" bind:value={alterColNewName} class="input" placeholder="new_name" /></div>
  <div class="mb-3"><label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">New Type (optional)</label><input type="text" bind:value={alterColNewType} class="input" placeholder="new_type" /></div>
  <div class="flex justify-end gap-3">
    <button onclick={() => alterColModal = false} class="btn btn-secondary">Cancel</button>
    <button onclick={handleAlterColumn} disabled={loadingStates['alterColumn']} class="btn btn-primary">{loadingStates['alterColumn'] ? '...' : 'Alter'}</button>
  </div>
</Modal>

<Modal title="Add Constraint" open={addConstraintModal} onclose={() => addConstraintModal = false}>
  <div class="mb-3"><label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">Constraint Name</label><input type="text" bind:value={constraintName} class="input" /></div>
  <div class="mb-3"><label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">Type</label>
    <select bind:value={constraintType} class="input">
      <option>PRIMARY KEY</option><option>FOREIGN KEY</option><option>UNIQUE</option><option>CHECK</option>
    </select>
  </div>
  <div class="mb-3"><label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">Columns (comma separated)</label><input type="text" bind:value={constraintCols} class="input" placeholder="col1, col2" /></div>
  <div class="flex justify-end gap-3">
    <button onclick={() => addConstraintModal = false} class="btn btn-secondary">Cancel</button>
    <button onclick={handleAddConstraint} disabled={loadingStates['addConstraint']} class="btn btn-primary">{loadingStates['addConstraint'] ? '...' : 'Add'}</button>
  </div>
</Modal>

<ConfirmDialog
  open={dropConstraintConfirm}
  title="Drop Constraint"
  onconfirm={handleDropConstraint}
  oncancel={() => dropConstraintConfirm = false}
  loading={loadingStates['dropConstraint']}
>
  <div class="mb-4"><input type="text" bind:value={dropConstraintName} class="input" placeholder="constraint_name" /></div>
</ConfirmDialog>

<Modal title="Create Index" open={createIndexModal} onclose={() => createIndexModal = false}>
  <div class="mb-3"><label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">Index Name</label><input type="text" bind:value={indexName} class="input" /></div>
  <div class="mb-3"><label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">Columns (comma separated)</label><input type="text" bind:value={indexCols} class="input" placeholder="col1, col2" /></div>
  <div class="mb-3"><label class="flex items-center gap-2 text-xs"><input type="checkbox" bind:checked={indexUnique} /> Unique</label></div>
  <div class="flex justify-end gap-3">
    <button onclick={() => createIndexModal = false} class="btn btn-secondary">Cancel</button>
    <button onclick={handleCreateIndex} disabled={loadingStates['createIndex']} class="btn btn-primary">{loadingStates['createIndex'] ? '...' : 'Create'}</button>
  </div>
</Modal>

<ConfirmDialog
  open={dropIndexConfirm}
  title="Drop Index"
  onconfirm={handleDropIndex}
  oncancel={() => dropIndexConfirm = false}
  loading={loadingStates['dropIndex']}
>
  <div class="mb-4"><input type="text" bind:value={dropIndexName} class="input" placeholder="index_name" /></div>
</ConfirmDialog>

<Modal title="Create Sequence" open={createSeqModal} onclose={() => createSeqModal = false}>
  <div class="mb-3"><label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">Sequence Name</label><input type="text" bind:value={seqName} class="input" /></div>
  <div class="mb-3"><label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">Start Value</label><input type="number" bind:value={seqStart} class="input" /></div>
  <div class="mb-3"><label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">Increment</label><input type="number" bind:value={seqIncrement} class="input" /></div>
  <div class="flex justify-end gap-3">
    <button onclick={() => createSeqModal = false} class="btn btn-secondary">Cancel</button>
    <button onclick={handleCreateSequence} disabled={loadingStates['createSequence']} class="btn btn-primary">{loadingStates['createSequence'] ? '...' : 'Create'}</button>
  </div>
</Modal>

<ConfirmDialog
  open={dropSeqConfirm}
  title="Drop Sequence"
  onconfirm={handleDropSequence}
  oncancel={() => dropSeqConfirm = false}
  loading={loadingStates['dropSequence']}
>
  <div class="mb-4"><input type="text" bind:value={dropSeqName} class="input" placeholder="sequence_name" /></div>
</ConfirmDialog>

<Modal title="DDL" open={ddlModal} onclose={() => ddlModal = false}>
  {#if ddlLoading}
    <p>Loading...</p>
  {:else}
    <pre class="text-xs font-mono whitespace-pre-wrap rounded-lg p-4 overflow-x-auto" style="background-color: var(--bg-tertiary); color: var(--text-secondary); max-height: 400px;">{ddlContent}</pre>
  {/if}
  <div class="flex justify-end mt-4">
    <button onclick={() => ddlModal = false} class="btn btn-secondary">Close</button>
  </div>
</Modal>
