<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';

	let keys = $state<APIKey[]>([]);
	let loading = $state(true);

	onMount(async () => {
		try {
			keys = await api.listAPIKeys($page.params.id);
		} catch (e) {
			console.error('Failed to load API keys', e);
		} finally {
			loading = false;
		}
	});
</script>

<div class="max-w-6xl mx-auto">
	<a href="/apps/{$page.params.id}" class="text-nexbic-600 hover:underline mb-4 inline-block">&larr; Back to App</a>
	<h1 class="text-3xl font-bold mb-8">API Keys</h1>

	{#if loading}
		<p class="text-gray-500">Loading...</p>
	{:else if keys.length === 0}
		<div class="bg-white rounded-lg shadow p-12 text-center">
			<p class="text-gray-500">No API keys yet. Create one from the app overview.</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow overflow-hidden">
			<table class="w-full">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Type</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Prefix</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Created</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-200">
					{#each keys as key}
						<tr>
							<td class="px-6 py-4 font-medium">{key.name}</td>
							<td class="px-6 py-4">
								<span class="text-xs uppercase px-2 py-1 rounded bg-blue-100 text-blue-700">{key.key_type}</span>
							</td>
							<td class="px-6 py-4 font-mono text-sm">{key.key_prefix}...</td>
							<td class="px-6 py-4">
								<span class="text-xs uppercase px-2 py-1 rounded {key.is_active ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}">
									{key.is_active ? 'Active' : 'Inactive'}
								</span>
							</td>
							<td class="px-6 py-4 text-sm text-gray-500">{new Date(key.created_at).toLocaleDateString()}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
