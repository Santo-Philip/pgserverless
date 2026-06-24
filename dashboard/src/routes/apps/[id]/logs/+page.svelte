<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';

	let logs = $state<AuditLog[]>([]);
	let loading = $state(true);

	onMount(async () => {
		try {
			const result = await api.get<AuditLog[]>(`/api/v1/platform/apps/${$page.params.id}/logs`);
			logs = result;
		} catch (e) {
			console.error('Failed to load logs', e);
		} finally {
			loading = false;
		}
	});
</script>

<div class="max-w-6xl mx-auto">
	<a href="/apps/{$page.params.id}" class="text-nexbic-600 hover:underline mb-4 inline-block">&larr; Back to App</a>
	<h1 class="text-3xl font-bold mb-8">Logs</h1>

	{#if loading}
		<p class="text-gray-500">Loading...</p>
	{:else if logs.length === 0}
		<div class="bg-white rounded-lg shadow p-12 text-center">
			<p class="text-gray-500">No logs yet.</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow overflow-hidden">
			<table class="w-full">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Method</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Path</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Time (ms)</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Timestamp</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-200">
					{#each logs as log}
						<tr>
							<td class="px-6 py-4">
								<span class="text-xs font-mono px-2 py-1 rounded {log.method === 'GET' ? 'bg-green-100 text-green-700' : log.method === 'POST' ? 'bg-blue-100 text-blue-700' : log.method === 'DELETE' ? 'bg-red-100 text-red-700' : 'bg-gray-100 text-gray-700'}">
									{log.method}
								</span>
							</td>
							<td class="px-6 py-4 font-mono text-sm">{log.path}</td>
							<td class="px-6 py-4">{log.status_code}</td>
							<td class="px-6 py-4">{log.response_time_ms}</td>
							<td class="px-6 py-4 text-sm text-gray-500">{new Date(log.created_at).toLocaleString()}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
