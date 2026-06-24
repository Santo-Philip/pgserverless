<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';

	let tables = $state<string[]>([]);
	let loading = $state(true);
	let selectedTable = $state<string | null>(null);
	let tableData = $state<Record<string, unknown>[]>([]);

	onMount(async () => {
		try {
			const result = await api.get<{ tables: string[] }>(`/api/v1/platform/apps/${$page.params.id}/tables`);
			tables = result.tables || [];
		} catch (e) {
			console.error('Failed to load tables', e);
		} finally {
			loading = false;
		}
	});

	async function selectTable(name: string) {
		selectedTable = name;
		try {
			tableData = await api.get<Record<string, unknown>[]>(`/api/v1/platform/apps/${$page.params.id}/tables/${name}`);
		} catch (e) {
			console.error('Failed to load table data', e);
			tableData = [];
		}
	}
</script>

<div class="max-w-6xl mx-auto">
	<a href="/apps/{$page.params.id}" class="text-nexbic-600 hover:underline mb-4 inline-block">&larr; Back to App</a>
	<h1 class="text-3xl font-bold mb-8">Database Explorer</h1>

	{#if loading}
		<p class="text-gray-500">Loading...</p>
	{:else}
		<div class="flex gap-6">
			<div class="w-64 flex-shrink-0">
				<div class="bg-white rounded-lg shadow p-4">
					<h2 class="font-semibold mb-3">Tables</h2>
					{#if tables.length === 0}
						<p class="text-sm text-gray-500">No tables yet.</p>
					{:else}
						<ul class="space-y-1">
							{#each tables as table}
								<li>
									<button
										onclick={() => selectTable(table)}
										class="w-full text-left px-3 py-2 rounded text-sm hover:bg-gray-100 {selectedTable === table ? 'bg-nexbic-50 text-nexbic-700 font-medium' : ''}"
									>
										{table}
									</button>
								</li>
							{/each}
						</ul>
					{/if}
				</div>
			</div>

			<div class="flex-1">
				{#if selectedTable}
					<div class="bg-white rounded-lg shadow overflow-hidden">
						<h2 class="text-lg font-semibold p-4 border-b">{selectedTable}</h2>
						{#if tableData.length === 0}
							<p class="p-4 text-gray-500">No data in this table.</p>
						{:else}
							<div class="overflow-x-auto">
								<table class="w-full">
									<thead class="bg-gray-50">
										<tr>
											{#each Object.keys(tableData[0]) as col}
												<th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">{col}</th>
											{/each}
										</tr>
									</thead>
									<tbody class="divide-y divide-gray-200">
										{#each tableData as row}
											<tr>
												{#each Object.keys(tableData[0]) as col}
													<td class="px-4 py-2 text-sm">{JSON.stringify(row[col])}</td>
												{/each}
											</tr>
										{/each}
									</tbody>
								</table>
							</div>
						{/if}
					</div>
				{:else}
					<div class="bg-white rounded-lg shadow p-12 text-center">
						<p class="text-gray-500">Select a table to browse its contents.</p>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>
