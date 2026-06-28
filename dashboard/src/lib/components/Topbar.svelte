<script lang="ts">
  import { goto } from '$app/navigation';
  import { env } from '$env/dynamic/public';
  import { theme, toggleTheme } from '$lib/stores/theme';

  const appName = env.PUBLIC_APP_NAME || env.PUBLIC_APP_NAME_SHORT || '';
  const logoLetter = appName.charAt(0).toUpperCase();

  let tapTimer = $state<ReturnType<typeof setTimeout> | null>(null);

  function handleThemeClick() {
    if (tapTimer) {
      clearTimeout(tapTimer);
      tapTimer = null;
      goto('/login?from=doubletap');
    } else {
      tapTimer = setTimeout(() => {
        tapTimer = null;
        toggleTheme();
      }, 300);
    }
  }
</script>

<header
  class="fixed top-0 left-0 right-0 z-40 flex items-center justify-between h-14 px-4 border-b"
  style="background-color: var(--bg-secondary); border-color: var(--border);"
>
  <div class="flex items-center gap-3">
    <div class="w-8 h-8 rounded-xl flex items-center justify-center text-sm font-bold shrink-0" style="background-color: var(--accent); color: #0f1117;">{logoLetter}</div>
    <span class="font-semibold text-sm">{appName}</span>
  </div>
  <button
    onclick={handleThemeClick}
    class="flex items-center justify-center w-8 h-8 rounded-lg transition-colors hover:bg-[var(--bg-hover)]"
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
</header>
