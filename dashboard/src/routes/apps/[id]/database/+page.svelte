<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';
	import Card from '$lib/components/Card.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';

	let tables = $state<{ name: string; columns?: { name: string; type: string }[] }[]>([]);
	let loading = $state(true);
	let selectedTable = $state<string | null>(null);
	let tableData = $state<Record<string, unknown>[]>([]);
	let loadingData = $state(false);

	onMount(async () => {
		try {
			const result = await api.get<{ tables: string[] }>(`/api/v1/platform/apps/${$page.params.id}/tables`);
			tables = (result.tables || []).map(n => ({ name: n }));
		} catch {}
		loading = false;
	});

	async function selectTable(name: string) {
		selectedTable = name;
		loadingData = true;
		try {
			tableData = await api.get<Record<string, unknown>[]>(`/api/v1/platform/apps/${$page.params.id}/tables/${name}`);
		} catch { tableData = []; }
		loadingData = false;
	}
</script>

<div class="flex gap-6">
	<div class="w-56 flex-shrink-0">
		<Card title="Tables">
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
							{t.name}
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
		{#if !selectedTable}
			<EmptyState icon="▤" title="Select a table" description="Choose a table from the sidebar to browse its contents." />
		{:else if loadingData}
			<Card title={selectedTable}>
				<Skeleton rows={5} />
			</Card>
		{:else if tableData.length === 0}
			<Card title={selectedTable}>
				<p class="text-sm" style="color: var(--text-secondary)">No data in this table.</p>
			</Card>
		{:else}
			<Card title={`${selectedTable} (${tableData.length} rows)`}>
				<div class="overflow-x-auto -mx-5 -mb-5">
					<div class="table-wrap">
						<table class="w-full">
							<thead>
								<tr>
									{#each Object.keys(tableData[0]) as col}
										<th>{col}</th>
									{/each}
								</tr>
							</thead>
							<tbody>
								{#each tableData as row}
									<tr>
										{#each Object.keys(tableData[0]) as col}
											<td class="font-mono text-xs max-w-[200px] truncate">{JSON.stringify(row[col]) ?? 'NULL'}</td>
										{/each}
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</div>
			</Card>
		{/if}
	</div>
</div>
