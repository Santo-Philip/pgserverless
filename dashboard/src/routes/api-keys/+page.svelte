<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import Card from '$lib/components/Card.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import Breadcrumbs from '$lib/components/Breadcrumbs.svelte';

	let keys: { app_name: string; keys: APIKey[] }[] = [];
	let loading = $state(true);

	onMount(async () => {
		try {
			const result = await api.listApps();
			for (const app of result.data) {
				try {
					const appKeys = await api.listAPIKeys(app.id);
					if (appKeys.length > 0) keys.push({ app_name: app.name, keys: appKeys });
				} catch {}
			}
		} catch {}
		loading = false;
	});
</script>

<Breadcrumbs items={[{ label: 'API Keys' }]} />

<div class="max-w-7xl mx-auto">
	<div class="mb-8">
		<h1 class="text-2xl font-bold">API Keys</h1>
		<p class="text-sm mt-1" style="color: var(--text-secondary)">All API keys across your applications</p>
	</div>

	{#if loading}
		<Skeleton rows={4} />
	{:else if keys.length === 0}
		<EmptyState icon="🔑" title="No API keys" description="Generate API keys from the application view." />
	{:else}
		<div class="space-y-6">
			{#each keys as group}
				<Card title={group.app_name}>
					<div class="table-wrap overflow-x-auto -mx-5 -mb-5">
						<table class="w-full">
							<thead><tr><th>Name</th><th>Type</th><th>Prefix</th><th>Status</th><th>Created</th></tr></thead>
							<tbody>
								{#each group.keys as key}
									<tr>
										<td class="font-medium">{key.name}</td>
										<td><Badge status={key.key_type} /></td>
										<td><span class="font-mono text-xs" style="color: var(--text-secondary)">{key.key_prefix}...</span></td>
										<td><Badge status={key.is_active ? 'active' : 'inactive'} /></td>
										<td class="text-xs" style="color: var(--text-tertiary)">{new Date(key.created_at).toLocaleDateString()}</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</Card>
			{/each}
		</div>
	{/if}
</div>
