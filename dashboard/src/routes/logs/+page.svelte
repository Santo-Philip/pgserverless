<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import Card from '$lib/components/Card.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import Breadcrumbs from '$lib/components/Breadcrumbs.svelte';
	import type { AuditLog } from '$lib/types';

	let logs = $state<AuditLog[]>([]);
	let loading = $state(true);
	let filter = $state('');
	let levelFilter = $state('');

	onMount(async () => {
		try {
			const result = await api.listApps();
			let all: AuditLog[] = [];
			for (const app of result.data) {
				try {
					const appLogs = await api.get<AuditLog[]>(`/api/v1/platform/apps/${app.id}/logs`);
					all = [...all, ...appLogs];
				} catch {}
			}
			all.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
			logs = all;
		} catch {}
		loading = false;
	});
</script>

<Breadcrumbs items={[{ label: 'Logs' }]} />

<div class="max-w-7xl mx-auto">
	<div class="flex items-center justify-between mb-8">
		<div>
			<h1 class="text-2xl font-bold">Logs</h1>
			<p class="text-sm mt-1" style="color: var(--text-secondary)">API request logs across all applications</p>
		</div>
		<div class="flex gap-2">
			<input type="text" bind:value={filter} class="input !w-48" placeholder="Search path or method..." />
		</div>
	</div>

	{#if loading}
		<Skeleton rows={8} />
	{:else if logs.length === 0}
		<EmptyState icon="☰" title="No logs yet" description="API requests will appear here when applications receive traffic." />
	{:else}
		<Card>
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
						{#each logs.filter(l => !filter || l.path.toLowerCase().includes(filter.toLowerCase()) || l.method.toLowerCase().includes(filter.toLowerCase())) as log}
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
		</Card>
	{/if}
</div>
