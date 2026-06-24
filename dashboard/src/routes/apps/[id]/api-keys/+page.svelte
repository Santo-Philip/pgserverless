<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';
	import Card from '$lib/components/Card.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import CopyButton from '$lib/components/CopyButton.svelte';
	import type { APIKey } from '$lib/types';

	let keys = $state<APIKey[]>([]);
	let loading = $state(true);
	let showNewKey = $state(false);
	let newKeyName = $state('');
	let newKeyType = $state('publishable');
	let newKeyResult = $state<APIKey | null>(null);

	onMount(async () => {
		try {
			keys = await api.listAPIKeys($page.params.id!);
		} catch {}
		loading = false;
	});

	async function handleGenerate() {
		try {
			const key = await api.createAPIKey($page.params.id!, newKeyName, newKeyType);
			keys = [key, ...keys];
			newKeyResult = key;
			newKeyName = '';
			showNewKey = false;
		} catch {}
	}

	async function handleDeactivate(keyId: string) {
		try {
			await api.deactivateAPIKey($page.params.id!, keyId);
			keys = keys.map(k => k.id === keyId ? { ...k, is_active: false } : k);
		} catch {}
	}
</script>

<div class="space-y-4">
	<Card title="API Keys" description="Manage API keys for this application">
		<div class="flex justify-end mb-4 -mt-1">
			<button onclick={() => showNewKey = true} class="btn btn-primary btn-sm">+ Generate Key</button>
		</div>
		{#if showNewKey}
			<div class="mb-4 p-4 rounded-lg" style="background-color: var(--bg-tertiary);">
				<div class="space-y-3">
					<div>
						<label class="block text-xs font-medium mb-1" style="color: var(--text-secondary)">Key Name</label>
						<input type="text" bind:value={newKeyName} class="input" placeholder="e.g. Production" />
					</div>
					<div>
						<label class="block text-xs font-medium mb-1" style="color: var(--text-secondary)">Key Type</label>
						<select bind:value={newKeyType} class="input">
							<option value="publishable">Publishable</option>
							<option value="secret">Secret</option>
							<option value="service">Service</option>
							<option value="admin">Admin</option>
						</select>
					</div>
					<button onclick={handleGenerate} class="btn btn-primary btn-sm">Generate</button>
				</div>
			</div>
		{/if}

		{#if newKeyResult?.raw_key}
			<div class="mb-4 p-4 rounded-lg" style="background-color: rgba(245,158,11,0.1); border: 1px solid rgba(245,158,11,0.3);">
				<p class="text-sm font-medium mb-2" style="color: var(--warning)}">Store this key securely — it will not be shown again.</p>
				<div class="flex items-center gap-2">
					<code class="flex-1 p-2 rounded text-xs font-mono" style="background-color: var(--bg-primary); color: var(--text-primary)">{newKeyResult.raw_key}</code>
					<CopyButton text={newKeyResult.raw_key} />
				</div>
			</div>
		{/if}

		{#if loading}
			<Skeleton rows={3} />
		{:else if keys.length === 0}
			<EmptyState title="No API keys" description="Generate your first API key to get started." />
		{:else}
			<div class="table-wrap overflow-x-auto -mx-5 -mb-5">
				<table class="w-full">
					<thead>
						<tr>
							<th>Name</th>
							<th>Type</th>
							<th>Prefix</th>
							<th>Status</th>
							<th>Created</th>
							<th class="text-right">Actions</th>
						</tr>
					</thead>
					<tbody>
						{#each keys as key}
							<tr>
								<td class="font-medium">{key.name}</td>
								<td><Badge status={key.key_type} /></td>
								<td><span class="font-mono text-xs" style="color: var(--text-secondary)">{key.key_prefix}...</span></td>
								<td><Badge status={key.is_active ? 'active' : 'inactive'} /></td>
								<td class="text-xs" style="color: var(--text-tertiary)">{new Date(key.created_at).toLocaleDateString()}</td>
								<td class="text-right">
									{#if key.is_active}
										<button onclick={() => handleDeactivate(key.id)} class="btn btn-ghost btn-sm" style="color: var(--danger)">Deactivate</button>
									{/if}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</Card>
</div>
