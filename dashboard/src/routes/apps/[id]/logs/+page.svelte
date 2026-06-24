<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';
	import Card from '$lib/components/Card.svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import Badge from '$lib/components/Badge.svelte';

	let logs = $state<AuditLog[]>([]);
	let loading = $state(true);
	let filter = $state('');

	onMount(async () => {
		try {
			const result = await api.get<AuditLog[]>(`/api/v1/platform/apps/${$page.params.id}/logs`);
			logs = result;
		} catch {}
		loading = false;
	});

	let filtered = $derived(
		filter ? logs.filter(l => l.path.toLowerCase().includes(filter.toLowerCase()) || l.method.toLowerCase().includes(filter.toLowerCase())) : logs
	);
</script>

<Card title="Request Logs" description="API request history for this application">
	<div class="flex justify-end mb-4 -mt-1">
		<input type="text" bind:value={filter} class="input !w-48" placeholder="Filter..." />
	</div>
	{#if loading}
		<Skeleton rows={5} />
	{:else if logs.length === 0}
		<EmptyState title="No logs yet" description="API requests will appear here." />
	{:else}
		<div class="table-wrap overflow-x-auto -mx-5 -mb-5">
			<table class="w-full">
				<thead>
					<tr>
						<th>Method</th>
						<th>Path</th>
						<th>Status</th>
						<th>Time (ms)</th>
						<th>Timestamp</th>
					</tr>
				</thead>
				<tbody>
					{#each filtered as log}
						<tr>
							<td>
								<span class="badge text-xs font-mono" style="background-color: {log.method === 'GET' ? 'rgba(34,197,94,0.1)' : log.method === 'POST' ? 'rgba(12,142,229,0.1)' : log.method === 'PATCH' ? 'rgba(245,158,11,0.1)' : 'rgba(239,68,68,0.1)'}; color: {log.method === 'GET' ? 'var(--success)' : log.method === 'POST' ? 'var(--accent)' : log.method === 'PATCH' ? 'var(--warning)' : 'var(--danger)'}">
									{log.method}
								</span>
							</td>
							<td class="font-mono text-xs">{log.path}</td>
							<td>{log.status_code}</td>
							<td>{log.response_time_ms}</td>
							<td class="text-xs" style="color: var(--text-tertiary)">{new Date(log.created_at).toLocaleString()}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</Card>
