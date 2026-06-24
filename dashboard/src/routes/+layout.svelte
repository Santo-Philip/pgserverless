<script lang="ts">
	import '../app.css';
	import { isAuthenticated } from '$lib/stores/auth';
	import { page } from '$app/stores';
	import { logout } from '$lib/stores/auth';
</script>

<div class="flex h-screen">
	<aside class="w-64 bg-slate-900 text-white p-4 flex flex-col">
		<a href="/" class="text-xl font-bold mb-8 text-white no-underline">
			Nexbic Platform
		</a>

		{#if $isAuthenticated}
			<nav class="flex-1 space-y-2">
				<a href="/" class="block px-4 py-2 rounded hover:bg-slate-700 transition-colors">
					Dashboard
				</a>
				<a href="/apps" class="block px-4 py-2 rounded hover:bg-slate-700 transition-colors">
					Applications
				</a>
				<a href="/users" class="block px-4 py-2 rounded hover:bg-slate-700 transition-colors">
					Users
				</a>
				<a href="/settings" class="block px-4 py-2 rounded hover:bg-slate-700 transition-colors">
					Settings
				</a>
			</nav>

			<button
				onclick={logout}
				class="w-full px-4 py-2 text-left rounded hover:bg-red-700 transition-colors"
			>
				Logout
			</button>
		{/if}
	</aside>

	<main class="flex-1 overflow-auto p-8">
		<slot />
	</main>
</div>
