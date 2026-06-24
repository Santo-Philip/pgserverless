<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';

	let query = $state('SELECT * FROM ...');
	let results = $state<Record<string, unknown>[] | null>(null);
	let error = $state('');
	let running = $state(false);

	async function handleRun() {
		error = '';
		running = true;
		results = null;

		try {
			results = await api.post<Record<string, unknown>[]>(`/api/v1/platform/apps/${$page.params.id}/sql`, { query });
		} catch (e) {
			error = e instanceof Error ? e.message : 'Query failed';
		} finally {
			running = false;
		}
	}
</script>

<div class="max-w-6xl mx-auto">
	<a href="/apps/{$page.params.id}" class="text-nexbic-600 hover:underline mb-4 inline-block">&larr; Back to App</a>
	<h1 class="text-3xl font-bold mb-8">SQL Editor</h1>

	<div class="bg-white rounded-lg shadow p-6 mb-6">
		<form onsubmit={handleRun} class="space-y-4">
			<div>
				<label for="sql" class="block text-sm font-medium text-gray-700 mb-1">SQL Query</label>
				<textarea
					id="sql"
					bind:value={query}
					rows="8"
					class="w-full px-3 py-2 border rounded font-mono text-sm"
					placeholder="SELECT * FROM ..."
				></textarea>
			</div>

			<button
				type="submit"
				disabled={running}
				class="bg-nexbic-600 text-white px-6 py-2 rounded hover:bg-nexbic-700 disabled:opacity-50"
			>
				{running ? 'Running...' : 'Run Query'}
			</button>
		</form>
	</div>

	{#if error}
		<div class="bg-red-50 text-red-700 p-4 rounded mb-6">{error}</div>
	{/if}

	{#if results}
		<div class="bg-white rounded-lg shadow overflow-hidden">
			<h2 class="text-lg font-semibold p-4 border-b">Results ({results.length} rows)</h2>
			{#if results.length === 0}
				<p class="p-4 text-gray-500">Query executed successfully. No rows returned.</p>
			{:else}
				<div class="overflow-x-auto">
					<table class="w-full">
						<thead class="bg-gray-50">
							<tr>
								{#each Object.keys(results[0]) as col}
									<th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">{col}</th>
								{/each}
							</tr>
						</thead>
						<tbody class="divide-y divide-gray-200">
							{#each results as row}
								<tr>
									{#each Object.keys(results[0]) as col}
										<td class="px-4 py-2 text-sm font-mono">{JSON.stringify(row[col])}</td>
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
