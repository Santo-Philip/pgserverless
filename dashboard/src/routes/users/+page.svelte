<script lang="ts">
	import { onMount } from 'svelte';
	import { isAuthenticated } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api/client';

	let users = $state<User[]>([]);
	let loading = $state(true);
	let currentUserId = $state('');
	let isAdmin = $state(false);
	let processing = $state<Record<string, boolean>>({});

	onMount(async () => {
		if (!$isAuthenticated) {
			goto('/login');
			return;
		}

		try {
			const me = await api.get<{ user_id: string; is_super_admin: boolean }>('/api/v1/platform/me');
			currentUserId = me.user_id;
			isAdmin = me.is_super_admin;
			const result = await api.listUsers();
			users = result.data;
		} catch (e) {
			console.error('Failed to load users', e);
		} finally {
			loading = false;
		}
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
		} catch (e) {
			console.error('Failed to update user', e);
		} finally {
			processing[user.id] = false;
		}
	}
</script>

<div class="max-w-6xl mx-auto">
	<h1 class="text-3xl font-bold mb-8">Users</h1>

	{#if loading}
		<p class="text-gray-500">Loading...</p>
	{:else if users.length === 0}
		<div class="bg-white rounded-lg shadow p-12 text-center">
			<p class="text-gray-500">No users found.</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow overflow-hidden">
			<table class="w-full">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Email</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Role</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Created</th>
						{#if isAdmin}
							<th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Actions</th>
						{/if}
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-200">
					{#each users as user}
						<tr>
							<td class="px-6 py-4 font-medium">
								{user.name || '—'}
								{#if user.is_super_admin}
									<span class="ml-2 text-xs uppercase px-2 py-0.5 rounded bg-purple-100 text-purple-700">Admin</span>
								{/if}
							</td>
							<td class="px-6 py-4">{user.email}</td>
							<td class="px-6 py-4">
								<span class="text-xs uppercase px-2 py-1 rounded {user.is_super_admin ? 'bg-purple-100 text-purple-700' : 'bg-gray-100 text-gray-600'}">
									{user.is_super_admin ? 'Admin' : 'User'}
								</span>
							</td>
							<td class="px-6 py-4">
								<span class="text-xs uppercase px-2 py-1 rounded {user.status === 'active' ? 'bg-green-100 text-green-700' : user.status === 'suspended' ? 'bg-red-100 text-red-700' : 'bg-gray-100 text-gray-600'}">
									{user.status}
								</span>
							</td>
							<td class="px-6 py-4 text-sm text-gray-500">{new Date(user.created_at).toLocaleDateString()}</td>
							{#if isAdmin}
								<td class="px-6 py-4 text-right">
									{#if user.id !== currentUserId}
										<button
											onclick={() => toggleStatus(user)}
											disabled={processing[user.id]}
											class="text-sm px-3 py-1 rounded border {user.status === 'active' ? 'border-red-300 text-red-600 hover:bg-red-50' : 'border-green-300 text-green-600 hover:bg-green-50'} disabled:opacity-50"
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
	{/if}
</div>
