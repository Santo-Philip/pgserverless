<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import type { Project } from '$lib/types';
	import Card from '$lib/components/Card.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import LoadingCard from '$lib/components/LoadingCard.svelte';
	import { goto } from '$app/navigation';

	let loading = $state(true);
	let projects: Project[] = $state([]);
	let showCreate = $state(false);
	let newName = $state('');
	let newSlug = $state('');
	let newDesc = $state('');
	let deleteTarget = $state<Project | null>(null);

	async function load() {
		loading = true;
		try {
			projects = await api.listProjects();
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	onMount(load);

	async function createProject() {
		try {
			await api.createProject(newName, newSlug, newDesc);
			showCreate = false;
			newName = '';
			newSlug = '';
			newDesc = '';
			await load();
		} catch (e) {
			alert('Failed to create project: ' + (e as Error).message);
		}
	}

	async function deleteProject() {
		if (!deleteTarget) return;
		try {
			await api.deleteProject(deleteTarget.id);
			deleteTarget = null;
			await load();
		} catch (e) {
			alert('Failed to delete project: ' + (e as Error).message);
		}
	}

	function generateSlug(name: string) {
		newSlug = name.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '');
	}
</script>

<div class="max-w-4xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<h1 class="text-2xl font-bold" style="color: var(--text-primary);">Projects</h1>
		<button onclick={() => showCreate = true} class="btn btn-primary btn-sm">New Project</button>
	</div>

	{#if loading}
		<div class="grid gap-4">
			{#each [1,2,3] as _}
				<LoadingCard />
			{/each}
		</div>
	{:else if projects.length === 0}
		<Card title="No Projects">
			<p class="text-sm" style="color: var(--text-secondary);">Create your first project to get started.</p>
		</Card>
	{:else}
		<div class="grid gap-3">
			{#each projects as project}
				<div
					onclick={() => goto(`/projects/${project.id}`)}
					onkeydown={(e) => { if (e.key === 'Enter') goto(`/projects/${project.id}`); }}
					role="button" tabindex="0"
					class="w-full text-left p-4 rounded-xl border transition-all hover:shadow-md cursor-pointer"
					style="background-color: var(--bg-secondary); border-color: var(--border-primary);"
				>
					<div class="flex items-center justify-between">
						<div>
							<div class="font-semibold" style="color: var(--text-primary);">{project.name}</div>
							<div class="text-xs mt-0.5" style="color: var(--text-secondary);">{project.slug}</div>
						</div>
						<div class="flex items-center gap-2">
							<span class="text-xs px-2 py-1 rounded-full" style="background-color: var(--bg-hover); color: var(--text-secondary);">{project.status}</span>
							<button
								onclick={(e) => { e.stopPropagation(); deleteTarget = project; }}
								class="text-xs text-red-500 hover:text-red-700"
							>Delete</button>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

{#if showCreate}
	<Modal title="Create Project" onclose={() => showCreate = false}>
		<div class="space-y-4">
			<div>
				<label class="block text-sm font-medium mb-1" style="color: var(--text-primary);">Name</label>
				<input
					type="text"
					class="w-full px-3 py-2 rounded-lg border text-sm"
					style="background-color: var(--bg-primary); border-color: var(--border-primary); color: var(--text-primary);"
					bind:value={newName}
					oninput={() => generateSlug(newName)}
					placeholder="My Project"
				/>
			</div>
			<div>
				<label class="block text-sm font-medium mb-1" style="color: var(--text-primary);">Slug</label>
				<input
					type="text"
					class="w-full px-3 py-2 rounded-lg border text-sm"
					style="background-color: var(--bg-primary); border-color: var(--border-primary); color: var(--text-primary);"
					bind:value={newSlug}
					placeholder="my-project"
				/>
			</div>
			<div>
				<label class="block text-sm font-medium mb-1" style="color: var(--text-primary);">Description (optional)</label>
				<textarea
					class="w-full px-3 py-2 rounded-lg border text-sm"
					style="background-color: var(--bg-primary); border-color: var(--border-primary); color: var(--text-primary);"
					bind:value={newDesc}
					placeholder="Project description"
					rows="2"
				></textarea>
			</div>
			<div class="flex justify-end gap-2 pt-2">
				<button onclick={() => showCreate = false} class="btn btn-ghost btn-sm">Cancel</button>
				<button onclick={createProject} class="btn btn-primary btn-sm" disabled={!newName || !newSlug}>Create</button>
			</div>
		</div>
	</Modal>
{/if}

{#if deleteTarget}
	<ConfirmDialog
		title="Delete Project"
		message={`Are you sure you want to delete "${deleteTarget.name}"? This will also delete all associated databases and API keys.`}
		onconfirm={deleteProject}
		oncancel={() => deleteTarget = null}
	/>
{/if}
