<script lang="ts">
  import type { Snippet } from 'svelte';

  let {
    variant = 'default' as 'default' | 'success' | 'danger' | 'warning' | 'accent',
    size = 'md' as 'sm' | 'md' | 'lg',
    children,
  }: {
    variant?: string;
    size?: 'sm' | 'md' | 'lg';
    children?: Snippet;
  } = $props();

  function getColors() {
    switch (variant) {
      case 'success': return { bg: 'rgba(102, 187, 106, 0.12)', text: 'var(--success)' };
      case 'danger': return { bg: 'rgba(239, 83, 80, 0.12)', text: 'var(--danger)' };
      case 'warning': return { bg: 'rgba(255, 167, 38, 0.12)', text: 'var(--warning)' };
      case 'accent': return { bg: 'var(--accent-muted)', text: 'var(--accent)' };
      default: return { bg: 'var(--bg-hover)', text: 'var(--text-secondary)' };
    }
  }

  let colors = $derived(getColors());

  function sizeClasses() {
    switch (size) {
      case 'sm': return 'px-1.5 py-0.5 text-[10px]';
      case 'lg': return 'px-3 py-1 text-sm';
      default: return 'px-2.5 py-0.5 text-xs';
    }
  }
</script>

<span class="badge {sizeClasses()}" style="background-color: {colors.bg}; color: {colors.text};">
  {#if children}{@render children()}{/if}
</span>
