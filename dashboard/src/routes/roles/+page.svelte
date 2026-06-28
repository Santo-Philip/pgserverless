<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';
  import type { PgRole, PgPrivilege } from '$lib/types';
  import Card from '$lib/components/Card.svelte';
  import Modal from '$lib/components/Modal.svelte';
  import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
  import LoadingCard from '$lib/components/LoadingCard.svelte';
  import Badge from '$lib/components/Badge.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';

  let loading = $state(true);
  let error = $state('');
  let roles = $state<PgRole[]>([]);
  let selectedRole = $state<string | null>(null);
  let roleDetails = $state<Record<string, unknown> | null>(null);
  let detailsLoading = $state(false);

  let createModal = $state(false);
  let roleName = $state('');
  let rolePassword = $state('');
  let roleLogin = $state(true);
  let roleSuperuser = $state(false);
  let roleCreatedb = $state(false);
  let roleCreaterole = $state(false);
  let roleReplication = $state(false);
  let roleConnLimit = $state(-1);

  let dropRoleName = $state<string | null>(null);
  let dropConfirm = $state(false);

  let resetPwdModal = $state(false);
  let resetPwdName = $state<string | null>(null);
  let resetPwdPassword = $state('');

  let activeTab = $state<'info' | 'members' | 'privileges'>('info');
  let members = $state<{ direct: string[]; indirect: string[] }>({ direct: [], indirect: [] });
  let privileges = $state<PgPrivilege[]>([]);
  let grantModal = $state(false);
  let grantPrivilege = $state('');
  let grantTarget = $state('');

  let loadingStates: Record<string, boolean> = $state({});

  onMount(async () => {
    try {
      roles = await api.listRoles();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load roles';
    } finally {
      loading = false;
    }
  });

  async function selectRole(name: string) {
    selectedRole = name;
    detailsLoading = true;
    activeTab = 'info';
    try {
      const [details, mems, privs] = await Promise.all([
        api.getRole(name),
        api.getRoleMembers(name),
        api.listDatabasePrivileges(name),
      ]);
      roleDetails = details as unknown as Record<string, unknown>;
      members = mems;
      privileges = privs;
    } catch {
      roleDetails = null;
    } finally {
      detailsLoading = false;
    }
  }

  async function handleCreate() {
    loadingStates['create'] = true;
    try {
      await api.createRole({
        name: roleName,
        password: rolePassword || undefined,
        login: roleLogin,
        superuser: roleSuperuser,
        createdb: roleCreatedb,
        createrole: roleCreaterole,
        replication: roleReplication,
        connection_limit: roleConnLimit > 0 ? roleConnLimit : undefined,
      });
      createModal = false;
      roleName = '';
      rolePassword = '';
      roles = await api.listRoles();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to create role';
    } finally {
      loadingStates['create'] = false;
    }
  }

  async function handleDrop() {
    if (dropRoleName === null) return;
    loadingStates['drop'] = true;
    try {
      await api.dropRole(dropRoleName);
      dropConfirm = false;
      dropRoleName = '';
      selectedRole = null;
      roles = await api.listRoles();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to drop role';
    } finally {
      loadingStates['drop'] = false;
    }
  }

  async function handleResetPassword() {
    if (resetPwdName === null) return;
    loadingStates['resetPwd'] = true;
    try {
      await api.resetPassword(resetPwdName, resetPwdPassword);
      resetPwdModal = false;
      resetPwdName = '';
      resetPwdPassword = '';
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to reset password';
    } finally {
      loadingStates['resetPwd'] = false;
    }
  }

  async function handleGrant() {
    loadingStates['grant'] = true;
    try {
      await api.grantDatabase(selectedRole!, grantTarget, [grantPrivilege]);
      grantModal = false;
      grantPrivilege = '';
      grantTarget = '';
      if (selectedRole) {
        privileges = await api.listDatabasePrivileges(selectedRole);
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to grant privilege';
    } finally {
      loadingStates['grant'] = false;
    }
  }
</script>

<div class="max-w-6xl mx-auto">
  <div class="flex items-center justify-between mb-6">
    <h1 class="text-2xl font-bold" style="color: var(--text-primary);">Role Management</h1>
    <button onclick={() => { roleName = ''; rolePassword = ''; createModal = true; }} class="btn btn-primary btn-sm">+ Create Role</button>
  </div>

  <div class="flex flex-col lg:flex-row gap-6">
    <div class="w-full lg:w-64 flex-shrink-0">
      <div class="card p-4">
        <h3 class="text-xs font-semibold uppercase tracking-wider mb-3" style="color: var(--text-tertiary);">Roles</h3>
        {#if loading}
          {#each [1,2,3] as _}
            <div class="skeleton h-10 w-full mb-2 rounded-lg"></div>
          {/each}
        {:else if roles.length === 0}
          <EmptyState title="No roles" />
        {:else}
          <div class="space-y-1">
            {#each roles as r}
              <button
                onclick={() => selectRole(r.rolname)}
                class="w-full text-left p-2 rounded-lg text-sm transition-colors"
                style="background-color: {selectedRole === r.rolname ? 'var(--accent-muted)' : 'transparent'}; color: {selectedRole === r.rolname ? 'var(--accent)' : 'var(--text-secondary)'};"
              >
                <div class="flex items-center justify-between">
                  <span class="font-medium">{r.rolname}</span>
                  {#if r.rolsuper}
                    <Badge variant="success">superuser</Badge>
                  {/if}
                </div>
                <div class="text-xs mt-0.5" style="color: var(--text-tertiary);">{r.rolcanlogin ? 'Can login' : 'No login'}</div>
              </button>
            {/each}
          </div>
        {/if}
      </div>
    </div>

    <div class="flex-1 min-w-0">
      {#if !selectedRole}
        <div class="card p-12 text-center">
          <div class="text-4xl mb-3">◎</div>
          <h3 class="text-base font-semibold" style="color: var(--text-secondary);">Select a Role</h3>
          <p class="text-sm mt-1" style="color: var(--text-tertiary);">Choose a role to view details, members, and privileges</p>
        </div>
      {:else if detailsLoading}
        <LoadingCard />
      {:else}
        <div class="card p-5 mb-4">
          <div class="flex items-center justify-between mb-4">
            <h2 class="text-lg font-semibold" style="color: var(--text-primary);">{selectedRole}</h2>
            <div class="flex gap-2">
              <button onclick={() => { resetPwdName = selectedRole; resetPwdPassword = ''; resetPwdModal = true; }} class="btn btn-secondary btn-sm">Reset Password</button>
              <button onclick={() => { dropRoleName = selectedRole; dropConfirm = true; }} class="btn btn-danger btn-sm">Drop</button>
            </div>
          </div>

          <div class="flex gap-0 border-b mb-4" style="border-color: var(--border-primary);">
            {#each ['info', 'members', 'privileges'] as tab}
              <button onclick={() => activeTab = tab as typeof activeTab} class="tab" class:active={activeTab === tab}>{tab.charAt(0).toUpperCase() + tab.slice(1)}</button>
            {/each}
          </div>

          {#if activeTab === 'info'}
            <div class="grid grid-cols-2 sm:grid-cols-3 gap-3">
              {#each Object.entries(roleDetails || {}).filter(([k]) => !k.startsWith('rol')) as [k, v]}
                <div class="card p-3">
                  <div class="text-xs" style="color: var(--text-tertiary);">{k}</div>
                  <div class="text-sm font-medium mt-0.5" style="color: var(--text-primary);">{String(v ?? '-')}</div>
                </div>
              {/each}
            </div>
          {:else if activeTab === 'members'}
            <div>
              <h4 class="text-sm font-semibold mb-2" style="color: var(--text-secondary);">Direct Members</h4>
              {#if members.direct.length === 0}
                <p class="text-xs" style="color: var(--text-tertiary);">No direct members</p>
              {:else}
                <div class="flex flex-wrap gap-2 mb-4">
                  {#each members.direct as m}
                    <span class="badge" style="background-color: var(--accent-muted); color: var(--accent);">{m}</span>
                  {/each}
                </div>
              {/if}
              <h4 class="text-sm font-semibold mb-2" style="color: var(--text-secondary);">Indirect Members</h4>
              {#if members.indirect.length === 0}
                <p class="text-xs" style="color: var(--text-tertiary);">No indirect members</p>
              {:else}
                <div class="flex flex-wrap gap-2">
                  {#each members.indirect as m}
                    <span class="badge" style="background-color: rgba(245,158,11,0.1); color: var(--warning);">{m}</span>
                  {/each}
                </div>
              {/if}
            </div>
          {:else}
            <div>
              <div class="flex items-center justify-between mb-3">
                <h4 class="text-sm font-semibold" style="color: var(--text-secondary);">Privileges</h4>
                <button onclick={() => { grantPrivilege = ''; grantTarget = ''; grantModal = true; }} class="btn btn-primary btn-sm">+ Grant</button>
              </div>
              {#if privileges.length === 0}
                <p class="text-xs" style="color: var(--text-tertiary);">No privileges granted</p>
              {:else}
                <div class="overflow-x-auto">
                  <table class="table-wrap w-full">
                    <thead><tr><th>Grantor</th><th>Grantee</th><th>Privilege</th><th>Grantable</th></tr></thead>
                    <tbody>
                      {#each privileges as p}
                        <tr>
                          <td class="font-medium">{p.grantor}</td>
                          <td>{p.grantee}</td>
                          <td>{p.privilege_type}</td>
                          <td>{p.is_grantable ? 'Yes' : 'No'}</td>
                        </tr>
                      {/each}
                    </tbody>
                  </table>
                </div>
              {/if}
            </div>
          {/if}
        </div>
      {/if}
    </div>
  </div>
</div>

<Modal title="Create Role" open={createModal} onclose={() => createModal = false}>
  <div class="mb-3"><label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">Role Name</label><input type="text" bind:value={roleName} class="input" placeholder="new_role" /></div>
  <div class="mb-3"><label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">Password</label><input type="password" bind:value={rolePassword} class="input" placeholder="optional" /></div>
  <div class="space-y-2 mb-3">
    {#each [{ key: 'roleLogin', label: 'Can Login' }, { key: 'roleSuperuser', label: 'Superuser' }, { key: 'roleCreatedb', label: 'Create DB' }, { key: 'roleCreaterole', label: 'Create Role' }, { key: 'roleReplication', label: 'Replication' }] as opt}
      <label class="flex items-center gap-2 text-xs">
        <input type="checkbox"
          checked={opt.key === 'roleLogin' ? roleLogin : opt.key === 'roleSuperuser' ? roleSuperuser : opt.key === 'roleCreatedb' ? roleCreatedb : opt.key === 'roleCreaterole' ? roleCreaterole : roleReplication}
          onchange={() => {
            if (opt.key === 'roleLogin') roleLogin = !roleLogin;
            else if (opt.key === 'roleSuperuser') roleSuperuser = !roleSuperuser;
            else if (opt.key === 'roleCreatedb') roleCreatedb = !roleCreatedb;
            else if (opt.key === 'roleCreaterole') roleCreaterole = !roleCreaterole;
            else if (opt.key === 'roleReplication') roleReplication = !roleReplication;
          }}
        /> {opt.label}
      </label>
    {/each}
  </div>
  <div class="mb-3"><label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">Connection Limit</label><input type="number" bind:value={roleConnLimit} class="input" placeholder="-1 (unlimited)" /></div>
  <div class="flex justify-end gap-3">
    <button onclick={() => createModal = false} class="btn btn-secondary">Cancel</button>
    <button onclick={handleCreate} disabled={loadingStates['create']} class="btn btn-primary">{loadingStates['create'] ? '...' : 'Create'}</button>
  </div>
</Modal>

<Modal title="Reset Password" open={resetPwdModal} onclose={() => resetPwdModal = false}>
  <div class="mb-3"><label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">New Password</label><input type="password" bind:value={resetPwdPassword} class="input" /></div>
  <div class="flex justify-end gap-3">
    <button onclick={() => resetPwdModal = false} class="btn btn-secondary">Cancel</button>
    <button onclick={handleResetPassword} disabled={loadingStates['resetPwd']} class="btn btn-primary">{loadingStates['resetPwd'] ? '...' : 'Reset'}</button>
  </div>
</Modal>

<ConfirmDialog
  open={dropConfirm}
  title="Drop Role"
  description={'Are you sure you want to drop "' + dropRoleName + '"? This action cannot be undone.'}
  onconfirm={handleDrop}
  oncancel={() => dropConfirm = false}
  loading={loadingStates['drop']}
/>

<Modal title="Grant Privilege" open={grantModal} onclose={() => grantModal = false}>
  <div class="mb-3"><label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">Privilege</label>
    <select bind:value={grantPrivilege} class="input">
      <option value="">Select...</option>
      <option>SELECT</option><option>INSERT</option><option>UPDATE</option><option>DELETE</option>
      <option>TRUNCATE</option><option>REFERENCES</option><option>TRIGGER</option>
      <option>CREATE</option><option>CONNECT</option><option>TEMPORARY</option>
      <option>EXECUTE</option><option>USAGE</option>
    </select>
  </div>
  <div class="mb-3"><label class="block text-xs font-medium mb-1" style="color: var(--text-secondary);">Database</label><input type="text" bind:value={grantTarget} class="input" placeholder="database_name" /></div>
  <div class="flex justify-end gap-3">
    <button onclick={() => grantModal = false} class="btn btn-secondary">Cancel</button>
    <button onclick={handleGrant} disabled={loadingStates['grant']} class="btn btn-primary">{loadingStates['grant'] ? '...' : 'Grant'}</button>
  </div>
</Modal>
