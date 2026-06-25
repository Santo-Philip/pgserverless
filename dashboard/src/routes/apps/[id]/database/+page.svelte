<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';
	import Card from '$lib/components/Card.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import Alert from '$lib/components/Alert.svelte';

	type Col = { name: string; type: string; nullable: boolean; is_pk: boolean; default_value: string };
	type Table = { name: string; columns: Col[] };

	let tables = $state<Table[]>([]);
	let loading = $state(true);
	let selectedTable = $state<string | null>(null);
	let tableData = $state<Record<string, unknown>[]>([]);
	let columns = $state<string[]>([]);
	let loadingData = $state(false);
	let message = $state('');
	let messageType = $state<'success' | 'error' | undefined>(undefined);

	let showCreateTable = $state(false);
	let newTableName = $state('');
	let newTableColumns = $state<{name: string; type: string; nullable: boolean; is_pk: boolean; default_value: string}[]>([
		{ name: 'id', type: 'uuid', nullable: false, is_pk: true, default_value: 'gen_random_uuid()' }
	]);
	let creating = $state(false);

	let showAddColumn = $state(false);
	let addColName = $state('');
	let addColType = $state('text');
	let addColNullable = $state(true);
	let addColDefault = $state('');
	let addingCol = $state(false);

	let showInsertRow = $state(false);
	let insertValues = $state<Record<string, string>>({});
	let inserting = $state(false);

	let showEditRow = $state(false);
	let editRowWhere = $state<Record<string, unknown>>({});
	let editValues = $state<Record<string, string>>({});
	let editing = $state(false);

	const PG_TYPES = [
		'uuid', 'text', 'varchar(255)', 'integer', 'bigint', 'smallint',
		'numeric', 'real', 'double precision', 'boolean', 'json', 'jsonb',
		'timestamp', 'timestamptz', 'date', 'time', 'bytea', 'serial', 'bigserial'
	];

	onMount(async () => {
		try {
			tables = await api.listTables($page.params.id!);
		} catch {}
		loading = false;
	});

	function addNewTableColumn() {
		newTableColumns = [...newTableColumns, { name: '', type: 'text', nullable: true, is_pk: false, default_value: '' }];
	}

	function removeNewTableColumn(index: number) {
		newTableColumns = newTableColumns.filter((_, i) => i !== index);
	}

	async function handleCreateTable() {
		creating = true;
		message = '';
		try {
			await api.createTable($page.params.id!, newTableName, newTableColumns);
			tables = await api.listTables($page.params.id!);
			showCreateTable = false;
			newTableName = '';
			newTableColumns = [{ name: 'id', type: 'uuid', nullable: false, is_pk: true, default_value: 'gen_random_uuid()' }];
			message = `Table "${newTableName}" created successfully.`;
			messageType = 'success';
		} catch (e) {
			message = e instanceof Error ? e.message : 'Failed to create table';
			messageType = 'error';
		}
		creating = false;
	}

	async function selectTable(name: string) {
		selectedTable = name;
		loadingData = true;
		try {
			const data = await api.getTableData($page.params.id!, name);
			tableData = data;
			columns = data.length > 0 ? Object.keys(data[0]) : [];
		} catch {
			tableData = [];
			columns = [];
		}
		loadingData = false;
	}

	function openAddColumn() {
		addColName = '';
		addColType = 'text';
		addColNullable = true;
		addColDefault = '';
		showAddColumn = true;
	}

	async function handleAddColumn() {
		addingCol = true;
		message = '';
		try {
			await api.addColumn($page.params.id!, selectedTable!, {
				name: addColName,
				type: addColType,
				nullable: addColNullable,
				default_value: addColDefault || undefined
			});
			tables = await api.listTables($page.params.id!);
			await selectTable(selectedTable!);
			showAddColumn = false;
			message = `Column "${addColName}" added successfully.`;
			messageType = 'success';
		} catch (e) {
			message = e instanceof Error ? e.message : 'Failed to add column';
			messageType = 'error';
		}
		addingCol = false;
	}

	function openInsertRow() {
		const table = tables.find(t => t.name === selectedTable);
		if (!table) return;
		const vals: Record<string, string> = {};
		for (const col of table.columns) {
			if (col.is_pk && col.default_value) {
				vals[col.name] = '';
			} else {
				vals[col.name] = '';
			}
		}
		insertValues = vals;
		showInsertRow = true;
	}

	async function handleInsertRow() {
		inserting = true;
		message = '';
		try {
			const values: Record<string, unknown> = {};
			for (const [k, v] of Object.entries(insertValues)) {
				if (v === '' || v === null) continue;
				values[k] = v;
			}
			await api.insertRow($page.params.id!, selectedTable!, values);
			await selectTable(selectedTable!);
			showInsertRow = false;
			message = 'Row inserted successfully.';
			messageType = 'success';
		} catch (e) {
			message = e instanceof Error ? e.message : 'Failed to insert row';
			messageType = 'error';
		}
		inserting = false;
	}

	function openEditRow(row: Record<string, unknown>) {
		editRowWhere = {};
		const vals: Record<string, string> = {};
		const table = tables.find(t => t.name === selectedTable);
		const pkCol = table?.columns.find(c => c.is_pk);
		if (pkCol && row[pkCol.name] !== undefined) {
			editRowWhere[pkCol.name] = row[pkCol.name];
		}
		for (const col of columns) {
			vals[col] = row[col] !== null && row[col] !== undefined ? String(row[col]) : '';
		}
		editValues = vals;
		showEditRow = true;
	}

	async function handleEditRow() {
		editing = true;
		message = '';
		try {
			const values: Record<string, unknown> = {};
			for (const [k, v] of Object.entries(editValues)) {
				if (k in editRowWhere) continue;
				values[k] = v;
			}
			if (Object.keys(values).length === 0) {
				message = 'No values to update';
				messageType = 'error';
				editing = false;
				return;
			}
			await api.updateRow($page.params.id!, selectedTable!, values, editRowWhere);
			await selectTable(selectedTable!);
			showEditRow = false;
			message = 'Row updated successfully.';
			messageType = 'success';
		} catch (e) {
			message = e instanceof Error ? e.message : 'Failed to update row';
			messageType = 'error';
		}
		editing = false;
	}

	async function handleDeleteRow(row: Record<string, unknown>) {
		const table = tables.find(t => t.name === selectedTable);
		const pkCol = table?.columns.find(c => c.is_pk);
		if (!pkCol || row[pkCol.name] === undefined) {
			message = 'Cannot delete: no primary key found';
			messageType = 'error';
			return;
		}
		if (!confirm(`Delete this row (${pkCol.name} = ${row[pkCol.name]})?`)) return;
		message = '';
		try {
			const where: Record<string, unknown> = {};
			where[pkCol.name] = row[pkCol.name];
			await api.deleteRow($page.params.id!, selectedTable!, where);
			await selectTable(selectedTable!);
			message = 'Row deleted successfully.';
			messageType = 'success';
		} catch (e) {
			message = e instanceof Error ? e.message : 'Failed to delete row';
			messageType = 'error';
		}
	}
