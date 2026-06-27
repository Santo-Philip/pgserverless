<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import type { AuditLog } from '$lib/types';
	import Card from '$lib/components/Card.svelte';
	import LoadingCard from '$lib/components/LoadingCard.svelte';

	let loading = $state(true);
	let logs: AuditLog[] = $state([]);

	onMount(load);

	async function load() {
		loading = true;
		try {
			logs = await api.listAuditLogs();
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}
</script>

<div class="max-w-5xl mx-auto">
	<h1 class="text-2xl font-bold mb-6" style="color: var(--text-primary);">Audit Logs</h1>

	{#if loading}
		<LoadingCard />
	{:else if logs.length === 0}
		<Card title="No Audit Logs"><p class="text-sm" style="color: var(--text-secondary);">No audit log entries yet.</p></Card>
	{:else}
		<div class="overflow-x-auto rounded-xl border" style="border-color: var(--border-primary);">
			<table class="w-full text-sm">
				<thead>
					<tr style="background-color: var(--bg-hover);">
						<th class="px-4 py-3 text-left font-medium" style="color: var(--text-primary);">Time</th>
						<th class="px-4 py-3 text-left font-medium" style="color: var(--text-primary);">Action</th>
						<th class="px-4 py-3 text-left font-medium" style="color: var(--text-primary);">Resource</th>
						<th class="px-4 py-3 text-left font-medium" style="color: var(--text-primary);">Resource ID</th>
						<th class="px-4 py-3 text-left font-medium" style="color: var(--text-primary);">Actor</th>
						<th class="px-4 py-3 text-left font-medium" style="color: var(--text-primary);">IP</th>
					</tr>
				</thead>
				<tbody>
					{#each logs as log}
						<tr class="border-t" style="border-color: var(--border-primary);">
							<td class="px-4 py-3" style="color: var(--text-secondary);">{new Date(log.created_at).toLocaleString()}</td>
							<td class="px-4 py-3">
								<span class="px-2 py-0.5 rounded-full text-xs font-medium"
									style={log.action === 'create' ? 'background-color: #dcfce7; color: #166534;' :
										log.action === 'delete' ? 'background-color: #fef2f2; color: #991b1b;' :
										'background-color: #f3f4f6; color: #374151;'}>
									{log.action}
								</span>
							</td>
							<td class="px-4 py-3" style="color: var(--text-secondary);">{log.resource}</td>
							<td class="px-4 py-3 font-mono text-xs" style="color: var(--text-secondary);">{log.resource_id || '-'}</td>
							<td class="px-4 py-3 font-mono text-xs" style="color: var(--text-secondary);">{log.actor_id}</td>
							<td class="px-4 py-3 text-xs" style="color: var(--text-secondary);">{log.ip_address || '-'}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
