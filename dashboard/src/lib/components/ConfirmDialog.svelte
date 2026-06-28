<script lang="ts">
  import type { Snippet } from 'svelte';

  let {
    open = false,
    title,
    description,
    confirmLabel = 'Confirm',
    variant = 'danger' as 'danger' | 'primary',
    onconfirm,
    oncancel,
    loading = false,
    children,
  }: {
    open?: boolean;
    title: string;
    description?: string;
    confirmLabel?: string;
    variant?: 'danger' | 'primary';
    onconfirm: () => void;
    oncancel?: () => void;
    loading?: boolean;
    children?: Snippet;
  } = $props();

  let dialogRef: HTMLDivElement | undefined = $state(undefined);
  let previousFocus: HTMLElement | null = null;

  $effect(() => {
    if (open) {
      previousFocus = document.activeElement as HTMLElement;
      requestAnimationFrame(() => {
        dialogRef?.focus();
      });
    } else if (previousFocus) {
      previousFocus.focus();
      previousFocus = null;
    }
  });

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') oncancel?.();
  }
</script>

{#if open}
  <div
    class="modal-overlay"
    role="dialog"
    aria-modal="true"
    aria-label={title}
    tabindex="-1"
    bind:this={dialogRef}
    onkeydown={handleKeydown}
    onclick={(e) => { if (e.target === e.currentTarget) oncancel?.(); }}
  >
    <div class="modal-content max-w-md" role="document">
      <div class="p-6">
        <div class="flex items-start gap-4">
          <div
            class="flex items-center justify-center w-12 h-12 rounded-xl shrink-0"
            style="background-color: {variant === 'danger' ? 'rgba(239, 83, 80, 0.12)' : 'var(--accent-muted)'};"
          >
            {#if variant === 'danger'}
              <svg width="22" height="22" viewBox="0 0 22 22" fill="none" aria-hidden="true" style="color: var(--danger);"><path d="M11 7v4m0 4h.01M3.5 11a7.5 7.5 0 1115 0 7.5 7.5 0 01-15 0z" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
            {:else}
              <svg width="22" height="22" viewBox="0 0 22 22" fill="none" aria-hidden="true" style="color: var(--accent);"><path d="M7 11l3 3 5-5M3.5 11a7.5 7.5 0 1115 0 7.5 7.5 0 01-15 0z" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
            {/if}
          </div>
          <div class="flex-1 min-w-0">
            <h3 class="text-lg font-semibold mb-1" style="color: var(--text-primary);">{title}</h3>
            {#if description}
              <p class="text-sm leading-relaxed" style="color: var(--text-secondary);">{description}</p>
            {/if}
            {#if children}
              <div class="mt-4">{@render children()}</div>
            {/if}
          </div>
        </div>
        <div class="flex justify-end gap-3 mt-6">
          <button
            onclick={() => oncancel?.()}
            class="btn btn-secondary"
            disabled={loading}
          >
            Cancel
          </button>
          <button
            onclick={onconfirm}
            disabled={loading}
            class="btn {variant === 'danger' ? 'btn-danger' : 'btn-primary'}"
            style="min-width: 90px;"
          >
            {#if loading}
              <svg class="animate-spin h-4 w-4" viewBox="0 0 24 24" fill="none" aria-hidden="true"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
            {:else}
              {confirmLabel}
            {/if}
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}
