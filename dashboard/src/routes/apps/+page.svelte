<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import { goto } from '$app/navigation';
	import Card from '$lib/components/Card.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';

	let apps = $state<App[]>([]);
	let loading = $state(true);
	let showCreate = $state(false);
	let deleting = $state<{ id: string; name: string } | null>(null);
	let deletingLoading = $state(false);

	let newName = $state('');
	let newSlug = $state('');
	let newDesc = $state('');
	let createError = $state('');
	let createLoading = $state(false);

	onMount(async () => {
		try {
			const result = await api.listApps();
			apps = result.data;
		} catch {}
		loading = false;
	});

	async function handleCreate() {
		createError = '';
		createLoading = true;
		try {
			const result = await api.createApp(newName, newSlug, newDesc || undefined);
			apps = [result.app, ...apps];
			showCreate = false;
			newName = '';
			newSlug = '';
			newDesc = '';
		} catch (e) {
			createError = e instanceof Error ? e.message : 'Failed to create app';
		} finally {
			createLoading = false;
		}
	}

	async function handleDelete() {
		if (!deleting) return;
		deletingLoading = true;
		try {
			await api.deleteApp(deleting.id);
			apps = apps.filter(a => a.id !== deleting.id);
		} catch {}
		deletingLoading = false;
		deleting = null;
	}
</script>

<div class="max-w-7xl mx-auto">
	<div class="flex items-center justify-between mb-8">
		<div>
			<h1 class="text-2xl font-bold">Applications</h1>
			<p class="text-sm mt-1" style="color: var(--text-secondary)">Manage your PostgreSQL-backed applications</p>
		</div>
		<button onclick={() => showCreate = true} class="btn btn-primary">
			+ Create Application
		</button>
	</div>

	<Modal title="Create Application" open={showCreate} onclose={() => showCreate = false}>
		<form onsubmit={handleCreate} class="space-y-4">
			{#if createError}
				<div class="px-4 py-3 rounded-lg text-sm" style="background-color: rgba(239,68,68,0.1); color: var(--danger)">{createError}</div>
			{/if}

			<div>
				<label class="block text-sm font-medium mb-1.5" style="color: var(--text-secondary)">Application Name</label>
				<input type="text" bind:value={newName} required class="input" placeholder="My App" />
			</div>

			<div>
				<label class="block text-sm font-medium mb-1.5" style="color: var(--text-secondary)">Slug</label>
				<input type="text" bind:value={newSlug} required class="input font-mono" placeholder="my-app" />
				<p class="text-xs mt-1" style="color: var(--text-tertiary)">Used in API URL: /api/v1/<span class="font-mono">{newSlug || 'slug'}</span></p>
			</div>

			<div>
				<label class="block text-sm font-medium mb-1.5" style="color: var(--text-secondary)">Description</label>
				<textarea bind:value={newDesc} class="input" rows="2" placeholder="Optional description"></textarea>
			</div>

			<div class="flex justify-end gap-3 pt-2">
				<button type="button" onclick={() => showCreate = false} class="btn btn-secondary">Cancel</button>
				<button type="submit" disabled={createLoading} class="btn btn-primary">{createLoading ? 'Creating...' : 'Create'}</button>
			</div>
		</form>
	</Modal>

	<ConfirmDialog
		open={!!deleting}
		title="Delete Application"
		description={`Are you sure you want to delete "${deleting?.name}"? This will remove all data, schemas, and API keys. This action cannot be undone.`}
		confirmLabel="Delete"
		variant="danger"
		loading={deletingLoading}
		onconfirm={handleDelete}
		oncancel={() => deleting = null}
	/>

	{#if loading}
		<div class="card p-0"><Skeleton rows={4} /></div>
	{:else if apps.length === 0}
		<EmptyState icon="▦" title="No applications yet" description="Create your first application to get started.">
			<button onclick={() => showCreate = true} class="btn btn-primary">Create Application</button>
		</EmptyState>
	{:else}
		<div class="card p-0 overflow-hidden">
			<div class="table-wrap overflow-x-auto">
				<table class="w-full">
					<thead>
						<tr>
							<th>Application</th>
							<th>Slug</th>
							<th>Schema</th>
							<th>Status</th>
							<th>Created</th>
							<th class="text-right">Actions</th>
						</tr>
					</thead>
					<tbody>
						{#each apps as app}
							<tr>
								<td>
									<a href="/apps/{app.id}" class="font-medium link">{app.name}</a>
								</td>
								<td><span class="font-mono text-xs" style="color: var(--text-secondary)">{app.slug}</span></td>
								<td><span class="font-mono text-xs" style="color: var(--text-secondary)">{app.schema_name}</span></td>
								<td><Badge status={app.status} /></td>
								<td class="text-xs" style="color: var(--text-tertiary)">{new Date(app.created_at).toLocaleDateString()}</td>
								<td class="text-right">
									<div class="flex items-center justify-end gap-1">
										<a href="/apps/{app.id}" class="btn btn-ghost btn-sm">Open</a>
										<a href="/apps/{app.id}/settings" class="btn btn-ghost btn-sm">Settings</a>
										<button onclick={() => deleting = { id: app.id, name: app.name }} class="btn btn-ghost btn-sm" style="color: var(--danger)">Delete</button>
									</div>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>
