<script lang="ts">
  import { toast, type ToastType } from '$lib/stores/toast';

  $effect(() => {
    toast.subscribe(() => {});
  });

  function getIcon(type: ToastType): string {
    switch (type) {
      case 'success': return '\u2713';
      case 'error': return '\u2717';
      case 'warning': return '\u26A0';
      case 'info': return '\u2139';
    }
  }

  function getColors(type: ToastType) {
    switch (type) {
      case 'success': return { bg: 'rgba(102, 187, 106, 0.15)', border: 'var(--success)', text: 'var(--success)' };
      case 'error': return { bg: 'rgba(239, 83, 80, 0.15)', border: 'var(--danger)', text: 'var(--danger)' };
      case 'warning': return { bg: 'rgba(255, 167, 38, 0.15)', border: 'var(--warning)', text: 'var(--warning)' };
      case 'info': return { bg: 'var(--accent-muted)', border: 'var(--accent)', text: 'var(--accent)' };
    }
  }
</script>

<svelte:window />

<div
  class="fixed top-4 right-4 z-[100] flex flex-col gap-2 pointer-events-none md:max-w-sm w-full max-w-[calc(100vw-2rem)]"
  role="log"
  aria-live="polite"
  aria-label="Notifications"
>
  {#each $toast as t (t.id)}
    {@const colors = getColors(t.type)}
    <div
      class="flex items-start gap-3 px-4 py-3.5 rounded-xl shadow-lg pointer-events-auto transition-all duration-300"
      style="background-color: {colors.bg}; border: 1px solid {colors.border}; color: {colors.text};"
      role="alert"
    >
      <span class="text-lg mt-0.5 shrink-0" aria-hidden="true">{getIcon(t.type)}</span>
      <p class="text-sm flex-1 min-w-0" style="color: var(--text-primary);">{t.message}</p>
      <button
        onclick={() => toast.remove(t.id)}
        class="flex items-center justify-center w-7 h-7 rounded-lg shrink-0 transition-colors hover:bg-[var(--bg-hover)]"
        style="color: var(--text-secondary);"
        aria-label="Dismiss notification"
      >
        &times;
      </button>
    </div>
  {/each}
</div>
