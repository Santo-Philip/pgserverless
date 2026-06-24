<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';

	let app = $state<App | null>(null);
	let loading = $state(true);

	onMount(async () => {
		try {
			app = await api.getApp($page.params.id);
		} catch (e) {
			console.error('Failed to load app', e);
		} finally {
			loading = false;
		}
	});
</script>

<div class="max-w-6xl mx-auto">
	<a href="/apps" class="text-nexbic-600 hover:underline mb-4 inline-block">&larr; Back to Apps</a>

	{#if loading}
		<p class="text-gray-500">Loading...</p>
	{:else if app}
		<h1 class="text-3xl font-bold mb-2">{app.name}</h1>
		<p class="text-lg text-gray-500 font-mono mb-8">/api/v1/{app.slug}</p>

		<div class="grid grid-cols-1 md:grid-cols-6 gap-4 mb-8">
			<a href="/apps/{app.id}" class="bg-nexbic-600 text-white px-4 py-2 rounded text-center no-underline">
				Overview
			</a>
			<a href="/apps/{app.id}/tables" class="bg-white text-gray-700 px-4 py-2 rounded text-center no-underline border hover:bg-gray-50">
				Tables
			</a>
			<a href="/apps/{app.id}/sql" class="bg-white text-gray-700 px-4 py-2 rounded text-center no-underline border hover:bg-gray-50">
				SQL Editor
			</a>
			<a href="/apps/{app.id}/domains" class="bg-white text-gray-700 px-4 py-2 rounded text-center no-underline border hover:bg-gray-50">
				Domains
			</a>
			<a href="/apps/{app.id}/api-keys" class="bg-white text-gray-700 px-4 py-2 rounded text-center no-underline border hover:bg-gray-50">
				API Keys
			</a>
			<a href="/apps/{app.id}/logs" class="bg-white text-gray-700 px-4 py-2 rounded text-center no-underline border hover:bg-gray-50">
				Logs
			</a>
		</div>

		<div class="bg-white rounded-lg shadow p-6">
			<h2 class="text-xl font-semibold mb-4">App Details</h2>
			<dl class="grid grid-cols-2 gap-4">
				<div>
					<dt class="text-sm text-gray-500">App ID</dt>
					<dd class="font-mono text-sm">{app.id}</dd>
				</div>
				<div>
					<dt class="text-sm text-gray-500">Schema</dt>
					<dd class="font-mono text-sm">{app.schema_name}</dd>
				</div>
				<div>
					<dt class="text-sm text-gray-500">Status</dt>
					<dd class="capitalize">{app.status}</dd>
				</div>
				<div>
					<dt class="text-sm text-gray-500">Created</dt>
					<dd>{new Date(app.created_at).toLocaleDateString()}</dd>
				</div>
			</dl>
		</div>
	{:else}
		<p class="text-red-500">App not found</p>
	{/if}
</div>
