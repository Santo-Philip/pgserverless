<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';
  import type { User } from '$lib/types';
  import Card from '$lib/components/Card.svelte';
  import Modal from '$lib/components/Modal.svelte';
  import LoadingCard from '$lib/components/LoadingCard.svelte';
  import Badge from '$lib/components/Badge.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';

  let loading = $state(true);
  let error = $state('');
  let user = $state<User | null>(null);
  let sessions = $state<{ id: string; ip: string; user_agent: string; created_at: string; current: boolean }[]>([]);
  let loginHistory = $state<{ id: string; ip: string; user_agent: string; success: boolean; created_at: string }[]>([]);

  let changePwdModal = $state(false);
  let currentPassword = $state('');
  let newPassword = $state('');
  let confirmPassword = $state('');
  let pwdLoading = $state(false);
  let pwdError = $state('');

  let activeTab = $state<'info' | 'sessions' | 'history'>('info');

  onMount(async () => {
    try {
      const [u, s, h] = await Promise.all([
        api.getMe(),
        api.getSessions().catch(() => []),
        api.getLoginHistory().catch(() => []),
      ]);
      user = u;
      sessions = s;
      loginHistory = h;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load profile';
    } finally {
      loading = false;
    }
  });

  async function handleChangePassword() {
    if (newPassword !== confirmPassword) {
      pwdError = 'Passwords do not match';
      return;
    }
    pwdLoading = true;
    pwdError = '';
    try {
      await api.updatePassword({ current_password: currentPassword, new_password: newPassword });
      changePwdModal = false;
      currentPassword = '';
      newPassword = '';
      confirmPassword = '';
    } catch (e) {
      pwdError = e instanceof Error ? e.message : 'Failed to change password';
    } finally {
      pwdLoading = false;
    }
  }
</script>

<div class="max-w-4xl mx-auto">
  <h1 class="text-2xl font-bold mb-6" style="color: var(--text-primary);">Profile & Settings</h1>

  {#if error}
    <div class="card p-4 mb-4" style="border-color: rgba(239,68,68,0.3);">
      <p class="text-sm" style="color: var(--danger);">{error}</p>
    </div>
  {/if}

  {#if loading}
    <div class="space-y-4">
      <LoadingCard />
      <LoadingCard />
    </div>
  {:else if user}
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <div class="lg:col-span-1">
        <div class="card p-6 text-center">
          <div class="w-16 h-16 rounded-2xl flex items-center justify-center text-2xl font-bold mx-auto mb-4" style="background-color: var(--accent-muted); color: var(--accent);">
            {user.name.charAt(0).toUpperCase()}
          </div>
          <h2 class="text-lg font-semibold" style="color: var(--text-primary);">{user.name}</h2>
          <p class="text-sm mt-1" style="color: var(--text-secondary);">{user.email}</p>
          <div class="mt-3">
            <Badge variant={user.role}>{user.role}</Badge>
          </div>
          <div class="mt-4">
            <button onclick={() => { changePwdModal = true; pwdError = ''; }} class="btn btn-secondary btn-sm w-full">Change Password</button>
          </div>
        </div>
      </div>

      <div class="lg:col-span-2">
        <div class="card">
          <div class="flex gap-0 border-b" style="border-color: var(--border-primary);">
            {#each ['info', 'sessions', 'history'] as tab}
              <button
                onclick={() => activeTab = tab as typeof activeTab}
                class="tab"
                class:active={activeTab === tab}
              >{tab.charAt(0).toUpperCase() + tab.slice(1)}</button>
            {/each}
          </div>

          <div class="p-5">
            {#if activeTab === 'info'}
              <div class="space-y-4">
                {#each Object.entries(user).filter(([k]) => !k.startsWith('_')) as [key, val]}
                  <div class="flex items-center justify-between py-2 border-b" style="border-color: var(--border-primary);">
                    <span class="text-xs font-medium" style="color: var(--text-tertiary); text-transform: capitalize;">{key.replace(/_/g, ' ')}</span>
                    <span class="text-sm" style="color: var(--text-primary);">
                      {#if key === 'is_active'}
                        <Badge variant={val ? 'success' : 'default'}>{val ? 'active' : 'inactive'}</Badge>
                      {:else if key === 'created_at' || key === 'last_login_at'}
                        {val ? new Date(val as string).toLocaleString() : 'Never'}
                      {:else}
                        {String(val ?? '-')}
                      {/if}
                    </span>
                  </div>
                {/each}
              </div>

            {:else if activeTab === 'sessions'}
              {#if sessions.length === 0}
                <EmptyState title="No active sessions" />
              {:else}
                <div class="space-y-2">
                  {#each sessions as s}
                    <div class="flex items-center justify-between p-3 rounded-lg" style="background-color: var(--bg-hover);">
                      <div>
                        <div class="flex items-center gap-2">
                          <span class="text-sm font-medium" style="color: var(--text-primary);">{s.ip}</span>
                          {#if s.current}
                            <span class="badge text-xs" style="background-color: rgba(34,197,94,0.1); color: var(--success);">Current</span>
                          {/if}
                        </div>
                        <div class="text-xs mt-0.5" style="color: var(--text-tertiary);">{s.user_agent}</div>
                      </div>
                      <div class="text-xs" style="color: var(--text-tertiary);">
                        {new Date(s.created_at).toLocaleString()}
                      </div>
                    </div>
                  {/each}
                </div>
              {/if}

            {:else if activeTab === 'history'}
              {#if loginHistory.length === 0}
                <EmptyState title="No login history" />
              {:else}
                <div class="overflow-x-auto">
                  <table class="table-wrap w-full">
                    <thead>
                      <tr>
                        <th>IP</th>
                        <th>User Agent</th>
                        <th>Success</th>
                        <th>Time</th>
                      </tr>
                    </thead>
                    <tbody>
                      {#each loginHistory as h}
                        <tr>
                          <td class="font-mono text-xs">{h.ip}</td>
                          <td class="text-xs truncate max-w-xs">{h.user_agent}</td>
                          <td>
                            <span style="color: {h.success ? 'var(--success)' : 'var(--danger)'};">{h.success ? '✓' : '✗'}</span>
                          </td>
                          <td class="text-xs whitespace-nowrap">{new Date(h.created_at).toLocaleString()}</td>
                        </tr>
                      {/each}
                    </tbody>
                  </table>
                </div>
              {/if}
            {/if}
          </div>
        </div>
      </div>
    </div>
  {/if}
</div>

<Modal title="Change Password" open={changePwdModal} onclose={() => changePwdModal = false}>
  {#if pwdError}
    <div class="px-4 py-3 rounded-lg text-sm mb-4" style="background-color: rgba(239,68,68,0.1); color: var(--danger);">{pwdError}</div>
  {/if}
  <div class="mb-3">
    <label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">Current Password</label>
    <input type="password" bind:value={currentPassword} class="input" />
  </div>
  <div class="mb-3">
    <label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">New Password</label>
    <input type="password" bind:value={newPassword} class="input" />
  </div>
  <div class="mb-3">
    <label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">Confirm New Password</label>
    <input type="password" bind:value={confirmPassword} class="input" />
  </div>
  <div class="flex justify-end gap-3">
    <button onclick={() => changePwdModal = false} class="btn btn-secondary">Cancel</button>
    <button onclick={handleChangePassword} disabled={pwdLoading} class="btn btn-primary">{pwdLoading ? '...' : 'Change Password'}</button>
  </div>
</Modal>
