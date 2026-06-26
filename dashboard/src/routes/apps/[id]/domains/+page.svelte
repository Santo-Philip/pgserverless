<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';
	import type { Domain } from '$lib/types';
	import Badge from '$lib/components/Badge.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';

	let domains = $state<Domain[]>([]);
	let loading = $state(true);
	let newDomain = $state('');
	let error = $state('');
	let deleteTarget = $state<string | null>(null);
	let deleting = $state(false);

	onMount(async () => {
		try {
			domains = await api.listDomains($page.params.id!);
		} catch (e) {
			console.error('Failed to load domains', e);
		} finally {
			loading = false;
		}
	});

	async function handleAdd() {
		error = '';
		try {
			const domain = await api.createDomain($page.params.id!, newDomain);
			domains = [domain, ...domains];
			newDomain = '';
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to add domain';
		}
	}

	async function handleVerify(domainId: string) {
		try {
			await api.verifyDomain($page.params.id!, domainId);
			domains = domains.map((d) =>
				d.id === domainId ? { ...d, verified: true, status: 'active' as const } : d
			);
		} catch (e) {
			console.error('Failed to verify domain', e);
		}
	}

	async function handleDelete() {
		if (!deleteTarget) return;
		deleting = true;
		try {
			await api.deleteDomain($page.params.id!, deleteTarget);
			domains = domains.filter((d) => d.id !== deleteTarget);
		} catch (e) {
			console.error('Failed to delete domain', e);
		}
		deleting = false;
		deleteTarget = null;
	}
</script>

<div class="max-w-4xl mx-auto">
	<a href="/apps/{$page.params.id}" class="link text-sm mb-4 inline-block" style="color: var(--accent)">&larr; Back to App</a>
	<h1 class="text-2xl font-bold mb-6">Custom Domains</h1>

	<div class="card p-5 mb-6">
		<h2 class="text-sm font-semibold mb-4" style="color: var(--text-primary)">Add Domain</h2>
		<form onsubmit={handleAdd} class="space-y-4">
			{#if error}
				<div class="px-4 py-3 rounded-lg text-sm" style="background-color: rgba(239,68,68,0.1); color: var(--danger)">{error}</div>
			{/if}

			<div>
				<label for="domain" class="block text-sm font-medium mb-1.5" style="color: var(--text-secondary)">Domain</label>
				<input
					id="domain"
					type="text"
					bind:value={newDomain}
					placeholder="api.example.com"
					required
					class="input font-mono"
				/>
			</div>

			<button type="submit" class="btn btn-primary">Add Domain</button>
		</form>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-12">
			<p style="color: var(--text-tertiary)">Loading...</p>
		</div>
	{:else if domains.length === 0}
		<EmptyState title="No custom domains" description="Add your first custom domain above." />
	{:else}
		<div class="card overflow-hidden">
			<div class="table-wrap">
				<table class="w-full">
					<thead>
						<tr>
							<th>Domain</th>
							<th>Status</th>
							<th>Verified</th>
							<th class="text-right">Actions</th>
						</tr>
					</thead>
					<tbody>
						{#each domains as domain}
							<tr>
								<td class="font-mono">{domain.domain}</td>
								<td><Badge status={domain.status} /></td>
								<td>
									<span class="text-sm" style="color: {domain.verified ? 'var(--success)' : 'var(--text-tertiary)'}">
										{domain.verified ? 'Yes' : 'No'}
									</span>
								</td>
								<td class="text-right space-x-2">
									{#if !domain.verified}
										<button onclick={() => handleVerify(domain.id)} class="btn btn-ghost btn-sm text-xs">
											Verify
										</button>
									{/if}
									<button onclick={() => deleteTarget = domain.id} class="btn btn-ghost btn-sm text-xs" style="color: var(--danger)">
										Delete
									</button>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>

<ConfirmDialog
	open={deleteTarget !== null}
	title="Delete Domain"
	description="Are you sure you want to delete this domain? This action cannot be undone."
	confirmLabel="Delete"
	variant="danger"
	loading={deleting}
	onconfirm={handleDelete}
	oncancel={() => deleteTarget = null}
/>
