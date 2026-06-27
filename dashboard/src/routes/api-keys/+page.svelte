<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import type { APIKey } from '$lib/types';
	import Card from '$lib/components/Card.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import LoadingCard from '$lib/components/LoadingCard.svelte';

	let loading = $state(true);
	let keys: APIKey[] = $state([]);
	let showCreate = $state(false);
	let newName = $state('');
	let newType = $state('service');
	let newKeyResult = $state<APIKey | null>(null);

	onMount(load);

	async function load() {
		loading = true;
		try {
			keys = await api.listAPIKeys();
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	async function createKey() {
		try {
			newKeyResult = await api.createAPIKey(newName, newType);
		} catch (e) {
			alert('Failed: ' + (e as Error).message);
		}
	}

	async function revokeKey(id: string) {
		if (!confirm('Revoke this API key? This cannot be undone.')) return;
		try {
			await api.revokeAPIKey(id);
			await load();
		} catch (e) {
			alert('Failed: ' + (e as Error).message);
		}
	}
</script>

<div class="max-w-4xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<h1 class="text-2xl font-bold" style="color: var(--text-primary);">API Keys</h1>
		<button onclick={() => { showCreate = true; newKeyResult = null; }} class="btn btn-primary btn-sm">New Key</button>
	</div>

	{#if loading}
		<LoadingCard />
	{:else if keys.length === 0}
		<Card title="No API Keys"><p class="text-sm" style="color: var(--text-secondary);">No API keys configured.</p></Card>
	{:else}
		<div class="grid gap-3">
			{#each keys as key}
				<div class="p-4 rounded-xl border" style="background-color: var(--bg-secondary); border-color: var(--border-primary);">
					<div class="flex items-center justify-between">
						<div>
							<div class="font-semibold text-sm" style="color: var(--text-primary);">{key.name}</div>
							<div class="text-xs mt-0.5" style="color: var(--text-secondary);">
								Type: {key.key_type} | Prefix: {key.key_prefix} | Rate: {key.rate_limit}/min
								{#if key.project_id} | Project: {key.project_id}{/if}
							</div>
						</div>
						<div class="flex items-center gap-2">
							<span class="text-xs px-2 py-1 rounded-full" style={key.is_active ? 'background-color: #dcfce7; color: #166534;' : 'background-color: #fef2f2; color: #991b1b;'}>
								{key.is_active ? 'Active' : 'Revoked'}
							</span>
							{#if key.is_active}
								<button onclick={() => revokeKey(key.id)} class="text-xs text-red-500">Revoke</button>
							{/if}
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

{#if showCreate}
	<Modal title="Create API Key" onclose={() => { showCreate = false; newKeyResult = null; }}>
		{#if newKeyResult}
			<div class="space-y-4">
				<div class="p-4 rounded-lg" style="background-color: #fef3c7;">
					<p class="text-sm font-medium">Save this key - it will only be shown once.</p>
					<p class="text-sm font-mono mt-2 break-all" style="color: var(--text-primary);">{newKeyResult.raw_key}</p>
				</div>
				<button onclick={() => { showCreate = false; newKeyResult = null; load(); }} class="btn btn-primary btn-sm w-full">Done</button>
			</div>
		{:else}
			<div class="space-y-4">
			<div>
				<label for="newKeyName" class="block text-sm font-medium mb-1" style="color: var(--text-primary);">Name</label>
				<input id="newKeyName" type="text" class="w-full px-3 py-2 rounded-lg border text-sm"
					style="background-color: var(--bg-primary); border-color: var(--border-primary); color: var(--text-primary);"
					bind:value={newName} placeholder="My Key" />
			</div>
			<div>
				<label for="newKeyType" class="block text-sm font-medium mb-1" style="color: var(--text-primary);">Type</label>
				<select id="newKeyType" class="w-full px-3 py-2 rounded-lg border text-sm"
					style="background-color: var(--bg-primary); border-color: var(--border-primary); color: var(--text-primary);"
					bind:value={newType}>
						<option value="system">System</option>
						<option value="service">Service</option>
						<option value="project">Project</option>
					</select>
				</div>
				<div class="flex justify-end gap-2 pt-2">
					<button onclick={() => showCreate = false} class="btn btn-ghost btn-sm">Cancel</button>
					<button onclick={createKey} class="btn btn-primary btn-sm" disabled={!newName}>Create</button>
				</div>
			</div>
		{/if}
	</Modal>
{/if}
