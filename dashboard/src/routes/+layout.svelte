<script lang="ts">
  import '../app.css';
  import Topbar from '$lib/components/Topbar.svelte';
  import BottomNav from '$lib/components/BottomNav.svelte';
  import Toast from '$lib/components/Toast.svelte';
  import { isAuthenticated } from '$lib/stores/auth';
  import { page } from '$app/stores';
  import { navigating } from '$app/stores';
  import { goto } from '$app/navigation';

  let { children } = $props();
  let authChecked = $state(false);

  const publicPaths = ['/login', '/'];

  $effect(() => {
    if (!authChecked) {
      if (!$isAuthenticated && !publicPaths.includes($page.url.pathname)) {
        sessionStorage.setItem('login_origin', 'theme_doubletap');
        goto('/login?from=doubletap');
      }
      authChecked = true;
    }
  });
</script>

{#if $navigating}
  <div class="fixed top-0 left-0 right-0 z-[100] h-0.5" style="background-color: var(--accent);">
    <div class="h-full w-full origin-left" style="background-color: var(--accent); animation: progress 30s ease-in-out infinite;"></div>
  </div>
{/if}

{#if $isAuthenticated || $page.url.pathname === '/'}
  <Topbar />
  <div class="pt-14 min-h-screen" style="background-color: var(--bg-primary);">
    <div class="pb-16 lg:pb-0">
      {@render children()}
    </div>
  </div>
  {#if $isAuthenticated}
    <BottomNav />
  {/if}
  <Toast />
{:else if $page.url.pathname.startsWith('/login')}
  {@render children()}
  <Toast />
{/if}

<style>
  @keyframes progress {
    0% { transform: scaleX(0); }
    50% { transform: scaleX(0.5); }
    100% { transform: scaleX(0.8); }
  }
</style>
