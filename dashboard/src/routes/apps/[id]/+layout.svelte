<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';
	import Tabs from '$lib/components/Tabs.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import Breadcrumbs from '$lib/components/Breadcrumbs.svelte';
	import LoadingCard from '$lib/components/LoadingCard.svelte';
	import type { Snippet } from 'svelte';

	let { children }: { children: Snippet } = $props();

	let app = $state<App | null>(null);
	let loading = $state(true);

	let tabs = $derived([
		{ label: 'Overview', href: `/apps/${$page.params.id}` },
		{ label: 'Database', href: `/apps/${$page.params.id}/database` },
		{ label: 'REST API', href: `/apps/${$page.params.id}/rest-api` },
		{ label: 'API Keys', href: `/apps/${$page.params.id}/api-keys` },
		{ label: 'Extensions', href: `/apps/${$page.params.id}/extensions` },
		{ label: 'Logs', href: `/apps/${$page.params.id}/logs` },
		{ label: 'Settings', href: `/apps/${$page.params.id}/settings` },
	]);

	onMount(async () => {
		try {
			app = await api.getApp($page.params.id);
		} catch {}
		loading = false;
	});
</script>

<div class="max-w-7xl mx-auto">
	<Breadcrumbs items={[
		{ label: 'Applications', href: '/apps' },
		{ label: app?.name || '...' },
	]} />

	{#if loading}
		<div class="space-y-4"><LoadingCard /><LoadingCard /></div>
	{:else if !app}
		<div class="card p-12 text-center" style="color: var(--text-secondary)">Application not found</div>
	{:else}
		<div class="flex items-start justify-between mb-6">
			<div>
				<h1 class="text-2xl font-bold">{app.name}</h1>
				<p class="text-sm mt-1 font-mono" style="color: var(--text-secondary)">/api/v1/{app.slug}</p>
			</div>
			<Badge status={app.status} />
		</div>

		<Tabs {tabs} />

		<div class="mt-6">
			{@render children()}
		</div>
	{/if}
</div>
