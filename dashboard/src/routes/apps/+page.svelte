<script lang="ts">
	import { onMount } from 'svelte';
	import { isAuthenticated } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api/client';

	let apps = $state<App[]>([]);
	let loading = $state(true);
	let showCreateForm = $state(false);
	let newName = $state('');
	let newSlug = $state('');
	let error = $state('');

	onMount(async () => {
		if (!$isAuthenticated) {
			goto('/login');
			return;
		}

		try {
			const result = await api.listApps();
			apps = result.data;
		} catch (e) {
			console.error('Failed to load apps', e);
		} finally {
			loading = false;
		}
	});

	async function handleCreate() {
		error = '';
		try {
			const result = await api.createApp(newName, newSlug);
			apps = [result.app, ...apps];
			showCreateForm = false;
			newName = '';
			newSlug = '';
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to create app';
		}
	}

	async function handleDelete(id: string) {
		if (!confirm('Delete this app? This action cannot be undone.')) return;
		try {
			await api.deleteApp(id);
			apps = apps.filter((a) => a.id !== id);
		} catch (e) {
			console.error('Failed to delete app', e);
		}
	}
</script>

<div class="max-w-6xl mx-auto">
	<div class="flex justify-between items-center mb-8">
		<h1 class="text-3xl font-bold">Applications</h1>
		<button
			onclick={() => (showCreateForm = !showCreateForm)}
			class="bg-nexbic-600 text-white px-4 py-2 rounded hover:bg-nexbic-700 transition-colors"
		>
			Create App
		</button>
	</div>

	{#if showCreateForm}
		<div class="bg-white rounded-lg shadow p-6 mb-6">
			<h2 class="text-xl font-semibold mb-4">New Application</h2>
			<form onsubmit={handleCreate} class="space-y-4">
				{#if error}
					<div class="bg-red-50 text-red-700 p-3 rounded text-sm">{error}</div>
				{/if}

				<div>
					<label for="app-name" class="block text-sm font-medium text-gray-700 mb-1">Name</label>
					<input id="app-name" type="text" bind:value={newName} required class="w-full px-3 py-2 border rounded" />
				</div>

				<div>
					<label for="app-slug" class="block text-sm font-medium text-gray-700 mb-1">Slug</label>
					<input id="app-slug" type="text" bind:value={newSlug} required class="w-full px-3 py-2 border rounded font-mono" />
					<p class="text-xs text-gray-500 mt-1">Used in API URL: /api/v1/{newSlug || 'slug'}/*</p>
				</div>

				<div class="flex gap-2">
					<button type="submit" class="bg-nexbic-600 text-white px-4 py-2 rounded hover:bg-nexbic-700">
						Create
					</button>
					<button type="button" onclick={() => (showCreateForm = false)} class="px-4 py-2 rounded border hover:bg-gray-50">
						Cancel
					</button>
				</div>
			</form>
		</div>
	{/if}

	{#if loading}
		<p class="text-gray-500">Loading...</p>
	{:else if apps.length === 0}
		<div class="bg-white rounded-lg shadow p-12 text-center">
			<p class="text-gray-500 text-lg mb-4">No applications yet</p>
			<button
				onclick={() => (showCreateForm = true)}
				class="bg-nexbic-600 text-white px-6 py-3 rounded hover:bg-nexbic-700"
			>
				Create your first app
			</button>
		</div>
	{:else}
		<div class="grid gap-4">
			{#each apps as app}
				<div class="bg-white rounded-lg shadow p-6 flex justify-between items-center">
					<div>
						<a href="/apps/{app.id}" class="text-lg font-semibold text-nexbic-600 hover:underline">
							{app.name}
						</a>
						<p class="text-sm text-gray-500 font-mono">/api/v1/{app.slug}</p>
						<p class="text-xs text-gray-400 mt-1">Created {new Date(app.created_at).toLocaleDateString()}</p>
					</div>
					<div class="flex items-center gap-3">
						<span class="text-xs uppercase px-2 py-1 rounded {app.status === 'active' ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'}">
							{app.status}
						</span>
						<button
							onclick={() => handleDelete(app.id)}
							class="text-red-500 hover:text-red-700 text-sm"
						>
							Delete
						</button>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
