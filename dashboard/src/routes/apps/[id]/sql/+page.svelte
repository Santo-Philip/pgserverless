<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';
	import Card from '$lib/components/Card.svelte';
	import CopyButton from '$lib/components/CopyButton.svelte';

	let query = $state('SELECT * FROM ...');
	let results = $state<Record<string, unknown>[] | null>(null);
	let error = $state('');
	let running = $state(false);
	let executionTime = $state<number | null>(null);
	let history = $state<string[]>([]);
	let showHistory = $state(false);

	async function handleRun() {
		error = '';
		running = true;
		results = null;
		executionTime = null;
		let start = performance.now();

		try {
			results = await api.post<Record<string, unknown>[]>(`/api/v1/platform/apps/${$page.params.id}/sql`, { query });
			history = [query, ...history.filter(h => h !== query)].slice(0, 20);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Query failed';
		} finally {
			executionTime = performance.now() - start;
			running = false;
		}
	}
</script>

<div class="space-y-4">
	<Card title="SQL Editor" description="Queries are scoped to this application's schema">
		<div class="space-y-3">
			<div class="flex items-center gap-2">
				<button onclick={() => showHistory = !showHistory} class="btn btn-ghost btn-sm">History ({history.length})</button>
			</div>

			{#if showHistory && history.length > 0}
				<div class="space-y-1 max-h-32 overflow-y-auto p-3 rounded-lg" style="background-color: var(--bg-tertiary);">
					{#each history as h}
						<button onclick={() => { query = h; showHistory = false; }} class="block w-full text-left text-xs font-mono px-2 py-1.5 rounded hover:bg-hover" style="color: var(--text-secondary)">{h}</button>
					{/each}
				</div>
			{/if}

			<textarea
				bind:value={query}
				rows="8"
				class="input font-mono text-sm"
				style="resize: vertical; min-height: 120px"
				placeholder="SELECT * FROM ..."
			></textarea>

			<div class="flex items-center justify-between">
				<div class="flex gap-2">
					<button onclick={handleRun} disabled={running} class="btn btn-primary">
						{running ? 'Running...' : 'Run Query'}
					</button>
				</div>
			</div>
		</div>
	</Card>

	{#if error}
		<div class="px-4 py-3 rounded-lg text-sm font-mono" style="background-color: rgba(239,68,68,0.1); color: var(--danger); white-space: pre-wrap">{error}</div>
	{/if}

	{#if results !== null}
		<Card title={`Results`} description={executionTime !== null ? `${results.length} rows in ${executionTime.toFixed(1)}ms` : ''}>
			{#if results.length === 0}
				<p class="text-sm" style="color: var(--text-secondary)">Query executed successfully. No rows returned.</p>
			{:else}
				<div class="overflow-x-auto -mx-5 -mb-5">
					<div class="table-wrap">
						<table class="w-full">
							<thead>
								<tr>
									{#each Object.keys(results[0]) as col}
										<th>{col}</th>
									{/each}
								</tr>
							</thead>
							<tbody>
								{#each results as row}
									<tr>
										{#each Object.keys(results[0]) as col}
											<td class="font-mono text-xs max-w-[200px] truncate">{JSON.stringify(row[col]) ?? 'NULL'}</td>
										{/each}
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
