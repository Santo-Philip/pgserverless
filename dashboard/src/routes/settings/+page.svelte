<script lang="ts">
	import { page } from '$app/stores';
	import { isAuthenticated, logout } from '$lib/stores/auth';
	import Card from '$lib/components/Card.svelte';
</script>

<div class="max-w-2xl mx-auto">
	<h1 class="text-2xl font-bold mb-6" style="color: var(--text-primary);">Settings</h1>

	<Card title="Account">
		<div class="space-y-4">
			<div>
				<label class="block text-sm font-medium mb-1" style="color: var(--text-primary);">Email</label>
				<p class="text-sm" style="color: var(--text-secondary);">Logged in via JWT authentication</p>
			</div>
			<button onclick={logout} class="btn btn-ghost text-red-500 text-sm">Logout</button>
		</div>
	</Card>

	<div class="mt-6">
		<Card title="About">
			<p class="text-sm" style="color: var(--text-secondary);">
				Nexbic Database Platform — Provision, manage, and monitor PostgreSQL databases through a secure REST API.
			</p>
		</Card>
	</div>
</div>
