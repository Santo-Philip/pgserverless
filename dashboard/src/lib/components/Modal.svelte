<script lang="ts">
  import type { Snippet } from 'svelte';

  let {
    open = false,
    title,
    onclose,
    wide = false,
    children,
    footer,
  }: {
    open?: boolean;
    title?: string;
    onclose?: () => void;
    wide?: boolean;
    children: Snippet;
    footer?: Snippet;
  } = $props();

  let modalRef: HTMLDivElement | undefined = $state(undefined);
  let previousFocus: HTMLElement | null = null;

  $effect(() => {
    if (open) {
      previousFocus = document.activeElement as HTMLElement;
      requestAnimationFrame(() => {
        modalRef?.focus();
      });
    } else if (previousFocus) {
      previousFocus.focus();
      previousFocus = null;
    }
  });

  function handleOverlayClick(e: MouseEvent) {
    if (e.target === e.currentTarget) {
      onclose?.();
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      onclose?.();
      return;
    }
    if (e.key === 'Tab' && modalRef) {
      const focusable = modalRef.querySelectorAll<HTMLElement>(
        'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
      );
      if (focusable.length === 0) return;
      const first = focusable[0];
      const last = focusable[focusable.length - 1];
      if (e.shiftKey && document.activeElement === first) {
        e.preventDefault();
        last.focus();
      } else if (!e.shiftKey && document.activeElement === last) {
        e.preventDefault();
        first.focus();
      }
    }
  }
</script>

{#if open}
  <div
    class="modal-overlay"
    onclick={handleOverlayClick}
    onkeydown={handleKeydown}
    role="dialog"
    aria-modal="true"
    aria-label={title || 'Dialog'}
    tabindex="-1"
    bind:this={modalRef}
  >
    <div class="modal-content {wide ? 'max-w-2xl' : 'max-w-lg'}" role="document">
      {#if title}
        <div class="flex items-center justify-between px-6 py-4 border-b shrink-0" style="border-color: var(--border);">
          <h2 class="text-lg font-semibold" style="color: var(--text-primary);">{title}</h2>
          <button
            onclick={() => onclose?.()}
            class="flex items-center justify-center w-10 h-10 rounded-xl transition-colors hover:bg-[var(--bg-hover)]"
            style="color: var(--text-secondary);"
            aria-label="Close dialog"
          >
            <svg width="18" height="18" viewBox="0 0 18 18" fill="none" aria-hidden="true"><path d="M14 4L4 14M4 4l10 10" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
          </button>
        </div>
      {/if}
      <div class="p-6">
        {@render children()}
      </div>
      {#if footer}
        <div class="flex items-center justify-end gap-3 px-6 py-4 border-t shrink-0" style="border-color: var(--border);">
          {@render footer()}
        </div>
      {/if}
    </div>
  </div>
{/if}
