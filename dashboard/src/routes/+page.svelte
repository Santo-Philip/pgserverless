<script lang="ts">
	import { onMount } from 'svelte';
	import { isAuthenticated } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api/client';

	let stats = $state({
		totalApps: 0,
		totalKeys: 0
	});

	let loading = $state(true);

	onMount(async () => {
		if (!$isAuthenticated) {
			goto('/login');
			return;
		}

		try {
			const apps = await api.listApps();
			stats.totalApps = apps.total;
		} catch (e) {
			console.error('Failed to load stats', e);
		} finally {
			loading = false;
		}
	});
</script>

<div class="max-w-6xl mx-auto">
	<h1 class="text-3xl font-bold mb-8">Dashboard</h1>

	{#if loading}
		<p class="text-gray-500">Loading...</p>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
			<div class="bg-white rounded-lg shadow p-6">
				<h3 class="text-sm font-medium text-gray-500 uppercase">Total Apps</h3>
				<p class="text-3xl font-bold mt-2">{stats.totalApps}</p>
			</div>
			<div class="bg-white rounded-lg shadow p-6">
				<h3 class="text-sm font-medium text-gray-500 uppercase">API Keys</h3>
				<p class="text-3xl font-bold mt-2">{stats.totalKeys}</p>
			</div>
			<div class="bg-white rounded-lg shadow p-6">
				<h3 class="text-sm font-medium text-gray-500 uppercase">Status</h3>
				<p class="text-3xl font-bold mt-2 text-green-600">Healthy</p>
			</div>
		</div>

		<div class="bg-white rounded-lg shadow p-6">
			<div class="flex justify-between items-center mb-4">
				<h2 class="text-xl font-semibold">Applications</h2>
				<a href="/apps" class="bg-nexbic-600 text-white px-4 py-2 rounded hover:bg-nexbic-700 transition-colors no-underline">
					New App
				</a>
			</div>
			<p class="text-gray-500">Your applications will appear here.</p>
		</div>
	{/if}
</div>
