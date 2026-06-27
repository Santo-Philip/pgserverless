<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import type { Database, TableInfo, Extension } from '$lib/types';
	import Card from '$lib/components/Card.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import LoadingCard from '$lib/components/LoadingCard.svelte';
	import { page } from '$app/stores';

	let loading = $state(true);
	let db: Database | null = $state(null);
	let tables: TableInfo[] = $state([]);
	let extensions: Extension[] = $state([]);
	let activeTab = $state('tables');

	// Table modals
	let showCreateTable = $state(false);
	let newTableName = $state('');
	let newTableCols = $state([{ name: '', type: 'text', nullable: true, is_pk: false }]);
	let selectedTable = $state<string | null>(null);
	let tableData = $state<Record<string, unknown>[]>([]);
	let tableLoading = $state(false);

	// SQL
	let sqlQuery = $state('');
	let sqlResults = $state<Record<string, unknown>[]>([]);
	let sqlLoading = $state(false);

	// Extension
	let showExtensions = $state(false);

	async function load() {
		loading = true;
		try {
			const id = $page.params.id;
			db = await api.getDatabase(id);
			[tables, extensions] = await Promise.all([
				api.listTables(id),
				api.listExtensions(),
			]);
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	onMount(load);

	async function loadTableData(table: string) {
		selectedTable = table;
		tableLoading = true;
		try {
			tableData = await api.getTableData($page.params.id, table);
		} catch (e) {
			console.error(e);
			tableData = [];
		} finally {
			tableLoading = false;
		}
	}

	async function createTable() {
		try {
			const columns = newTableCols.map(c => ({
				name: c.name,
				type: c.type,
				nullable: c.nullable,
				is_pk: c.is_pk,
			}));
			await api.createTable($page.params.id, newTableName, columns);
			showCreateTable = false;
			newTableName = '';
			newTableCols = [{ name: '', type: 'text', nullable: true, is_pk: false }];
			tables = await api.listTables($page.params.id);
		} catch (e) {
			alert('Failed: ' + (e as Error).message);
		}
	}

	async function runSQL() {
		sqlLoading = true;
		try {
			sqlResults = await api.runSQL($page.params.id, sqlQuery);
		} catch (e) {
			alert('Query failed: ' + (e as Error).message);
			sqlResults = [];
		} finally {
			sqlLoading = false;
		}
	}

	function addCol() {
		newTableCols = [...newTableCols, { name: '', type: 'text', nullable: true, is_pk: false }];
	}

	function removeCol(idx: number) {
		newTableCols = newTableCols.filter((_, i) => i !== idx);
	}
</script>

{#if loading}
	<div class="max-w-5xl mx-auto"><LoadingCard /></div>
{:else if db}
	<div class="max-w-5xl mx-auto">
		<div class="mb-6">
			<h1 class="text-2xl font-bold" style="color: var(--text-primary);">{db.name}</h1>
			<p class="text-sm mt-1" style="color: var(--text-secondary);">
				Schema: {db.schema_name} | User: {db.db_user} | Status: {db.status} | Size: {(db.size_bytes / 1024 / 1024).toFixed(2)} MB
			</p>
		</div>

		<div class="flex gap-2 mb-6 border-b" style="border-color: var(--border-primary);">
			<button onclick={() => activeTab = 'tables'} class="px-4 py-2 text-sm font-medium border-b-2 transition-colors"
				style={activeTab === 'tables' ? 'border-color: var(--accent); color: var(--accent);' : 'border-color: transparent; color: var(--text-secondary);'}>Tables</button>
			<button onclick={() => activeTab = 'sql'} class="px-4 py-2 text-sm font-medium border-b-2 transition-colors"
				style={activeTab === 'sql' ? 'border-color: var(--accent); color: var(--accent);' : 'border-color: transparent; color: var(--text-secondary);'}>SQL</button>
			<button onclick={() => { showExtensions = true; }} class="px-4 py-2 text-sm font-medium border-b-2 transition-colors"
				style="border-color: transparent; color: var(--text-secondary);">Extensions</button>
		</div>

		{#if activeTab === 'tables'}
			<div>
				<div class="flex items-center justify-between mb-4">
					<h2 class="text-lg font-semibold" style="color: var(--text-primary);">Tables</h2>
					<button onclick={() => showCreateTable = true} class="btn btn-primary btn-sm">New Table</button>
				</div>

				<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
					<div class="lg:col-span-1 space-y-2">
						{#each tables as t}
							<button onclick={() => loadTableData(t.name)}
								class="w-full text-left p-3 rounded-lg text-sm transition-colors"
								style={selectedTable === t.name ? 'background-color: var(--accent); color: white;' : 'background-color: var(--bg-hover);'}
							>
								<div class="font-medium">{t.name}</div>
								<div class="text-xs mt-0.5 opacity-70">{t.columns.length} columns</div>
							</button>
						{/each}
					</div>

					<div class="lg:col-span-2">
						{#if tableLoading}
							<LoadingCard />
						{:else if selectedTable && tableData.length > 0}
							<div class="overflow-x-auto rounded-xl border" style="border-color: var(--border-primary);">
								<table class="w-full text-sm">
									<thead>
										<tr style="background-color: var(--bg-hover);">
											{#each Object.keys(tableData[0]) as col}
												<th class="px-4 py-2 text-left font-medium" style="color: var(--text-primary);">{col}</th>
											{/each}
										</tr>
									</thead>
									<tbody>
										{#each tableData as row}
											<tr class="border-t" style="border-color: var(--border-primary);">
												{#each Object.values(row) as val}
													<td class="px-4 py-2" style="color: var(--text-secondary);">{JSON.stringify(val)}</td>
												{/each}
											</tr>
										{/each}
									</tbody>
								</table>
							</div>
						{:else if selectedTable}
							<Card title={selectedTable}><p class="text-sm" style="color: var(--text-secondary);">No data in this table.</p></Card>
						{:else}
							<Card title="Select a Table"><p class="text-sm" style="color: var(--text-secondary);">Choose a table from the left to view data.</p></Card>
						{/if}
					</div>
				</div>
			</div>
		{:else if activeTab === 'sql'}
			<div>
				<div class="mb-4">
					<textarea bind:value={sqlQuery} rows="6" class="w-full px-3 py-2 rounded-lg border text-sm font-mono"
						style="background-color: var(--bg-primary); border-color: var(--border-primary); color: var(--text-primary);"
						placeholder="SELECT * FROM my_table LIMIT 10;"></textarea>
					<div class="flex justify-end mt-2">
						<button onclick={runSQL} class="btn btn-primary btn-sm" disabled={sqlLoading || !sqlQuery}>
							{sqlLoading ? 'Running...' : 'Run Query'}
						</button>
					</div>
				</div>

				{#if sqlResults.length > 0}
					<div class="overflow-x-auto rounded-xl border" style="border-color: var(--border-primary);">
						<table class="w-full text-sm">
							<thead>
								<tr style="background-color: var(--bg-hover);">
									{#each Object.keys(sqlResults[0]) as col}
										<th class="px-4 py-2 text-left font-medium" style="color: var(--text-primary);">{col}</th>
									{/each}
								</tr>
							</thead>
							<tbody>
								{#each sqlResults as row}
									<tr class="border-t" style="border-color: var(--border-primary);">
										{#each Object.values(row) as val}
											<td class="px-4 py-2" style="color: var(--text-secondary);">{JSON.stringify(val)}</td>
										{/each}
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</div>
		{/if}
	</div>

	<!-- Create Table Modal -->
	{#if showCreateTable}
		<Modal title="Create Table" onclose={() => showCreateTable = false}>
			<div class="space-y-4">
			<div>
				<label for="newTableName" class="block text-sm font-medium mb-1" style="color: var(--text-primary);">Table Name</label>
				<input id="newTableName" type="text" class="w-full px-3 py-2 rounded-lg border text-sm"
					style="background-color: var(--bg-primary); border-color: var(--border-primary); color: var(--text-primary);"
					bind:value={newTableName} placeholder="my_table" />
			</div>

			<div>
				<label class="block text-sm font-medium mb-1" style="color: var(--text-primary);">Columns</label>
					{#each newTableCols as col, i}
						<div class="flex gap-2 mb-2">
							<input type="text" bind:value={col.name} placeholder="name" class="flex-1 px-2 py-1 rounded border text-sm"
								style="background-color: var(--bg-primary); border-color: var(--border-primary); color: var(--text-primary);" />
							<select bind:value={col.type} class="px-2 py-1 rounded border text-sm"
								style="background-color: var(--bg-primary); border-color: var(--border-primary); color: var(--text-primary);">
								<option value="text">text</option>
								<option value="integer">integer</option>
								<option value="bigint">bigint</option>
								<option value="boolean">boolean</option>
								<option value="timestamp with time zone">timestamptz</option>
								<option value="jsonb">jsonb</option>
								<option value="uuid">uuid</option>
								<option value="numeric">numeric</option>
							</select>
							<label class="flex items-center gap-1 text-xs">
								<input type="checkbox" bind:checked={col.nullable} /> Null
							</label>
							<label class="flex items-center gap-1 text-xs">
								<input type="checkbox" bind:checked={col.is_pk} /> PK
							</label>
							{#if newTableCols.length > 1}
								<button onclick={() => removeCol(i)} class="text-xs text-red-500">X</button>
							{/if}
						</div>
					{/each}
					<button onclick={addCol} class="text-xs" style="color: var(--accent);">+ Add Column</button>
				</div>

				<div class="flex justify-end gap-2 pt-2">
					<button onclick={() => showCreateTable = false} class="btn btn-ghost btn-sm">Cancel</button>
					<button onclick={createTable} class="btn btn-primary btn-sm" disabled={!newTableName}>Create</button>
				</div>
			</div>
		</Modal>
	{/if}

	<!-- Extensions Modal -->
	{#if showExtensions}
		<Modal title="PostgreSQL Extensions" onclose={() => showExtensions = false}>
			<div class="space-y-2">
				{#each extensions as ext}
					<div class="flex items-center justify-between p-3 rounded-lg" style="background-color: var(--bg-hover);">
						<div>
							<div class="text-sm font-medium" style="color: var(--text-primary);">{ext.name}</div>
							<div class="text-xs" style="color: var(--text-secondary);">{ext.description}</div>
						</div>
						<button
							onclick={async () => {
								try {
									await api.toggleExtension(ext.name, !ext.installed);
									extensions = await api.listExtensions();
								} catch (e) { alert('Failed: ' + (e as Error).message); }
							}}
							class="btn btn-ghost btn-xs"
						>
							{ext.installed ? 'Deactivate' : 'Activate'}
						</button>
					</div>
				{/each}
			</div>
		</Modal>
	{/if}
{/if}