</script>

<div class="flex gap-6">
	<div class="w-56 flex-shrink-0">
		<Card title="Tables">
			<div class="flex justify-end mb-3 -mt-1">
				<button onclick={() => { showCreateTable = true; }} class="btn btn-primary btn-sm">+ New</button>
			</div>
			{#if loading}
				<div class="space-y-2"><div class="skeleton h-8 w-full"></div><div class="skeleton h-8 w-full"></div></div>
			{:else if tables.length === 0}
				<p class="text-xs" style="color: var(--text-tertiary)">No tables yet</p>
			{:else}
				<div class="space-y-0.5">
					{#each tables as t}
						<button
							onclick={() => selectTable(t.name)}
							class="w-full text-left px-3 py-2 rounded-lg text-sm transition-colors"
							class:active={selectedTable === t.name}
							style={selectedTable === t.name ? 'background-color: var(--accent-muted); color: var(--accent)' : 'color: var(--text-secondary)'}
						>
							<span>{t.name}</span>
							<span class="text-xs ml-1" style="color: var(--text-tertiary)">({t.columns.length})</span>
						</button>
					{/each}
				</div>
			{/if}
		</Card>

		<div class="mt-4 space-y-2">
			<a href="/apps/{$page.params.id}/sql" class="btn btn-secondary w-full text-xs">Open SQL Editor</a>
		</div>
	</div>

	<div class="flex-1 min-w-0">
		<Alert {message} type={messageType} />

		{#if !selectedTable}
			<EmptyState icon="▤" title="Select a table" description="Choose a table from the sidebar or create a new one." />
		{:else if loadingData}
			<Card title={selectedTable}>
				<Skeleton rows={5} />
			</Card>
		{:else}
			<Card title={selectedTable}>
				<div class="flex items-center justify-between mb-4 -mt-1">
					<div class="flex gap-2">
						<button onclick={openInsertRow} class="btn btn-primary btn-sm">+ Insert Row</button>
						<button onclick={openAddColumn} class="btn btn-secondary btn-sm">+ Add Column</button>
					</div>
					{#if tableData.length > 0}
						<span class="text-xs" style="color: var(--text-tertiary)">{tableData.length} rows</span>
					{/if}
				</div>

				{#if tableData.length === 0}
					<p class="text-sm" style="color: var(--text-secondary)">No data in this table. Insert a row to get started.</p>
				{:else}
					<div class="overflow-x-auto -mx-5 -mb-5">
						<div class="table-wrap">
							<table class="w-full">
								<thead>
									<tr>
										{#each columns as col}
											<th>{col}</th>
										{/each}
										<th class="text-right">Actions</th>
									</tr>
								</thead>
								<tbody>
									{#each tableData as row}
										<tr>
											{#each columns as col}
												<td class="font-mono text-xs max-w-[200px] truncate">{row[col] !== null && row[col] !== undefined ? JSON.stringify(row[col]) : 'NULL'}</td>
											{/each}
											<td class="text-right">
												<div class="flex items-center justify-end gap-1">
													<button onclick={() => openEditRow(row)} class="btn btn-ghost btn-sm text-xs">Edit</button>
													<button onclick={() => handleDeleteRow(row)} class="btn btn-ghost btn-sm text-xs" style="color: var(--danger)">Delete</button>
												</div>
											</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					</div>
				{/if}
			</Card>
		{/if}
	</div>
</div>

<Modal title="Create Table" open={showCreateTable} onclose={() => showCreateTable = false}>
	<div class="space-y-4">
		<div>
			<label class="block text-sm font-medium mb-1.5" style="color: var(--text-secondary)">Table Name</label>
			<input type="text" bind:value={newTableName} class="input font-mono" placeholder="users" />
		</div>

		<div>
			<div class="flex items-center justify-between mb-2">
				<label class="text-sm font-medium" style="color: var(--text-secondary)">Columns</label>
				<button onclick={addNewTableColumn} class="btn btn-ghost btn-sm">+ Add Column</button>
			</div>
			<div class="space-y-2 max-h-64 overflow-y-auto">
				{#each newTableColumns as col, i}
					<div class="flex items-center gap-2 p-2 rounded-lg" style="background-color: var(--bg-tertiary);">
						<input type="text" bind:value={col.name} class="input flex-1 font-mono text-xs" placeholder="column_name" />
						<select bind:value={col.type} class="input flex-1 text-xs">
							{#each PG_TYPES as t}
								<option value={t}>{t}</option>
							{/each}
						</select>
						<label class="flex items-center gap-1 text-xs" style="color: var(--text-tertiary)">
							<input type="checkbox" bind:checked={col.is_pk} class="rounded" /> PK
						</label>
						<label class="flex items-center gap-1 text-xs" style="color: var(--text-tertiary)">
							<input type="checkbox" bind:checked={col.nullable} class="rounded" /> Null
						</label>
						<input type="text" bind:value={col.default_value} class="input w-28 font-mono text-xs" placeholder="default" />
						{#if i > 0}
							<button onclick={() => removeNewTableColumn(i)} class="btn btn-ghost btn-sm text-xs" style="color: var(--danger)">&times;</button>
						{/if}
					</div>
				{/each}
			</div>
		</div>

		<div class="flex justify-end gap-3 pt-2">
			<button type="button" onclick={() => showCreateTable = false} class="btn btn-secondary">Cancel</button>
			<button type="button" onclick={handleCreateTable} disabled={creating || !newTableName || newTableColumns.length === 0} class="btn btn-primary">
				{creating ? 'Creating...' : 'Create Table'}
			</button>
		</div>
	</div>
</Modal>

<Modal title="Add Column" open={showAddColumn} onclose={() => showAddColumn = false}>
	<div class="space-y-4">
		<div>
			<label class="block text-sm font-medium mb-1.5" style="color: var(--text-secondary)">Column Name</label>
			<input type="text" bind:value={addColName} class="input font-mono" placeholder="column_name" />
		</div>
		<div>
			<label class="block text-sm font-medium mb-1.5" style="color: var(--text-secondary)">Type</label>
			<select bind:value={addColType} class="input">
				{#each PG_TYPES as t}
					<option value={t}>{t}</option>
				{/each}
			</select>
		</div>
		<div class="flex items-center gap-4">
			<label class="flex items-center gap-2 text-sm" style="color: var(--text-secondary)">
				<input type="checkbox" bind:checked={addColNullable} class="rounded" /> Allow NULL
			</label>
		</div>
		<div>
			<label class="block text-sm font-medium mb-1.5" style="color: var(--text-secondary)">Default Value (optional)</label>
			<input type="text" bind:value={addColDefault} class="input font-mono" placeholder="e.g. now(), 0, ''" />
		</div>
		<div class="flex justify-end gap-3 pt-2">
			<button type="button" onclick={() => showAddColumn = false} class="btn btn-secondary">Cancel</button>
			<button type="button" onclick={handleAddColumn} disabled={addingCol || !addColName} class="btn btn-primary">
				{addingCol ? 'Adding...' : 'Add Column'}
			</button>
		</div>
	</div>
</Modal>

<Modal title="Insert Row" open={showInsertRow} onclose={() => showInsertRow = false}>
	<div class="space-y-4">
		{#each tables.find(t => t.name === selectedTable)?.columns || [] as col}
			{#if !col.is_pk || !col.default_value}
				<div>
					<label class="block text-sm font-medium mb-1.5" style="color: var(--text-secondary)">
						{col.name}
						<span class="text-xs ml-1" style="color: var(--text-tertiary)">{col.type}{col.is_pk ? ' (PK)' : ''}</span>
					</label>
					<input type="text" bind:value={insertValues[col.name]} class="input font-mono" placeholder={col.default_value || 'NULL'} />
				</div>
			{/if}
		{/each}
		<div class="flex justify-end gap-3 pt-2">
			<button type="button" onclick={() => showInsertRow = false} class="btn btn-secondary">Cancel</button>
			<button type="button" onclick={handleInsertRow} disabled={inserting} class="btn btn-primary">
				{inserting ? 'Inserting...' : 'Insert Row'}
			</button>
		</div>
	</div>
</Modal>

<Modal title="Edit Row" open={showEditRow} onclose={() => showEditRow = false}>
	<div class="space-y-4">
		{#each columns as col}
			<div>
				<label class="block text-sm font-medium mb-1.5" style="color: var(--text-secondary)">
					{col}
					{#if editRowWhere[col] !== undefined}
						<span class="text-xs ml-1" style="color: var(--warning)">(WHERE)</span>
					{/if}
				</label>
				<input
					type="text"
					bind:value={editValues[col]}
					disabled={editRowWhere[col] !== undefined}
					class="input font-mono"
				/>
			</div>
		{/each}
		<div class="flex justify-end gap-3 pt-2">
			<button type="button" onclick={() => showEditRow = false} class="btn btn-secondary">Cancel</button>
			<button type="button" onclick={handleEditRow} disabled={editing} class="btn btn-primary">
				{editing ? 'Saving...' : 'Save Changes'}
			</button>
		</div>
	</div>
</Modal>
