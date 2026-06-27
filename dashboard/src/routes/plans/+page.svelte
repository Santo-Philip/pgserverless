<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import type { Plan } from '$lib/types';
	import Card from '$lib/components/Card.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import LoadingCard from '$lib/components/LoadingCard.svelte';

	let loading = $state(true);
	let plans: Plan[] = $state([]);
	let showCreate = $state(false);
	let newPlan = { name: '', slug: '', max_databases: 1, max_storage_mb: 100, max_connections: 20, max_requests: 10000, max_api_keys: 5, price: 0 };

	onMount(load);

	async function load() {
		loading = true;
		try {
			plans = await api.listPlans();
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	async function createPlan() {
		try {
			await api.createPlan(newPlan as any);
			showCreate = false;
			await load();
		} catch (e) {
			alert('Failed: ' + (e as Error).message);
		}
	}
</script>

<div class="max-w-4xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<h1 class="text-2xl font-bold" style="color: var(--text-primary);">Plans</h1>
		<button onclick={() => showCreate = true} class="btn btn-primary btn-sm">New Plan</button>
	</div>

	{#if loading}
		<LoadingCard />
	{:else if plans.length === 0}
		<Card title="No Plans"><p class="text-sm" style="color: var(--text-secondary);">No plans configured.</p></Card>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
			{#each plans as plan}
				<div class="p-6 rounded-xl border" style="background-color: var(--bg-secondary); border-color: var(--border-primary);">
					<h3 class="text-lg font-bold" style="color: var(--text-primary);">{plan.name}</h3>
					<p class="text-2xl font-bold mt-2" style="color: var(--accent);">${plan.price}<span class="text-sm font-normal" style="color: var(--text-secondary);">/mo</span></p>
					<div class="mt-4 space-y-2 text-sm" style="color: var(--text-secondary);">
						<p>Databases: {plan.max_databases}</p>
						<p>Storage: {plan.max_storage_mb} MB</p>
						<p>Connections: {plan.max_connections}</p>
						<p>Requests: {plan.max_requests.toLocaleString()}/mo</p>
						<p>API Keys: {plan.max_api_keys}</p>
					</div>
					<span class="inline-block mt-4 text-xs px-2 py-1 rounded-full" style={plan.is_active ? 'background-color: #dcfce7; color: #166534;' : 'background-color: #fef2f2; color: #991b1b;'}>
						{plan.is_active ? 'Active' : 'Inactive'}
					</span>
				</div>
			{/each}
		</div>
	{/if}
</div>

{#if showCreate}
	<Modal title="Create Plan" onclose={() => showCreate = false}>
		<div class="space-y-4">
			<div class="grid grid-cols-2 gap-4">
				<div>
					<label class="block text-sm font-medium mb-1">Name</label>
					<input type="text" bind:value={newPlan.name} class="w-full px-3 py-2 rounded-lg border text-sm"
						style="background-color: var(--bg-primary); border-color: var(--border-primary); color: var(--text-primary);" />
				</div>
				<div>
					<label class="block text-sm font-medium mb-1">Slug</label>
					<input type="text" bind:value={newPlan.slug} class="w-full px-3 py-2 rounded-lg border text-sm"
						style="background-color: var(--bg-primary); border-color: var(--border-primary); color: var(--text-primary);" />
				</div>
				<div>
					<label class="block text-sm font-medium mb-1">Max Databases</label>
					<input type="number" bind:value={newPlan.max_databases} class="w-full px-3 py-2 rounded-lg border text-sm"
						style="background-color: var(--bg-primary); border-color: var(--border-primary); color: var(--text-primary);" />
				</div>
				<div>
					<label class="block text-sm font-medium mb-1">Max Storage (MB)</label>
					<input type="number" bind:value={newPlan.max_storage_mb} class="w-full px-3 py-2 rounded-lg border text-sm"
						style="background-color: var(--bg-primary); border-color: var(--border-primary); color: var(--text-primary);" />
				</div>
				<div>
					<label class="block text-sm font-medium mb-1">Max Connections</label>
					<input type="number" bind:value={newPlan.max_connections} class="w-full px-3 py-2 rounded-lg border text-sm"
						style="background-color: var(--bg-primary); border-color: var(--border-primary); color: var(--text-primary);" />
				</div>
				<div>
					<label class="block text-sm font-medium mb-1">Max Requests</label>
					<input type="number" bind:value={newPlan.max_requests} class="w-full px-3 py-2 rounded-lg border text-sm"
						style="background-color: var(--bg-primary); border-color: var(--border-primary); color: var(--text-primary);" />
				</div>
				<div>
					<label class="block text-sm font-medium mb-1">Max API Keys</label>
					<input type="number" bind:value={newPlan.max_api_keys} class="w-full px-3 py-2 rounded-lg border text-sm"
						style="background-color: var(--bg-primary); border-color: var(--border-primary); color: var(--text-primary);" />
				</div>
				<div>
					<label class="block text-sm font-medium mb-1">Price ($)</label>
					<input type="number" bind:value={newPlan.price} step="0.01" class="w-full px-3 py-2 rounded-lg border text-sm"
						style="background-color: var(--bg-primary); border-color: var(--border-primary); color: var(--text-primary);" />
				</div>
			</div>
			<div class="flex justify-end gap-2 pt-2">
				<button onclick={() => showCreate = false} class="btn btn-ghost btn-sm">Cancel</button>
				<button onclick={createPlan} class="btn btn-primary btn-sm">Create</button>
			</div>
		</div>
	</Modal>
{/if}
