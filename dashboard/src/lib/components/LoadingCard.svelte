<script lang="ts">
  import type { Snippet } from 'svelte';

  let {
    rows = 3,
    children,
  }: {
    rows?: number;
    children?: Snippet;
  } = $props();
</script>

<div class="card p-5">
  {#if children}
    {@render children()}
  {:else}
    {#each Array(rows) as _}
      <div class="skeleton h-4 w-full mb-3 rounded-lg"></div>
    {/each}
  {/if}
</div>
