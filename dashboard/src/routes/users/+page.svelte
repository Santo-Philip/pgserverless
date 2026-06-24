<script lang="ts">
	import { onMount } from 'svelte';
	import { isAuthenticated } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api/client';
	import Card from '$lib/components/Card.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import Breadcrumbs from '$lib/components/Breadcrumbs.svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';

	let users = $state<User[]>([]);
	let loading = $state(true);
	let currentUserId = $state('');
	let isAdmin = $state(false);
	let processing = $state<Record<string, boolean>>({});

	onMount(async () => {
		if (!$isAuthenticated) { goto('/login'); return; }
		try {
			const me = await api.get<{ user_id: string; is_super_admin: boolean }>('/api/v1/platform/me');
			currentUserId = me.user_id;
			isAdmin = me.is_super_admin;
			const result = await api.listUsers();
			users = result.data;
		} catch {}
		loading = false;
	});

	async function toggleStatus(user: User) {
		if (processing[user.id]) return;
		processing[user.id] = true;
		try {
			if (user.status === 'active') {
				await api.suspendUser(user.id);
				user.status = 'suspended';
			} else {
				await api.activateUser(user.id);
				user.status = 'active';
			}
		} catch {}
		processing[user.id] = false;
	}
</script>

<Breadcrumbs items={[{ label: 'Users' }]} />

<div class="max-w-7xl mx-auto">
	<div class="mb-8">
		<h1 class="text-2xl font-bold">Users</h1>
		<p class="text-sm mt-1" style="color: var(--text-secondary)">Manage platform user accounts</p>
	</div>

	{#if loading}
		<div class="card p-0"><Skeleton rows={5} /></div>
	{:else if users.length === 0}
		<div class="card p-12 text-center" style="color: var(--text-secondary)">No users found.</div>
	{:else}
		<Card>
			<div class="table-wrap overflow-x-auto -mx-5 -mb-5">
				<table class="w-full">
					<thead>
						<tr>
							<th>Name</th>
							<th>Email</th>
							<th>Role</th>
							<th>Status</th>
							<th>Created</th>
							{#if isAdmin}
								<th class="text-right">Actions</th>
							{/if}
						</tr>
					</thead>
					<tbody>
						{#each users as user}
							<tr>
								<td>
									<span class="font-medium">{user.name || '—'}</span>
									{#if user.is_super_admin}
										<span class="ml-2 badge text-xs" style="background-color: rgba(168,85,247,0.1); color: #a855f7">Admin</span>
									{/if}
								</td>
								<td style="color: var(--text-secondary)">{user.email}</td>
								<td>
									<span class="badge text-xs" style="background-color: {user.is_super_admin ? 'rgba(168,85,247,0.1)' : 'rgba(107,114,128,0.1)'}; color: {user.is_super_admin ? '#a855f7' : 'var(--text-tertiary)'}">
										{user.is_super_admin ? 'Admin' : 'User'}
									</span>
								</td>
								<td><Badge status={user.status} /></td>
								<td class="text-xs" style="color: var(--text-tertiary)">{new Date(user.created_at).toLocaleDateString()}</td>
								{#if isAdmin}
									<td class="text-right">
										{#if user.id !== currentUserId}
											<button
												onclick={() => toggleStatus(user)}
												disabled={processing[user.id]}
												class="btn btn-sm {user.status === 'active' ? 'btn-ghost' : 'btn-primary'}"
												style={user.status === 'active' ? 'color: var(--danger)' : ''}
											>
												{processing[user.id] ? '...' : user.status === 'active' ? 'Suspend' : 'Activate'}
											</button>
										{/if}
									</td>
								{/if}
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</Card>
	{/if}
</div>
