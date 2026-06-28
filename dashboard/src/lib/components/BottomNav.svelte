<script lang="ts">
  import { page } from '$app/stores';

  type NavItem = { label: string; href: string; icon: string };

  const items: NavItem[] = [
    { label: 'Dashboard', href: '/dashboard', icon: '\u25C9' },
    { label: 'Explorer', href: '/explorer', icon: '\u25C8' },
    { label: 'SQL', href: '/sql', icon: '\u2328' },
    { label: 'Monitoring', href: '/monitoring', icon: '\u25CE' },
    { label: 'More', href: '#more', icon: '\u22EF' },
  ];

  let moreOpen = $state(false);

  const moreItems: NavItem[] = [
    { label: 'Schema', href: '/schema', icon: '\u25CE' },
    { label: 'Roles', href: '/roles', icon: '\uD83D\uDC64' },
    { label: 'Extensions', href: '/extensions', icon: '\u25C6' },
    { label: 'Backups', href: '/backups', icon: '\u25FC' },
    { label: 'Logs', href: '/logs', icon: '\u25C7' },
    { label: 'Audit', href: '/audit', icon: '\u269A' },
    { label: 'Admin', href: '/admin', icon: '\u2699' },
    { label: 'Settings', href: '/settings', icon: '\u26A1' },
  ];

  function isActive(href: string) {
    if (href === '/') return $page.url.pathname === '/';
    if (href === '#more') return false;
    return $page.url.pathname.startsWith(href);
  }
</script>

<nav class="bottom-nav md:hidden" aria-label="Bottom navigation">
  {#each items as item}
    {#if item.href === '#more'}
      <button
        onclick={() => moreOpen = !moreOpen}
        class="flex flex-col items-center justify-center gap-0.5 transition-colors relative"
        style="color: {moreOpen ? 'var(--accent)' : 'var(--text-secondary)'}; min-width: 56px; min-height: 48px;"
        aria-label="More navigation"
      >
        <span class="text-xl">{item.icon}</span>
        <span class="text-[10px] font-medium">{item.label}</span>
      </button>
    {:else}
      <a
        href={item.href}
        class="flex flex-col items-center justify-center gap-0.5 transition-colors relative"
        style="color: {isActive(item.href) ? 'var(--accent)' : 'var(--text-secondary)'}; min-width: 56px; min-height: 48px;"
      >
        {#if isActive(item.href)}
          <span class="absolute -top-0.5 w-5 h-0.5 rounded-full" style="background-color: var(--accent)"></span>
        {/if}
        <span class="text-xl">{item.icon}</span>
        <span class="text-[10px] font-medium">{item.label}</span>
      </a>
    {/if}
  {/each}
</nav>

{#if moreOpen}
  <div class="fixed inset-0 z-30 md:hidden" style="margin-bottom: var(--bottomnav-height);" onclick={() => moreOpen = false}>
    <div class="absolute bottom-0 left-0 right-0 p-4 pb-2" style="background-color: var(--bg-secondary); border-top: 1px solid var(--border); border-radius: 20px 20px 0 0;" onclick={(e) => e.stopPropagation()}>
      <div class="flex items-center justify-between mb-3 px-1">
        <span class="text-sm font-semibold" style="color: var(--text-primary);">More</span>
        <button onclick={() => moreOpen = false} class="w-8 h-8 rounded-lg flex items-center justify-center" style="color: var(--text-secondary);">&times;</button>
      </div>
      <div class="grid grid-cols-4 gap-1">
        {#each moreItems as item}
          <a
            href={item.href}
            onclick={() => moreOpen = false}
            class="flex flex-col items-center gap-1 py-3 px-2 rounded-xl transition-colors hover:bg-[var(--bg-hover)]"
            style="color: {isActive(item.href) ? 'var(--accent)' : 'var(--text-secondary)'};"
          >
            <span class="text-lg">{item.icon}</span>
            <span class="text-[10px] font-medium text-center">{item.label}</span>
          </a>
        {/each}
      </div>
    </div>
  </div>
{/if}
