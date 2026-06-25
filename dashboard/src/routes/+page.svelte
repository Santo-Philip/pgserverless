<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import StatCard from '$lib/components/StatCard.svelte';
	import Card from '$lib/components/Card.svelte';
	import LoadingCard from '$lib/components/LoadingCard.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import { goto } from '$app/navigation';
	import type { App } from '$lib/types';

	let loading = $state(true);
	let stats = $state({
		apps: 0,
		users: 0,
		apiKeys: 0,
		schemas: 0,
		requestsToday: 0,
		dbSize: '0 MB',
		storageUsed: '0 MB',
	});
	let recentApps = $state<App[]>([]);

	onMount(async () => {
		try {
			const apps = await api.listApps();
			stats.apps = apps.length;
			recentApps = apps.slice(0, 5);
		} catch {}
		loading = false;
	});

	let hoveredApp = $state<string | null>(null);
</script>

<div class="max-w-7xl mx-auto">
	<div class="flex items-center justify-between mb-8">
		<div>
			<h1 class="text-2xl font-bold">Dashboard</h1>
			<p class="text-sm mt-1" style="color: var(--text-secondary)">Platform overview</p>
		</div>
		<div class="flex items-center gap-2">
			<div class="flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs" style="background-color: rgba(34,197,94,0.1); color: var(--success)">
				<span class="w-1.5 h-1.5 rounded-full" style="background-color: var(--success)"></span>
				All Systems Normal
			</div>
		</div>
	</div>

	{#if loading}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
			{#each Array(4) as _}<LoadingCard />{/each}
		</div>
	{:else}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
			<StatCard title="Applications" value={stats.apps} icon="▦" subtitle="Total active applications" />
			<StatCard title="Database Size" value={stats.dbSize} icon="▤" subtitle="Across all schemas" />
			<StatCard title="API Keys" value={stats.apiKeys} icon="🔑" subtitle="Active keys" />
			<StatCard title="Active Schemas" value={stats.schemas} icon="◆" subtitle="PostgreSQL schemas" />
		</div>

		<div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
			<div class="space-y-4">
				<Card title="Recent Applications">
					<div class="flex items-center justify-end mb-3 -mt-1">
						<a href="/apps" class="text-xs link">View all</a>
					</div>
					{#if recentApps.length === 0}
						<EmptyState title="No applications yet" description="Create your first application to get started.">
							<button onclick={() => goto('/apps')} class="btn btn-primary btn-sm">Create App</button>
						</EmptyState>
					{:else}
						<div class="space-y-2">
							{#each recentApps as app}
								<a
									href="/apps/{app.id}"
									class="flex items-center justify-between p-3 rounded-lg transition-colors no-underline"
									style="color: var(--text-primary); background-color: {hoveredApp === app.id ? 'var(--bg-hover)' : 'transparent'}"
									onmouseenter={() => hoveredApp = app.id}
									onmouseleave={() => hoveredApp = null}
								>
									<div>
										<span class="text-sm font-medium">{app.name}</span>
										<p class="text-xs mt-0.5" style="color: var(--text-tertiary)">/api/v1/{app.slug}</p>
									</div>
									<span class="text-xs" style="color: var(--text-tertiary)">{new Date(app.created_at).toLocaleDateString()}</span>
								</a>
							{/each}
						</div>
					{/if}
				</Card>
			</div>

			<div class="space-y-4">
				<Card title="Quick Actions">
					<div class="grid grid-cols-2 gap-3">
						<a href="/apps" class="flex flex-col items-center gap-2 p-4 rounded-lg transition-colors no-underline text-center" style="color: var(--text-secondary); background-color: var(--bg-tertiary)">
							<span class="text-xl">▦</span>
							<span class="text-xs">Create Application</span>
						</a>
						<a href="/database" class="flex flex-col items-center gap-2 p-4 rounded-lg transition-colors no-underline text-center" style="color: var(--text-secondary); background-color: var(--bg-tertiary)">
							<span class="text-xl">⚡</span>
							<span class="text-xs">Open SQL Editor</span>
						</a>
						<a href="/api-keys" class="flex flex-col items-center gap-2 p-4 rounded-lg transition-colors no-underline text-center" style="color: var(--text-secondary); background-color: var(--bg-tertiary)">
							<span class="text-xl">🔑</span>
							<span class="text-xs">Generate API Key</span>
						</a>
						<a href="/logs" class="flex flex-col items-center gap-2 p-4 rounded-lg transition-colors no-underline text-center" style="color: var(--text-secondary); background-color: var(--bg-tertiary)">
							<span class="text-xl">☰</span>
							<span class="text-xs">View Logs</span>
						</a>
					</div>
				</Card>
			</div>
		</div>
	{/if}
</div>
