<script lang="ts">
  import type { Snippet } from 'svelte';

  let {
    title,
    subtitle,
    actions,
    padding = true,
    children,
  }: {
    title?: string;
    subtitle?: string;
    actions?: Snippet;
    padding?: boolean;
    children: Snippet;
  } = $props();
</script>

<div class="card overflow-hidden">
  {#if title || actions}
    <div class="flex items-center justify-between gap-4 px-5 py-4 border-b" style="border-color: var(--border);">
      <div class="min-w-0">
        {#if title}
          <h3 class="text-sm font-semibold" style="color: var(--text-primary);">{title}</h3>
        {/if}
        {#if subtitle}
          <p class="text-xs mt-0.5" style="color: var(--text-secondary);">{subtitle}</p>
        {/if}
      </div>
      {#if actions}
        <div class="flex items-center gap-2 shrink-0">
          {@render actions()}
        </div>
      {/if}
    </div>
  {/if}
  <div class:px-5={padding} class:py-4={padding}>
    {@render children()}
  </div>
</div>
