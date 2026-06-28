<script lang="ts">
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { isAuthenticated, logout } from '$lib/stores/auth';
  import { APP_NAME_SHORT, APP_LOGO_LETTER } from '$lib/config/brand';
  import { theme, toggleTheme } from '$lib/stores/theme';
  import type { Snippet } from 'svelte';

  let { children }: { children: Snippet } = $props();

  let collapsed = $state(false);
  let mobileOpen = $state(false);
  let lastTap = $state(0);

  function handleThemeClick() {
    const now = Date.now();
    if (now - lastTap < 300) {
      lastTap = 0;
      goto('/login');
    } else {
      lastTap = now;
      toggleTheme();
    }
  }

  type NavItem = { label: string; href: string; icon: string };

  const navItems: NavItem[] = [
    { label: 'Dashboard', href: '/dashboard', icon: '\u25C9' },
    { label: 'Explorer', href: '/explorer', icon: '\u25C8' },
    { label: 'SQL', href: '/sql', icon: '\u2328' },
    { label: 'Schema', href: '/schema', icon: '\u25CE' },
    { label: 'Roles', href: '/roles', icon: '\uD83D\uDC64' },
    { label: 'Extensions', href: '/extensions', icon: '\u25C6' },
    { label: 'Monitoring', href: '/monitoring', icon: '\u25CE' },
    { label: 'Backups', href: '/backups', icon: '\u25FC' },
    { label: 'Logs', href: '/logs', icon: '\u25C7' },
    { label: 'Audit', href: '/audit', icon: '\u269A' },
    { label: 'Admin', href: '/admin', icon: '\u2699' },
    { label: 'Settings', href: '/settings', icon: '\u26A1' },
  ];

  function isActive(href: string) {
    if (href === '/') return $page.url.pathname === '/';
    return $page.url.pathname.startsWith(href);
  }
</script>

<svelte:window onkeydown={(e) => { if (e.key === 'Escape') mobileOpen = false; }} />

<button
  class="fixed inset-0 z-30 bg-black/50 transition-opacity duration-200 md:hidden"
  class:opacity-100={mobileOpen}
  class:opacity-0={!mobileOpen}
  class:pointer-events-auto={mobileOpen}
  class:pointer-events-none={!mobileOpen}
  onclick={() => mobileOpen = false}
  aria-label="Close sidebar"
></button>

<aside
  class="sidebar"
  class:-translate-x-full={!mobileOpen}
  class:translate-x-0={mobileOpen}
  class:collapsed={collapsed}
  class:md:translate-x-0={true}
