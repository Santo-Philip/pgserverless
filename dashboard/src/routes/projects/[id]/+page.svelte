<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import type { Project, Database, APIKey, TableInfo } from '$lib/types';
	import Card from '$lib/components/Card.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import LoadingCard from '$lib/components/LoadingCard.svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';

	let loading = $state(true);
	let project: Project | null = $state(null);
	let databases: Database[] = $state([]);
	let apiKeys: APIKey[] = $state([]);
	let activeTab = $state('databases');

	// Database modal
	let showCreateDB = $state(false);
	let newDBName = $state('');
	let deleteDBTarget = $state<Database | null>(null);

	// API key modal
	let showCreateKey = $state(false);
	let newKeyName = $state('');
	let newKeyType = $state('project');
	let deleteKeyTarget = $state<APIKey | null>(null);

	async function load() {
		loading = true;
		try {
			const id = $page.params.id;
			project = await api.getProject(id);
			[databases, apiKeys] = await Promise.all([
				api.listDatabases(id),
				api.listProjectAPIKeys(id),
			]);
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	onMount(load);

	async function createDatabase() {
		try {
			await api.createDatabase($page.params.id, newDBName);
			showCreateDB = false;
			newDBName = '';
			await load();
		} catch (e) {
			alert('Failed: ' + (e as Error).message);
		}
	}

	async function deleteDatabase() {
		if (!deleteDBTarget) return;
		try {
			await api.deleteDatabase(deleteDBTarget.id);
			deleteDBTarget = null;
			await load();
		} catch (e) {
			alert('Failed: ' + (e as Error).message);
		}
	}

	async function createKey() {
		try {
			await api.createAPIKey(newKeyName, newKeyType, $page.params.id);
			showCreateKey = false;
			newKeyName = '';
			await load();
		} catch (e) {
			alert('Failed: ' + (e as Error).message);
		}
	}

	async function deleteKey() {
		if (!deleteKeyTarget) return;
		try {
			await api.revokeAPIKey(deleteKeyTarget.id);
			deleteKeyTarget = null;
			await load();
		} catch (e) {
			alert('Failed: ' + (e as Error).message);
		}
	}
</script>

{#if loading}
	<div class="max-w-4xl mx-auto"><LoadingCard /></div>
{:else if project}
	<div class="max-w-4xl mx-auto">
		<div class="mb-6">
			<h1 class="text-2xl font-bold" style="color: var(--text-primary);">{project.name}</h1>
			<p class="text-sm mt-1" style="color: var(--text-secondary);">{project.slug} — {project.status}</p>
		</div>

		<div class="flex gap-2 mb-6 border-b" style="border-color: var(--border-primary);">
			<button onclick={() => activeTab = 'databases'} class="px-4 py-2 text-sm font-medium border-b-2 transition-colors"
				style={activeTab === 'databases' ? 'border-color: var(--accent); color: var(--accent);' : 'border-color: transparent; color: var(--text-secondary);'}>
				Databases
			</button>
			<button onclick={() => activeTab = 'api-keys'} class="px-4 py-2 text-sm font-medium border-b-2 transition-colors"
				style={activeTab === 'api-keys' ? 'border-color: var(--accent); color: var(--accent);' : 'border-color: transparent; color: var(--text-secondary);'}>
				API Keys
			</button>
		</div>

		{#if activeTab === 'databases'}
			<div>
				<div class="flex items-center justify-between mb-4">
					<h2 class="text-lg font-semibold" style="color: var(--text-primary);">Databases</h2>
					<button onclick={() => showCreateDB = true} class="btn btn-primary btn-sm">New Database</button>
				</div>

				{#if databases.length === 0}
					<Card title="No Databases"><p class="text-sm" style="color: var(--text-secondary);">Create a database to get started.</p></Card>
				{:else}
					<div class="grid gap-3">
						{#each databases as db}
							<div class="p-4 rounded-xl border" style="background-color: var(--bg-secondary); border-color: var(--border-primary);">
								<div class="flex items-center justify-between">
									<div>
										<div class="font-semibold text-sm" style="color: var(--text-primary);">{db.name}</div>
										<div class="text-xs mt-0.5" style="color: var(--text-secondary);">Schema: {db.schema_name} | User: {db.db_user}</div>
										<div class="text-xs" style="color: var(--text-secondary);">Status: {db.status} | Size: {(db.size_bytes / 1024).toFixed(1)} KB</div>
									</div>
									<div class="flex gap-2">
										<button onclick={() => goto(`/databases/${db.id}`)} class="btn btn-ghost btn-xs">Manage</button>
										<button onclick={() => deleteDBTarget = db} class="text-xs text-red-500">Delete</button>
									</div>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		{:else if activeTab === 'api-keys'}
			<div>
				<div class="flex items-center justify-between mb-4">
					<h2 class="text-lg font-semibold" style="color: var(--text-primary);">API Keys</h2>
					<button onclick={() => showCreateKey = true} class="btn btn-primary btn-sm">New Key</button>
				</div>

				{#if apiKeys.length === 0}
					<Card title="No API Keys"><p class="text-sm" style="color: var(--text-secondary);">No API keys for this project.</p></Card>
				{:else}
					<div class="grid gap-3">
						{#each apiKeys as key}
							<div class="p-4 rounded-xl border" style="background-color: var(--bg-secondary); border-color: var(--border-primary);">
								<div class="flex items-center justify-between">
									<div>
										<div class="font-semibold text-sm" style="color: var(--text-primary);">{key.name}</div>
										<div class="text-xs mt-0.5" style="color: var(--text-secondary);">
											Type: {key.key_type} | Prefix: {key.key_prefix} | Scopes: {key.scopes?.join(', ') || 'none'}
										</div>
									</div>
									<button onclick={() => deleteKeyTarget = key} class="text-xs text-red-500">Revoke</button>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		{/if}
	</div>

	<!-- Create Database Modal -->
	{#if showCreateDB}
		<Modal title="Create Database" onclose={() => showCreateDB = false}>
			<div class="space-y-4">
				<div>
					<label class="block text-sm font-medium mb-1" style="color: var(--text-primary);">Database Name</label>
					<input type="text" class="w-full px-3 py-2 rounded-lg border text-sm"
						style="background-color: var(--bg-primary); border-color: var(--border-primary); color: var(--text-primary);"
						bind:value={newDBName} placeholder="my-database" />
				</div>
				<div class="flex justify-end gap-2 pt-2">
					<button onclick={() => showCreateDB = false} class="btn btn-ghost btn-sm">Cancel</button>
					<button onclick={createDatabase} class="btn btn-primary btn-sm" disabled={!newDBName}>Create</button>
				</div>
			</div>
		</Modal>
	{/if}

	{#if deleteDBTarget}
		<ConfirmDialog title="Delete Database" message={`Delete "${deleteDBTarget.name}"? This is irreversible.`}
			onconfirm={deleteDatabase} oncancel={() => deleteDBTarget = null} />
	{/if}

	<!-- Create API Key Modal -->
	{#if showCreateKey}
		<Modal title="Create API Key" onclose={() => showCreateKey = false}>
			<div class="space-y-4">
				<div>
					<label class="block text-sm font-medium mb-1" style="color: var(--text-primary);">Name</label>
					<input type="text" class="w-full px-3 py-2 rounded-lg border text-sm"
						style="background-color: var(--bg-primary); border-color: var(--border-primary); color: var(--text-primary);"
						bind:value={newKeyName} placeholder="My Key" />
				</div>
				<div>
					<label class="block text-sm font-medium mb-1" style="color: var(--text-primary);">Type</label>
					<select class="w-full px-3 py-2 rounded-lg border text-sm"
						style="background-color: var(--bg-primary); border-color: var(--border-primary); color: var(--text-primary);"
						bind:value={newKeyType}>
						<option value="project">Project</option>
						<option value="service">Service</option>
						<option value="system">System</option>
					</select>
				</div>
				<div class="flex justify-end gap-2 pt-2">
					<button onclick={() => showCreateKey = false} class="btn btn-ghost btn-sm">Cancel</button>
					<button onclick={createKey} class="btn btn-primary btn-sm" disabled={!newKeyName}>Create</button>
				</div>
			</div>
		</Modal>
	{/if}

	{#if deleteKeyTarget}
		<ConfirmDialog title="Revoke API Key" message={`Revoke "${deleteKeyTarget.name}"? This cannot be undone.`}
			onconfirm={deleteKey} oncancel={() => deleteKeyTarget = null} />
	{/if}
{/if}
