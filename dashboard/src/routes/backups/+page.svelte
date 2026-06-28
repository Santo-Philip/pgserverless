<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';
  import type { BackupInfo } from '$lib/types';
  import Card from '$lib/components/Card.svelte';
  import Badge from '$lib/components/Badge.svelte';
  import LoadingCard from '$lib/components/LoadingCard.svelte';
  import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';

  let loading = $state(true);
  let error = $state('');
  let backups = $state<BackupInfo[]>([]);
  let actionLoading = $state(false);

  let deleteId = $state<string | null>(null);
  let restoreId = $state<string | null>(null);

  function formatBytes(bytes: number): string {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  }

  onMount(async () => {
    try {
      backups = await api.listBackups();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load backups';
    } finally {
      loading = false;
    }
  });

  async function handleCreate() {
    actionLoading = true;
    try {
      await api.createBackup();
      backups = await api.listBackups();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to create backup';
    } finally {
      actionLoading = false;
    }
  }

  async function handleRestore() {
    if (!restoreId) return;
    actionLoading = true;
    try {
      await api.restoreBackup(restoreId);
      restoreId = null;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to restore backup';
    } finally {
      actionLoading = false;
    }
  }

  async function handleDelete() {
    if (!deleteId) return;
    actionLoading = true;
    try {
      await api.deleteBackup(deleteId);
      backups = backups.filter(b => b.id !== deleteId);
      deleteId = null;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to delete backup';
    } finally {
      actionLoading = false;
    }
  }

  async function handleVerify(id: string) {
    actionLoading = true;
    try {
      await api.verifyBackup(id);
      backups = await api.listBackups();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to verify backup';
    } finally {
      actionLoading = false;
    }
  }

  async function handleDownload(id: string) {
    try {
      const url = await api.downloadBackup(id);
      if (url) window.open(url, '_blank');
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to download backup';
    }
  }
</script>

<div class="max-w-6xl mx-auto">
  <div class="flex items-center justify-between mb-6">
    <h1 class="text-2xl font-bold" style="color: var(--text-primary);">Backup Manager</h1>
    <button onclick={handleCreate} disabled={actionLoading} class="btn btn-primary btn-sm">
      {actionLoading ? 'Creating...' : '+ Create Backup'}
    </button>
  </div>

  {#if error}
    <div class="card p-4 mb-4" style="border-color: rgba(239,68,68,0.3);">
      <p class="text-sm" style="color: var(--danger);">{error}</p>
    </div>
  {/if}

  {#if loading}
    {#each [1,2,3] as _}
      <LoadingCard />
    {/each}
  {:else if backups.length === 0}
    <EmptyState title="No backups" description="Create your first backup to get started">
      <button onclick={handleCreate} disabled={actionLoading} class="btn btn-primary">{actionLoading ? 'Creating...' : 'Create Backup'}</button>
    </EmptyState>
  {:else}
    <div class="card overflow-hidden">
      <table class="table-wrap w-full">
        <thead>
          <tr>
            <th>ID</th>
            <th>Database</th>
            <th>Type</th>
            <th>Size</th>
            <th>Status</th>
            <th>Started</th>
            <th>Verified</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each backups as b}
            <tr>
              <td class="font-mono text-xs">{b.id.slice(0, 8)}...</td>
              <td class="font-medium">{b.database}</td>
              <td><Badge variant={b.type} /></td>
              <td class="font-mono text-xs">{formatBytes(b.size_bytes)}</td>
              <td><Badge variant={b.status} /></td>
              <td class="text-xs">{new Date(b.started_at).toLocaleString()}</td>
              <td>{b.verified ? '✓' : '✗'}</td>
              <td>
                <div class="flex gap-1">
                  <button onclick={() => handleDownload(b.id)} class="btn btn-ghost btn-sm" title="Download">↓</button>
                  <button onclick={() => restoreId = b.id} class="btn btn-ghost btn-sm" title="Restore">↻</button>
                  <button onclick={() => handleVerify(b.id)} class="btn btn-ghost btn-sm" title="Verify">✓</button>
                  <button onclick={() => deleteId = b.id} class="btn btn-ghost btn-sm" style="color: var(--danger);" title="Delete">✕</button>
                </div>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<ConfirmDialog
  open={deleteId !== null}
  title="Delete Backup"
  description="Are you sure you want to delete this backup? This cannot be undone."
  confirmLabel="Delete"
  variant="danger"
  onconfirm={handleDelete}
  oncancel={() => deleteId = null}
  loading={actionLoading}
/>

<ConfirmDialog
  open={restoreId !== null}
  title="Restore Backup"
  description="Are you sure you want to restore this backup? Current data may be overwritten."
  confirmLabel="Restore"
  variant="primary"
  onconfirm={handleRestore}
  oncancel={() => restoreId = null}
  loading={actionLoading}
/>
