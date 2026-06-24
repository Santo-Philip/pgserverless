<script lang="ts">
	import NavItem from './NavItem.svelte';
	import CommandPalette from './CommandPalette.svelte';
	import { page } from '$app/stores';
	import { isAuthenticated, logout } from '$lib/stores/auth';
	import type { Snippet } from 'svelte';

	let { children }: { children: Snippet } = $props();

	let commandOpen = $state(false);

	function handleKeydown(e: KeyboardEvent) {
		if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
			e.preventDefault();
			commandOpen = true;
		}
		if (e.key === 'Escape') commandOpen = false;
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="flex h-screen overflow-hidden" style="background-color: var(--bg-primary);">
	<aside class="flex-shrink-0 flex flex-col border-r" style="width: var(--sidebar-width); background-color: var(--bg-secondary); border-color: var(--border-primary);">
		<div class="flex items-center gap-3 px-5 h-14 border-b flex-shrink-0" style="border-color: var(--border-primary);">
			<div class="w-7 h-7 rounded-lg flex items-center justify-center text-sm font-bold" style="background-color: var(--accent); color: #fff;">N</div>
			<span class="font-semibold text-sm">Nexbic</span>
		</div>

		<nav class="flex-1 overflow-y-auto p-3 space-y-1">
			<NavItem href="/" icon="◉" active={$page.url.pathname === '/'}>Dashboard</NavItem>
			<NavItem href="/apps" icon="▦" active={$page.url.pathname.startsWith('/apps')}>Applications</NavItem>
			<NavItem href="/database" icon="▤" active={$page.url.pathname === '/database'}>Database</NavItem>
			<NavItem href="/api-keys" icon="🔑" active={$page.url.pathname === '/api-keys'}>API Keys</NavItem>
			<NavItem href="/rest-api" icon="↗" active={$page.url.pathname === '/rest-api'}>REST API</NavItem>
			<NavItem href="/extensions" icon="◆" active={$page.url.pathname === '/extensions'}>Extensions</NavItem>
			<NavItem href="/logs" icon="☰" active={$page.url.pathname.startsWith('/logs')}>Logs</NavItem>
			<NavItem href="/settings" icon="⚙" active={$page.url.pathname === '/settings'}>Settings</NavItem>

			<div class="pt-4 mt-4 border-t" style="border-color: var(--border-primary);">
				<NavItem href="/users" icon="👥" active={$page.url.pathname === '/users'}>Users</NavItem>
			</div>
		</nav>
	</aside>

	<div class="flex-1 flex flex-col overflow-hidden">
		<header class="flex items-center justify-between px-6 flex-shrink-0 border-b" style="height: var(--topbar-height); background-color: var(--bg-secondary); border-color: var(--border-primary);">
			<div class="flex items-center gap-4">
				<button onclick={() => commandOpen = true} class="btn btn-ghost btn-sm flex items-center gap-2" style="color: var(--text-tertiary)">
					<span class="text-xs">⌘K</span>
					<span class="text-xs hidden sm:inline">Search...</span>
				</button>
			</div>
			<div class="flex items-center gap-3">
				<button onclick={logout} class="btn btn-ghost btn-sm">Logout</button>
			</div>
		</header>

		<main class="flex-1 overflow-y-auto p-6">
			{@render children()}
		</main>
	</div>
</div>

<CommandPalette open={commandOpen} onclose={() => commandOpen = false} />
