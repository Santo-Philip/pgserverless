<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';
	import Card from '$lib/components/Card.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import StatCard from '$lib/components/StatCard.svelte';
	import LoadingCard from '$lib/components/LoadingCard.svelte';
	import type { App } from '$lib/types';

	let app = $state<App | null>(null);
	let loading = $state(true);

	onMount(async () => {
		try {
			app = await api.getApp($page.params.id!);
		} catch {}
		loading = false;
	});
</script>

{#if loading}
	<div class="space-y-4"><LoadingCard /><LoadingCard /></div>
{:else if app}
	<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
		<StatCard title="Status" value={app.status} subtitle="Current state" />
		<StatCard title="Schema" value={app.schema_name} subtitle="PostgreSQL schema" />
		<StatCard title="Region" value={app.region} subtitle="Deployment region" />
		<StatCard title="Created" value={new Date(app.created_at).toLocaleDateString()} subtitle={new Date(app.created_at).toLocaleTimeString()} />
	</div>

	<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
		<Card title="Application Details">
			<dl class="space-y-3">
				<div class="flex justify-between"><span class="text-sm" style="color: var(--text-secondary)">App ID</span><span class="text-sm font-mono">{app.id}</span></div>
				<div class="flex justify-between"><span class="text-sm" style="color: var(--text-secondary)">Name</span><span class="text-sm">{app.name}</span></div>
				<div class="flex justify-between"><span class="text-sm" style="color: var(--text-secondary)">Slug</span><span class="text-sm font-mono">{app.slug}</span></div>
				<div class="flex justify-between"><span class="text-sm" style="color: var(--text-secondary)">Schema</span><span class="text-sm font-mono">{app.schema_name}</span></div>
				<div class="flex justify-between"><span class="text-sm" style="color: var(--text-secondary)">Visibility</span><span class="text-sm">{app.visibility}</span></div>
				<div class="flex justify-between"><span class="text-sm" style="color: var(--text-secondary)">Status</span><Badge status={app.status} /></div>
				<div class="flex justify-between"><span class="text-sm" style="color: var(--text-secondary)">Created</span><span class="text-sm">{new Date(app.created_at).toLocaleString()}</span></div>
				<div class="flex justify-between"><span class="text-sm" style="color: var(--text-secondary)">Updated</span><span class="text-sm">{new Date(app.updated_at).toLocaleString()}</span></div>
			</dl>
		</Card>

		<Card title="REST Endpoint">
			<div class="space-y-3">
				<div>
					<p class="text-xs mb-1" style="color: var(--text-tertiary)">API Base URL</p>
					<div class="flex items-center justify-between p-3 rounded-lg font-mono text-sm" style="background-color: var(--bg-tertiary);">
						<span>/api/v1/{app.slug}</span>
						<button onclick={() => navigator.clipboard.writeText(`/api/v1/${app!.slug}`)} class="btn btn-ghost btn-sm">Copy</button>
					</div>
				</div>
				<div>
					<p class="text-xs mb-1" style="color: var(--text-tertiary)">Example: GET /users</p>
					<div class="flex items-center justify-between p-3 rounded-lg font-mono text-sm" style="background-color: var(--bg-tertiary);">
						<span>GET /api/v1/{app.slug}/users</span>
						<button onclick={() => navigator.clipboard.writeText(`/api/v1/${app!.slug}/users`)} class="btn btn-ghost btn-sm">Copy</button>
					</div>
				</div>
			</div>
		</Card>
	</div>
{/if}
