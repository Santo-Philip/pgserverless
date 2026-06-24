<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';

	let domains = $state<Domain[]>([]);
	let loading = $state(true);
	let newDomain = $state('');
	let error = $state('');

	onMount(async () => {
		try {
			domains = await api.listDomains($page.params.id);
		} catch (e) {
			console.error('Failed to load domains', e);
		} finally {
			loading = false;
		}
	});

	async function handleAdd() {
		error = '';
		try {
			const domain = await api.createDomain($page.params.id, newDomain);
			domains = [domain, ...domains];
			newDomain = '';
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to add domain';
		}
	}

	async function handleVerify(domainId: string) {
		try {
			await api.verifyDomain($page.params.id, domainId);
			domains = domains.map((d) =>
				d.id === domainId ? { ...d, verified: true, status: 'active' as const } : d
			);
		} catch (e) {
			console.error('Failed to verify domain', e);
		}
	}

	async function handleDelete(domainId: string) {
		if (!confirm('Delete this domain?')) return;
		try {
			await api.deleteDomain($page.params.id, domainId);
			domains = domains.filter((d) => d.id !== domainId);
		} catch (e) {
			console.error('Failed to delete domain', e);
		}
	}
</script>

<div class="max-w-6xl mx-auto">
	<a href="/apps/{$page.params.id}" class="text-nexbic-600 hover:underline mb-4 inline-block">&larr; Back to App</a>
	<h1 class="text-3xl font-bold mb-8">Custom Domains</h1>

	<div class="bg-white rounded-lg shadow p-6 mb-6">
		<h2 class="text-xl font-semibold mb-4">Add Domain</h2>
		<form onsubmit={handleAdd} class="space-y-4">
			{#if error}
				<div class="bg-red-50 text-red-700 p-3 rounded text-sm">{error}</div>
			{/if}

			<div>
				<label for="domain" class="block text-sm font-medium text-gray-700 mb-1">Domain</label>
				<input
					id="domain"
					type="text"
					bind:value={newDomain}
					placeholder="api.example.com"
					required
					class="w-full px-3 py-2 border rounded font-mono"
				/>
			</div>

			<button type="submit" class="bg-nexbic-600 text-white px-4 py-2 rounded hover:bg-nexbic-700">
				Add Domain
			</button>
		</form>
	</div>

	{#if loading}
		<p class="text-gray-500">Loading...</p>
	{:else if domains.length === 0}
		<div class="bg-white rounded-lg shadow p-12 text-center">
			<p class="text-gray-500">No custom domains yet. Add one above.</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow overflow-hidden">
			<table class="w-full">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Domain</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Verified</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-200">
					{#each domains as domain}
						<tr>
							<td class="px-6 py-4 font-mono">{domain.domain}</td>
							<td class="px-6 py-4">
								<span class="text-xs uppercase px-2 py-1 rounded {domain.status === 'active' ? 'bg-green-100 text-green-700' : 'bg-yellow-100 text-yellow-700'}">
									{domain.status}
								</span>
							</td>
							<td class="px-6 py-4">
								{domain.verified ? 'Yes' : 'No'}
							</td>
							<td class="px-6 py-4 space-x-2">
								{#if !domain.verified}
									<button
										onclick={() => handleVerify(domain.id)}
										class="text-nexbic-600 hover:underline text-sm"
									>
										Verify
									</button>
								{/if}
								<button
									onclick={() => handleDelete(domain.id)}
									class="text-red-500 hover:underline text-sm"
								>
									Delete
								</button>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
