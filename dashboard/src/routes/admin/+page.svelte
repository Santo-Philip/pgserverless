<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';
  import type { User } from '$lib/types';
  import Card from '$lib/components/Card.svelte';
  import Modal from '$lib/components/Modal.svelte';
  import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
  import LoadingCard from '$lib/components/LoadingCard.svelte';
  import Badge from '$lib/components/Badge.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';

  let loading = $state(true);
  let error = $state('');
  let users = $state<User[]>([]);
  let currentUser = $state<User | null>(null);

  let createModal = $state(false);
  let editModal = $state(false);
  let resetPwdModal = $state(false);
  let deleteConfirm = $state(false);

  let formEmail = $state('');
  let formName = $state('');
  let formPassword = $state('');
  let formRole = $state<'admin' | 'superadmin' | 'viewer'>('admin');
  let editUserId = $state('');
  let editName = $state('');
  let editRole = $state<'admin' | 'superadmin' | 'viewer'>('admin');
  let editActive = $state(true);
  let resetPwdId = $state('');
  let resetPwdNewPassword = $state('');
  let deleteUserId = $state('');
  let deleteUserName = $state('');

  let loadingStates: Record<string, boolean> = $state({});

  onMount(async () => {
    try {
      const me = await api.getMe();
      currentUser = me;
      users = await api.listUsers();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load users';
    } finally {
      loading = false;
    }
  });

  async function handleCreate() {
    loadingStates['create'] = true;
    try {
      await api.createUser({ email: formEmail, name: formName, password: formPassword, role: formRole });
      createModal = false;
      formEmail = '';
      formName = '';
      formPassword = '';
      formRole = 'admin';
      users = await api.listUsers();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to create user';
    } finally {
      loadingStates['create'] = false;
    }
  }

  function openEdit(u: User) {
    editUserId = u.id;
    editName = u.name;
    editRole = u.role;
    editActive = u.is_active;
    editModal = true;
  }

  async function handleEdit() {
    loadingStates['edit'] = true;
    try {
      await api.updateUser(editUserId, { name: editName, role: editRole, is_active: editActive });
      editModal = false;
      users = await api.listUsers();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to update user';
    } finally {
      loadingStates['edit'] = false;
    }
  }

  function openResetPwd(u: User) {
    resetPwdId = u.id;
    resetPwdNewPassword = '';
    resetPwdModal = true;
  }

  async function handleResetPwd() {
    loadingStates['resetPwd'] = true;
    try {
      await api.updateUserPassword(resetPwdId, resetPwdNewPassword);
      resetPwdModal = false;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to reset password';
    } finally {
      loadingStates['resetPwd'] = false;
    }
  }

  function openDelete(u: User) {
    deleteUserId = u.id;
    deleteUserName = u.name;
    deleteConfirm = true;
  }

  async function handleDelete() {
    if (deleteUserId === currentUser?.id) {
      error = 'Cannot delete your own account';
      return;
    }
    loadingStates['delete'] = true;
    try {
      await api.deleteUser(deleteUserId);
      deleteConfirm = false;
      users = users.filter(u => u.id !== deleteUserId);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to delete user';
    } finally {
      loadingStates['delete'] = false;
    }
  }

  function roleBadgeColor(role: string): { bg: string; color: string } {
    switch (role) {
      case 'superadmin': return { bg: 'rgba(239,68,68,0.1)', color: 'var(--danger)' };
      case 'admin': return { bg: 'rgba(59,130,246,0.1)', color: 'var(--accent)' };
      case 'viewer': return { bg: 'rgba(107,114,128,0.1)', color: 'var(--text-tertiary)' };
      default: return { bg: 'rgba(107,114,128,0.1)', color: 'var(--text-tertiary)' };
    }
  }
</script>

