<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';
  import type { ExtensionInfo } from '$lib/types';
  import Card from '$lib/components/Card.svelte';
  import LoadingCard from '$lib/components/LoadingCard.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';

  let loading = $state(true);
  let error = $state('');
  let extensions = $state<ExtensionInfo[]>([]);
  let search = $state('');
  let actionLoading = $state<Record<string, boolean>>({});

  let filtered = $derived(
    extensions.filter(e =>
      e.name.toLowerCase().includes(search.toLowerCase()) ||
      e.comment?.toLowerCase().includes(search.toLowerCase())
    )
  );

  onMount(async () => {
    try {
      extensions = await api.listExtensions();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load extensions';
    } finally {
      loading = false;
    }
  });

  async function toggleExtension(ext: ExtensionInfo) {
    const key = ext.name;
    actionLoading[key] = true;
    try {
      if (ext.installed) {
        await api.uninstallExtension(ext.name);
      } else {
        await api.installExtension(ext.name);
      }
      extensions = extensions.map(e =>
        e.name === ext.name ? { ...e, installed: !e.installed, installed_version: e.installed ? null : e.version } : e
      );
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to toggle extension';
    } finally {
      actionLoading[key] = false;
    }
  }
</script>

<div class="max-w-6xl mx-auto">
  <h1 class="text-2xl font-bold mb-6" style="color: var(--text-primary);">Extension Manager</h1>

  {#if error}
    <div class="card p-4 mb-4" style="border-color: rgba(239,68,68,0.3);">
      <p class="text-sm" style="color: var(--danger);">{error}</p>
    </div>
  {/if}

  <div class="mb-4">
    <input
      type="text"
      bind:value={search}
      placeholder="Search extensions..."
      class="input max-w-md"
    />
  </div>

  {#if loading}
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      {#each [1,2,3,4,5,6] as _}
        <LoadingCard />
      {/each}
    </div>
  {:else if filtered.length === 0}
    <EmptyState title="No extensions found" description={search ? 'Try a different search term' : 'No extensions available'} />
  {:else}
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      {#each filtered as ext}
        <div class="card p-5 card-hover">
          <div class="flex items-start justify-between mb-3">
            <div>
              <h3 class="text-sm font-semibold" style="color: var(--text-primary);">{ext.name}</h3>
              <p class="text-xs mt-0.5 font-mono" style="color: var(--text-tertiary);">
                {ext.installed_version || ext.version}
              </p>
            </div>
            <span
              class="inline-flex items-center px-2 py-0.5 rounded-md text-xs font-medium"
              style="background-color: {ext.installed ? 'rgba(34,197,94,0.1)' : 'rgba(107,114,128,0.1)'}; color: {ext.installed ? 'var(--success)' : 'var(--text-tertiary)'};"
            >
              {ext.installed ? 'Installed' : 'Available'}
            </span>
          </div>
          <p class="text-xs mb-4 truncate-2" style="color: var(--text-secondary); min-height: 2.5em;">
            {ext.comment || 'No description available'}
          </p>
          <button
            onclick={() => toggleExtension(ext)}
            disabled={actionLoading[ext.name]}
            class="btn w-full {ext.installed ? 'btn-danger' : 'btn-primary'} btn-sm"
          >
            {actionLoading[ext.name] ? '...' : ext.installed ? 'Uninstall' : 'Install'}
          </button>
        </div>
      {/each}
    </div>
  {/if}
</div>
