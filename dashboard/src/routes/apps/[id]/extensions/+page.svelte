<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import Card from '$lib/components/Card.svelte';
	import Alert from '$lib/components/Alert.svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';

	let appId = $derived($page.params.id ?? '');

	let extensions = $state<{name: string; version: string; description: string; installed: boolean}[]>([]);
	let loading = $state(true);
	let message = $state('');
	let messageType = $state<'success' | 'error' | undefined>(undefined);
	let toggling = $state<string | null>(null);

	onMount(async () => {
		try {
			const result = await api.listExtensions(appId);
			extensions = result as typeof extensions;
		} catch {
			message = 'Failed to load extensions.';
			messageType = 'error';
		}
		loading = false;
	});

	async function toggleExtension(name: string, install: boolean) {
		message = '';
		toggling = name;
		try {
			await api.toggleExtension(appId, name, install);
			extensions = extensions.map(e => e.name === name ? { ...e, installed: install } : e);
			message = `Extension "${name}" ${install ? 'enabled' : 'disabled'} successfully.`;
			messageType = 'success';
		} catch {
			message = `Failed to ${install ? 'enable' : 'disable'} extension "${name}".`;
			messageType = 'error';
		}
		toggling = null;
	}
</script>

<div class="space-y-4">
	<Alert {message} type={messageType} />

	<Card title="PostgreSQL Extensions" description="Enable additional functionality for your application's database">
		{#if loading}
			<Skeleton rows={5} />
		{:else if extensions.length === 0}
			<EmptyState icon="⚡" title="No extensions" description="No PostgreSQL extensions available." />
		{:else}
			<div class="space-y-1 -mx-5 -mb-5">
				{#each extensions as ext}
					<div class="flex items-center justify-between px-5 py-4 border-t" style="border-color: var(--border-primary);">
						<div class="flex-1 min-w-0">
							<div class="flex items-center gap-2">
								<span class="text-sm font-medium">{ext.name}</span>
								{#if ext.version}
									<span class="text-xs" style="color: var(--text-tertiary)">{ext.version}</span>
								{/if}
								{#if ext.installed}
									<span class="badge text-xs" style="background-color: rgba(34,197,94,0.1); color: var(--success)">Enabled</span>
								{/if}
							</div>
							<p class="text-xs mt-0.5 truncate" style="color: var(--text-tertiary)">{ext.description}</p>
						</div>
						<button
							onclick={() => toggleExtension(ext.name, !ext.installed)}
							disabled={toggling === ext.name}
							class="btn btn-sm {ext.installed ? 'btn-ghost' : 'btn-primary'}"
							style={ext.installed ? 'color: var(--danger)' : ''}
						>
							{#if toggling === ext.name}
								<span class="inline-block animate-spin">⟳</span>
							{:else}
								{ext.installed ? 'Disable' : 'Enable'}
							{/if}
						</button>
					</div>
				{/each}
			</div>
		{/if}
	</Card>
</div>