<div class="max-w-6xl mx-auto">
  <div class="flex items-center justify-between mb-6">
    <h1 class="text-2xl font-bold" style="color: var(--text-primary);">Admin Users</h1>
    <button onclick={() => { formEmail = ''; formName = ''; formPassword = ''; createModal = true; }} class="btn btn-primary btn-sm">+ Create User</button>
  </div>

  {#if error}
    <div class="card p-4 mb-4" style="border-color: rgba(239,68,68,0.3);">
      <p class="text-sm" style="color: var(--danger);">{error}</p>
    </div>
  {/if}

  {#if loading}
    <LoadingCard />
  {:else if users.length === 0}
    <EmptyState title="No users" description="Create your first admin user" />
  {:else}
    <div class="card overflow-hidden">
      <table class="table-wrap w-full">
        <thead>
          <tr>
            <th>Name</th>
            <th>Email</th>
            <th>Role</th>
            <th>Status</th>
            <th>Last Login</th>
            <th>Created</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each users as u}
            <tr class={u.id === currentUser?.id ? 'opacity-80' : ''}>
              <td class="font-medium">
                {u.name}
                {#if u.id === currentUser?.id}
                  <span class="text-xs ml-1" style="color: var(--text-tertiary);">(you)</span>
                {/if}
              </td>
              <td class="text-xs font-mono">{u.email}</td>
              <td>
                <span
                  class="badge"
                  style="background-color: {roleBadgeColor(u.role).bg}; color: {roleBadgeColor(u.role).color};"
                >{u.role}</span>
              </td>
              <td><Badge variant={u.is_active ? 'success' : 'default'}>active</Badge></td>
              <td class="text-xs">{u.last_login_at ? new Date(u.last_login_at).toLocaleString() : 'Never'}</td>
              <td class="text-xs">{new Date(u.created_at).toLocaleDateString()}</td>
              <td>
                <div class="flex gap-1">
                  <button onclick={() => openEdit(u)} class="btn btn-ghost btn-sm" title="Edit">✎</button>
                  <button onclick={() => openResetPwd(u)} class="btn btn-ghost btn-sm" title="Reset Password">🔑</button>
                  <button onclick={() => openDelete(u)} class="btn btn-ghost btn-sm" style="color: var(--danger);" title="Delete" disabled={u.id === currentUser?.id}>✕</button>
                </div>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<Modal title="Create User" open={createModal} onclose={() => createModal = false}>
  <div class="mb-3">
    <label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">Email</label>
    <input type="email" bind:value={formEmail} class="input" placeholder="user@example.com" />
  </div>
  <div class="mb-3">
    <label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">Name</label>
    <input type="text" bind:value={formName} class="input" placeholder="John Doe" />
  </div>
  <div class="mb-3">
    <label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">Password</label>
    <input type="password" bind:value={formPassword} class="input" />
  </div>
  <div class="mb-3">
    <label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">Role</label>
    <select bind:value={formRole} class="input">
      <option value="admin">Admin</option>
      <option value="superadmin">Super Admin</option>
      <option value="viewer">Viewer</option>
    </select>
  </div>
  <div class="flex justify-end gap-3">
    <button onclick={() => createModal = false} class="btn btn-secondary">Cancel</button>
    <button onclick={handleCreate} disabled={loadingStates['create']} class="btn btn-primary">{loadingStates['create'] ? '...' : 'Create'}</button>
  </div>
</Modal>

<Modal title="Edit User" open={editModal} onclose={() => editModal = false}>
  <div class="mb-3">
    <label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">Name</label>
    <input type="text" bind:value={editName} class="input" />
  </div>
  <div class="mb-3">
    <label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">Role</label>
    <select bind:value={editRole} class="input">
      <option value="admin">Admin</option>
      <option value="superadmin">Super Admin</option>
      <option value="viewer">Viewer</option>
    </select>
  </div>
  <div class="mb-3">
    <label class="flex items-center gap-2 text-xs">
      <input type="checkbox" bind:checked={editActive} /> Active
    </label>
  </div>
  <div class="flex justify-end gap-3">
    <button onclick={() => editModal = false} class="btn btn-secondary">Cancel</button>
    <button onclick={handleEdit} disabled={loadingStates['edit']} class="btn btn-primary">{loadingStates['edit'] ? '...' : 'Save'}</button>
  </div>
</Modal>

<Modal title="Reset Password" open={resetPwdModal} onclose={() => resetPwdModal = false}>
  <div class="mb-3">
    <label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">New Password</label>
    <input type="password" bind:value={resetPwdNewPassword} class="input" />
  </div>
  <div class="flex justify-end gap-3">
    <button onclick={() => resetPwdModal = false} class="btn btn-secondary">Cancel</button>
    <button onclick={handleResetPwd} disabled={loadingStates['resetPwd']} class="btn btn-primary">{loadingStates['resetPwd'] ? '...' : 'Reset'}</button>
  </div>
</Modal>

<ConfirmDialog
  open={deleteConfirm}
  title="Delete User"
  description={'Are you sure you want to delete "' + deleteUserName + '"? This action cannot be undone.'}
  confirmLabel="Delete"
  variant="danger"
  onconfirm={handleDelete}
  oncancel={() => deleteConfirm = false}
  loading={loadingStates['delete']}
/>
