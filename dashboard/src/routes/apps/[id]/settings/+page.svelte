<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';
	import Card from '$lib/components/Card.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import Alert from '$lib/components/Alert.svelte';
	import type { App } from '$lib/types';

	let app = $state<App | null>(null);
	let loading = $state(true);
	let saving = $state(false);
	let message = $state('');
	let showDelete = $state(false);
	let deleteLoading = $state(false);

	onMount(async () => {
		try {
			app = await api.getApp($page.params.id!);
		} catch {}
		loading = false;
	});

	async function handleSave() {
		if (!app) return;
		saving = true;
		message = '';
		try {
			await api.patch(`/api/v1/platform/apps/${$page.params.id}`, {
				name: app.name,
				description: app.description,
			});
			message = 'Settings saved.';
		} catch { message = 'Failed to save settings.'; }
		saving = false;
	}

	async function handleDelete() {
		deleteLoading = true;
		try {
			await api.deleteApp($page.params.id!);
			window.location.href = '/apps';
		} catch { message = 'Failed to delete app.'; }
		deleteLoading = false;
		showDelete = false;
	}
</script>

<div class="space-y-6">
	<Alert message={message} type={message.includes('saved') || !message ? undefined : 'error'} />

	{#if loading}
		<div class="card p-12"><div class="skeleton h-8 w-full mb-4"></div><div class="skeleton h-8 w-full"></div></div>
	{:else if app}
		<Card title="General Settings">
			<div class="space-y-4">
				<div>
					<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Application Name</label>
					<input type="text" bind:value={app.name} class="input" />
				</div>
				<div>
					<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Slug</label>
					<input type="text" value={app.slug} disabled class="input !opacity-50" />
					<p class="text-xs mt-1" style="color: var(--text-tertiary)">Slug cannot be changed after creation.</p>
				</div>
				<div>
					<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Description</label>
					<textarea bind:value={app.description} class="input" rows="2"></textarea>
				</div>
				<button onclick={handleSave} disabled={saving} class="btn btn-primary">{saving ? 'Saving...' : 'Save Changes'}</button>
			</div>
		</Card>

		<Card title="Danger Zone">
			<p class="text-sm mb-4" style="color: var(--text-secondary)">Permanently delete this application and all of its data. This action cannot be undone.</p>
			<button onclick={() => showDelete = true} class="btn btn-danger">Delete Application</button>
		</Card>
	{/if}
</div>

<ConfirmDialog
	open={showDelete}
	title="Delete Application"
	description={`This will permanently delete "${app?.name}" and all its data, schemas, and API keys.`}
	confirmLabel="Delete Forever"
	variant="danger"
	loading={deleteLoading}
	onconfirm={handleDelete}
	oncancel={() => showDelete = false}
/>
