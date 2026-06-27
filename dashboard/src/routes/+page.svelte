<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import type { Project, Database, Plan, APIKey } from '$lib/types';
	import StatCard from '$lib/components/StatCard.svelte';
	import Card from '$lib/components/Card.svelte';
	import LoadingCard from '$lib/components/LoadingCard.svelte';
	import { goto } from '$app/navigation';

	let loading = $state(true);
	let projects: Project[] = $state([]);
	let plans: Plan[] = $state([]);

	onMount(async () => {
		try {
			const [p, pl] = await Promise.all([
				api.listProjects(),
				api.listPlans(),
			]);
			projects = p;
			plans = pl;
		} catch (e) {
			console.error('Failed to load dashboard data', e);
		} finally {
			loading = false;
		}
	});
</script>

<div class="max-w-6xl mx-auto">
	<h1 class="text-2xl font-bold mb-6" style="color: var(--text-primary);">Dashboard</h1>

	{#if loading}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
			{#each [1,2,3,4] as _}
				<LoadingCard />
			{/each}
		</div>
	{:else}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
			<StatCard title="Projects" value={projects.length} icon="▦" />
			<StatCard title="Plans" value={plans.length} icon="☆" />
			<StatCard title="API Keys" value="-" icon="🔑" />
			<StatCard title="Databases" value="-" icon="▤" />
		</div>
	{/if}

	<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
		<Card title="Recent Projects">
			{#if projects.length === 0}
				<p class="text-sm" style="color: var(--text-secondary);">No projects yet.</p>
			{:else}
				<div class="space-y-2">
					{#each projects.slice(0, 5) as project}
						<button
							onclick={() => goto(`/projects/${project.id}`)}
							class="w-full text-left p-3 rounded-lg text-sm transition-colors"
							style="background-color: var(--bg-hover);"
						>
							<div class="font-medium" style="color: var(--text-primary);">{project.name}</div>
							<div class="text-xs mt-0.5" style="color: var(--text-secondary);">{project.slug}</div>
						</button>
					{/each}
				</div>
			{/if}
		</Card>

		<Card title="Quick Actions">
			<div class="space-y-2">
				<button onclick={() => goto('/projects')} class="w-full text-left p-3 rounded-lg text-sm transition-colors" style="background-color: var(--bg-hover);">
					<span style="color: var(--text-primary);">Create Project</span>
					<span class="block text-xs mt-0.5" style="color: var(--text-secondary);">Provision a new database project</span>
				</button>
				<button onclick={() => goto('/plans')} class="w-full text-left p-3 rounded-lg text-sm transition-colors" style="background-color: var(--bg-hover);">
					<span style="color: var(--text-primary);">Manage Plans</span>
					<span class="block text-xs mt-0.5" style="color: var(--text-secondary);">Configure resource limits and pricing</span>
				</button>
				<button onclick={() => goto('/api-keys')} class="w-full text-left p-3 rounded-lg text-sm transition-colors" style="background-color: var(--bg-hover);">
					<span style="color: var(--text-primary);">API Keys</span>
					<span class="block text-xs mt-0.5" style="color: var(--text-secondary);">Manage platform API keys</span>
				</button>
			</div>
		</Card>
	</div>
</div>
