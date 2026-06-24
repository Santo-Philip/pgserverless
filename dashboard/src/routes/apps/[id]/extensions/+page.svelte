<script lang="ts">
	import Card from '$lib/components/Card.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import Alert from '$lib/components/Alert.svelte';

	let extensions = $state([
		{ name: 'pgcrypto', description: 'Cryptographic functions for PostgreSQL', installed: true },
		{ name: 'uuid-ossp', description: 'UUID generation functions', installed: true },
		{ name: 'pgvector', description: 'Vector similarity search for embeddings', installed: false },
		{ name: 'citext', description: 'Case-insensitive character string type', installed: false },
		{ name: 'pg_trgm', description: 'Trigram text search', installed: false },
		{ name: 'hstore', description: 'Key-value store data type', installed: false },
		{ name: 'ltree', description: 'Hierarchical tree data type', installed: false },
		{ name: 'pg_stat_statements', description: 'SQL statement execution statistics', installed: false },
		{ name: 'postgis', description: 'Geographic Information System (GIS) support', installed: false },
	]);

	let message = $state('');

	async function toggleExtension(name: string, install: boolean) {
		message = '';
		try {
			const { api } = await import('$lib/api/client');
			await api.post(`/api/v1/platform/apps/${window.location.pathname.split('/')[2]}/extensions/toggle`, { name, install });
			extensions = extensions.map(e => e.name === name ? { ...e, installed: install } : e);
			message = `Extension "${name}" ${install ? 'enabled' : 'disabled'} successfully.`;
		} catch {
			message = `Failed to ${install ? 'enable' : 'disable'} extension "${name}".`;
		}
	}
</script>

<div class="space-y-4">
	<Alert message={message} type={message.includes('successfully') ? 'success' : message ? 'error' : undefined} />

	<Card title="PostgreSQL Extensions" description="Enable additional functionality for your application's database">
		{#if extensions.length === 0}
			<p class="text-sm" style="color: var(--text-secondary)">No extensions available.</p>
		{:else}
			<div class="space-y-1 -mx-5 -mb-5">
				{#each extensions as ext}
					<div class="flex items-center justify-between px-5 py-4 border-t" style="border-color: var(--border-primary);">
						<div class="flex-1">
							<div class="flex items-center gap-2">
								<span class="text-sm font-medium">{ext.name}</span>
								{#if ext.installed}
									<span class="badge text-xs" style="background-color: rgba(34,197,94,0.1); color: var(--success)">Enabled</span>
								{/if}
							</div>
							<p class="text-xs mt-0.5" style="color: var(--text-tertiary)">{ext.description}</p>
						</div>
						<button
							onclick={() => toggleExtension(ext.name, !ext.installed)}
							class="btn btn-sm {ext.installed ? 'btn-ghost' : 'btn-primary'}"
							style={ext.installed ? 'color: var(--danger)' : ''}
						>
							{ext.installed ? 'Disable' : 'Enable'}
						</button>
					</div>
				{/each}
			</div>
		{/if}
	</Card>
</div>
