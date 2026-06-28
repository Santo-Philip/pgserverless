<script lang="ts">
  let {
    rows = 3,
    type = 'text',
  }: {
    rows?: number;
    type?: 'text' | 'card' | 'table-row';
  } = $props();
</script>

{#if type === 'card'}
  {#each Array(rows) as _}
    <div class="card p-5 space-y-3">
      <div class="skeleton h-3 w-1/3"></div>
      <div class="skeleton h-7 w-1/2"></div>
      <div class="skeleton h-3 w-2/3"></div>
    </div>
  {/each}
{:else if type === 'table-row'}
  {#each Array(rows) as _}
    <div class="flex gap-4 px-4 py-3.5 border-b items-center" style="border-color: var(--border);">
      <div class="skeleton h-4 w-1/4"></div>
      <div class="skeleton h-4 w-1/3"></div>
      <div class="skeleton h-4 w-1/6"></div>
      <div class="skeleton h-4 w-1/6"></div>
      <div class="skeleton h-4 w-12 ml-auto"></div>
    </div>
  {/each}
{:else}
  <div class="space-y-3">
    {#each Array(rows) as _}
      <div class="flex gap-4 items-center">
        <div class="skeleton h-4 w-1/4"></div>
        <div class="skeleton h-4 w-1/3"></div>
        <div class="skeleton h-4 w-1/6"></div>
      </div>
    {/each}
  </div>
{/if}
