<script lang="ts">
	import NavItem from './NavItem.svelte';
	import CommandPalette from './CommandPalette.svelte';
	import { page } from '$app/stores';
	import { isAuthenticated, logout } from '$lib/stores/auth';
	import type { Snippet } from 'svelte';
	import { goto } from '$app/navigation';

	let { children }: { children: Snippet } = $props();

	let commandOpen = $state(false);
	let sidebarOpen = $state(false);

	function toggleSidebar() {
		sidebarOpen = !sidebarOpen;
	}

	function closeSidebar() {
		sidebarOpen = false;
	}

	function handleKeydown(e: KeyboardEvent) {
		if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
			e.preventDefault();
			commandOpen = true;
		}
		if (e.key === 'Escape') {
			commandOpen = false;
			closeSidebar();
		}
	}

	$effect(() => {
		$page.url.pathname;
		closeSidebar();
	});
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="flex h-screen overflow-hidden" style="background-color: var(--bg-primary);">
	{#if sidebarOpen}
		<button class="fixed inset-0 z-40 bg-black/50 lg:hidden" onclick={closeSidebar} aria-label="Close sidebar"></button>
	{/if}

	<aside
		class="flex-shrink-0 flex flex-col border-r fixed inset-y-0 left-0 z-50 transition-transform duration-200 lg:static lg:translate-x-0"
		class:-translate-x-full={!sidebarOpen}
		class:translate-x-0={sidebarOpen}
		style="width: var(--sidebar-width); background-color: var(--bg-secondary); border-color: var(--border-primary);"
		role="navigation"
		aria-label="Main navigation"
	>
		<div class="flex items-center justify-between gap-3 px-5 h-14 border-b flex-shrink-0" style="border-color: var(--border-primary);">
			<div class="flex items-center gap-3">
				<div class="w-7 h-7 rounded-lg flex items-center justify-center text-sm font-bold" style="background-color: var(--accent); color: #fff;" aria-hidden="true">N</div>
				<span class="font-semibold text-sm">Nexbic</span>
			</div>
			<button onclick={closeSidebar} class="lg:hidden p-1 rounded hover:bg-[var(--bg-hover)]" style="color: var(--text-secondary);" aria-label="Close sidebar">
				<svg width="20" height="20" viewBox="0 0 20 20" fill="none" aria-hidden="true"><path d="M15 5L5 15M5 5l10 10" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
			</button>
		</div>

		<nav class="flex-1 overflow-y-auto p-3 space-y-1" aria-label="Sidebar">
			<NavItem href="/" icon="◉" active={$page.url.pathname === '/'}>Dashboard</NavItem>
			<NavItem href="/projects" icon="▦" active={$page.url.pathname.startsWith('/projects')}>Projects</NavItem>
			<NavItem href="/api-keys" icon="🔑" active={$page.url.pathname === '/api-keys'}>API Keys</NavItem>
			<NavItem href="/audit-logs" icon="☰" active={$page.url.pathname === '/audit-logs'}>Audit Logs</NavItem>
			<NavItem href="/plans" icon="☆" active={$page.url.pathname === '/plans'}>Plans</NavItem>
			<NavItem href="/settings" icon="⚙" active={$page.url.pathname === '/settings'}>Settings</NavItem>
		</nav>
	</aside>

	<div class="flex-1 flex flex-col overflow-hidden min-w-0">
		<header class="flex items-center justify-between px-4 sm:px-6 flex-shrink-0 border-b" style="height: var(--topbar-height); background-color: var(--bg-secondary); border-color: var(--border-primary);">
			<div class="flex items-center gap-3">
				<button onclick={toggleSidebar} class="lg:hidden p-1.5 rounded-lg hover:bg-[var(--bg-hover)]" style="color: var(--text-secondary);" aria-label="Toggle sidebar">
					<svg width="20" height="20" viewBox="0 0 20 20" fill="none" aria-hidden="true"><path d="M3 5h14M3 10h14M3 15h14" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
				</button>
				<button onclick={() => commandOpen = true} class="btn btn-ghost btn-sm flex items-center gap-2" style="color: var(--text-tertiary)" aria-label="Open command palette">
					<span class="text-xs" aria-hidden="true">⌘K</span>
					<span class="text-xs hidden sm:inline">Search...</span>
				</button>
			</div>
			<div class="flex items-center gap-3">
				<button onclick={logout} class="btn btn-ghost btn-sm">Logout</button>
			</div>
		</header>

		<main class="flex-1 overflow-y-auto p-4 sm:p-6" id="main-content" role="main">
			{@render children()}
		</main>
	</div>
</div>

<CommandPalette open={commandOpen} onclose={() => commandOpen = false} />
