<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import Card from '$lib/components/Card.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import Breadcrumbs from '$lib/components/Breadcrumbs.svelte';

	let apps = $state<{ id: string; name: string; tables: string[] }[]>([]);
	let loading = $state(true);
	let selectedApp = $state<string | null>(null);
	let selectedTable = $state<string | null>(null);
	let tableData = $state<Record<string, unknown>[]>([]);

	onMount(async () => {
		try {
			const result = await api.listApps();
			for (const app of result) {
				try {
					const t = await api.get<{ tables: string[] }>(`/api/v1/platform/apps/${app.id}/tables`);
					apps.push({ id: app.id, name: app.name, tables: t.tables || [] });
				} catch {}
			}
		} catch {}
		loading = false;
	});

	async function loadTable(appId: string, table: string) {
		selectedApp = appId;
		selectedTable = table;
		try {
			tableData = await api.get<Record<string, unknown>[]>(`/api/v1/platform/apps/${appId}/tables/${table}`);
		} catch { tableData = []; }
	}
</script>

<Breadcrumbs items={[{ label: 'Database' }]} />

<div class="max-w-7xl mx-auto">
	<div class="mb-8">
		<h1 class="text-2xl font-bold">Database</h1>
		<p class="text-sm mt-1" style="color: var(--text-secondary)">Browse tables and schemas across all applications</p>
	</div>

	{#if loading}
		<div class="grid grid-cols-1 md:grid-cols-4 gap-4"><Skeleton rows={6} /><Skeleton rows={6} /></div>
	{:else if apps.length === 0}
		<EmptyState icon="▤" title="No applications" description="Create an application to get started." />
	{:else}
		<div class="flex gap-6">
			<div class="w-64 flex-shrink-0 space-y-4">
				{#each apps as app}
					<Card title={app.name}>
						{#if app.tables.length === 0}
							<p class="text-xs" style="color: var(--text-tertiary)">No tables</p>
						{:else}
							<div class="space-y-0.5">
								{#each app.tables as table}
									<button
										onclick={() => loadTable(app.id, table)}
										class="w-full text-left px-3 py-2 rounded-lg text-sm transition-colors"
										class:active={selectedApp === app.id && selectedTable === table}
										style={selectedApp === app.id && selectedTable === table ? 'background-color: var(--accent-muted); color: var(--accent)' : 'color: var(--text-secondary)'}
									>
										{table}
									</button>
								{/each}
							</div>
						{/if}
					</Card>
				{/each}
			</div>

			<div class="flex-1 min-w-0">
				{#if !selectedTable}
					<EmptyState icon="▤" title="Select a table" description="Choose a table from the sidebar to browse its contents." />
				{:else}
					<Card title={`${selectedTable} (${tableData.length} rows)`}>
						{#if tableData.length === 0}
							<p class="text-sm" style="color: var(--text-secondary)">No data.</p>
						{:else}
							<div class="overflow-x-auto -mx-5 -mb-5">
								<div class="table-wrap">
									<table class="w-full">
										<thead><tr>{#each Object.keys(tableData[0]) as col}<th>{col}</th>{/each}</tr></thead>
										<tbody>
											{#each tableData as row}
												<tr>{#each Object.keys(tableData[0]) as col}<td class="font-mono text-xs max-w-[200px] truncate">{JSON.stringify(row[col])}</td>{/each}</tr>
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
	{/if}
</div>