>
  <div class="flex items-center justify-between h-14 px-4 border-b shrink-0" style="border-color: var(--border);">
    {#if !collapsed}
      <div class="flex items-center gap-3 min-w-0">
        <div class="w-8 h-8 rounded-xl flex items-center justify-center text-sm font-bold shrink-0" style="background-color: var(--accent); color: #0f1117;">{APP_LOGO_LETTER}</div>
        <span class="font-semibold text-sm truncate">{APP_NAME_SHORT}</span>
      </div>
    {:else}
      <div class="w-8 h-8 rounded-xl flex items-center justify-center text-sm font-bold mx-auto shrink-0" style="background-color: var(--accent); color: #0f1117;">{APP_LOGO_LETTER}</div>
    {/if}
    <div class="flex items-center gap-1">
      <button
        onclick={handleThemeClick}
        class="flex items-center justify-center w-8 h-8 rounded-lg transition-colors shrink-0 hover:bg-[var(--bg-hover)]"
        style="color: var(--text-secondary);"
        aria-label="Toggle theme"
      >
        {#if $theme === 'dark'}
          <svg width="16" height="16" viewBox="0 0 20 20" fill="none" aria-hidden="true">
            <path d="M10 2a8 8 0 000 16 6 6 0 010-12 6 6 0 000-12z" fill="currentColor"/>
          </svg>
        {:else}
          <svg width="16" height="16" viewBox="0 0 20 20" fill="none" aria-hidden="true">
            <circle cx="10" cy="10" r="4" fill="currentColor"/>
            <path d="M10 2v2m0 12v2m-8-8h2m12 0h2M4.93 4.93l1.41 1.41m7.32 7.32l1.41 1.41M4.93 15.07l1.41-1.41m7.32-7.32l1.41-1.41" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
          </svg>
        {/if}
      </button>
      <button
        onclick={() => collapsed = !collapsed}
        class="flex items-center justify-center w-8 h-8 rounded-lg transition-colors shrink-0 hover:bg-[var(--bg-hover)]"
        style="color: var(--text-secondary);"
        aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
      >
        <svg width="16" height="16" viewBox="0 0 16 16" fill="none" aria-hidden="true" class:rotate-180={collapsed}>
          <path d="M10 12L6 8L10 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
      </button>
    </div>
  </div>

  <nav class="flex-1 overflow-y-auto p-2 space-y-0.5 scrollbar-hide">
    {#if $isAuthenticated}
      {#each navItems as item}
        <a
          href={item.href}
          class="flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-all duration-150 hover:bg-[var(--bg-hover)]"
          style="color: {isActive(item.href) ? 'var(--accent)' : 'var(--text-secondary)'}; background-color: {isActive(item.href) ? 'var(--accent-muted)' : 'transparent'};"
        >
          <span class="text-lg w-6 text-center shrink-0" aria-hidden="true">{item.icon}</span>
          {#if !collapsed}
            <span class="truncate">{item.label}</span>
          {/if}
        </a>
      {/each}
    {:else if !collapsed}
      <div class="px-3 py-4 text-center">
        <p class="text-xs" style="color: var(--text-secondary);">Sign in to access the admin dashboard.</p>
      </div>
    {/if}
  </nav>

  <div class="p-2 border-t shrink-0" style="border-color: var(--border);">
    {#if $isAuthenticated}
      <button
        onclick={logout}
        class="flex items-center gap-3 w-full px-3 py-2.5 rounded-xl text-sm font-medium transition-colors hover:bg-[var(--bg-hover)]"
        style="color: var(--text-secondary);"
      >
        <span class="text-lg w-6 text-center shrink-0">\u23FB</span>
        {#if !collapsed}
          <span>Logout</span>
        {/if}
      </button>
    {/if}
  </div>
</aside>

<div class="flex flex-col min-h-screen" class:md:ml-[var(--sidebar-width)]={!collapsed} class:md:ml-16={collapsed}>
  <header class="flex items-center justify-between h-14 px-4 md:px-6 border-b shrink-0 md:hidden" style="background-color: var(--bg-secondary); border-color: var(--border);">
    <button
      onclick={() => mobileOpen = true}
      class="flex items-center justify-center w-10 h-10 rounded-xl transition-colors hover:bg-[var(--bg-hover)]"
      style="color: var(--text-secondary);"
      aria-label="Open sidebar"
    >
      <svg width="22" height="22" viewBox="0 0 22 22" fill="none" aria-hidden="true"><path d="M3 6h16M3 11h16M3 16h16" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
    </button>
    <span class="font-semibold text-sm">{APP_NAME_SHORT}</span>
    <button
      onclick={handleThemeClick}
      class="flex items-center justify-center w-10 h-10 rounded-xl transition-colors hover:bg-[var(--bg-hover)]"
      style="color: var(--text-secondary);"
      aria-label="Toggle theme"
    >
      {#if $theme === 'dark'}
        <svg width="20" height="20" viewBox="0 0 20 20" fill="none" aria-hidden="true">
          <path d="M10 2a8 8 0 000 16 6 6 0 010-12 6 6 0 000-12z" fill="currentColor"/>
        </svg>
      {:else}
        <svg width="20" height="20" viewBox="0 0 20 20" fill="none" aria-hidden="true">
          <circle cx="10" cy="10" r="4" fill="currentColor"/>
          <path d="M10 2v2m0 12v2m-8-8h2m12 0h2M4.93 4.93l1.41 1.41m7.32 7.32l1.41 1.41M4.93 15.07l1.41-1.41m7.32-7.32l1.41-1.41" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
        </svg>
      {/if}
    </button>
  </header>

  <main class="flex-1 p-4 md:p-6 lg:p-8 max-w-7xl mx-auto w-full" id="main-content" role="main">
    {@render children()}
  </main>
</div>
