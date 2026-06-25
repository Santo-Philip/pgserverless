<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import Card from '$lib/components/Card.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import CopyButton from '$lib/components/CopyButton.svelte';
	import type { App, APIKey } from '$lib/types';

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

	let createdResult = $state<{
		app: App;
		admin_key: APIKey;
		service_key: APIKey;
		jwt_secret: string;
		connection_uri: string;
	} | null>(null);

	onMount(async () => {
		try {
			apps = await api.listApps();
		} catch {}
		loading = false;
	});

	async function handleCreate() {
		createError = '';
		createLoading = true;
		try {
			const result = await api.createApp(newName, newSlug, newDesc || undefined);
			apps = [result.app, ...apps];
			createdResult = result;
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

	function dismissCreated() {
		createdResult = null;
	}

	async function handleDelete() {
		if (!deleting) return;
		deletingLoading = true;
		try {
			await api.deleteApp(deleting.id);
			apps = apps.filter(a => a.id !== deleting!.id);
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

	{#if createdResult}
		<div class="mb-6 p-5 rounded-xl border" style="background-color: rgba(34,197,94,0.05); border-color: rgba(34,197,94,0.2);">
			<div class="flex items-start justify-between mb-4">
				<div>
					<h3 class="text-lg font-semibold" style="color: var(--success)">Application Created</h3>
					<p class="text-sm mt-1" style="color: var(--text-secondary)">Database schema and API keys have been generated.</p>
				</div>
				<button onclick={dismissCreated} class="btn btn-ghost btn-sm">Dismiss</button>
			</div>

			<div class="space-y-4">
				<div class="p-3 rounded-lg" style="background-color: var(--bg-tertiary);">
					<p class="text-xs font-medium mb-1" style="color: var(--text-tertiary)">Database Schema</p>
					<p class="text-sm font-mono">{createdResult.app.schema_name}</p>
				</div>

				<div class="p-3 rounded-lg" style="background-color: var(--bg-tertiary);">
					<p class="text-xs font-medium mb-1" style="color: var(--text-tertiary)">API Endpoint</p>
					<div class="flex items-center justify-between">
						<p class="text-sm font-mono">/api/v1/{createdResult.app.slug}</p>
						<CopyButton text={`/api/v1/${createdResult.app.slug}`} />
					</div>
				</div>

				<div class="p-3 rounded-lg" style="background-color: rgba(245,158,11,0.08); border: 1px solid rgba(245,158,11,0.2);">
					<p class="text-xs font-medium mb-2" style="color: var(--warning)">Admin API Key — store securely, shown only once</p>
					<div class="flex items-center gap-2">
						<code class="flex-1 p-2 rounded text-xs font-mono break-all" style="background-color: var(--bg-primary); color: var(--text-primary)">{createdResult.admin_key.raw_key}</code>
						<CopyButton text={createdResult.admin_key.raw_key ?? ''} />
					</div>
					<p class="text-xs mt-1.5" style="color: var(--text-tertiary)">Key prefix: {createdResult.admin_key.key_prefix}</p>
				</div>

				<div class="p-3 rounded-lg" style="background-color: var(--bg-tertiary);">
					<p class="text-xs font-medium mb-2" style="color: var(--text-tertiary)">Service API Key</p>
					<div class="flex items-center gap-2">
						<code class="flex-1 p-2 rounded text-xs font-mono break-all" style="background-color: var(--bg-primary); color: var(--text-primary)">{createdResult.service_key.raw_key}</code>
						<CopyButton text={createdResult.service_key.raw_key ?? ''} />
					</div>
					<p class="text-xs mt-1.5" style="color: var(--text-tertiary)">Key prefix: {createdResult.service_key.key_prefix}</p>
				</div>
			</div>
		</div>
	{/if}

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
