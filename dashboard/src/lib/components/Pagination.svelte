<script lang="ts">
  let {
    total,
    limit,
    offset,
    onpage,
  }: {
    total: number;
    limit: number;
    offset: number;
    onpage: (offset: number) => void;
  } = $props();

  let totalPages = $derived(Math.max(1, Math.ceil(total / limit)));
  let currentPage = $derived(Math.floor(offset / limit) + 1);

  let pages = $derived.by(() => {
    const result: (number | '...')[] = [];
    const delta = 1;
    const rangeStart = Math.max(2, currentPage - delta);
    const rangeEnd = Math.min(totalPages - 1, currentPage + delta);

    result.push(1);
    if (rangeStart > 2) result.push('...');
    for (let i = rangeStart; i <= rangeEnd; i++) result.push(i);
    if (rangeEnd < totalPages - 1) result.push('...');
    if (totalPages > 1) result.push(totalPages);
    return result;
  });
</script>

{#if totalPages > 1}
  <div class="flex items-center justify-between gap-4 px-1 py-3">
    <span class="text-xs whitespace-nowrap" style="color: var(--text-secondary);">
      {offset + 1}\u2013{Math.min(offset + limit, total)} of {total}
    </span>
    <div class="flex items-center gap-1">
      <button
        class="btn btn-ghost btn-sm"
        disabled={currentPage <= 1}
        onclick={() => onpage((currentPage - 2) * limit)}
        aria-label="Previous page"
      >
        <svg width="16" height="16" viewBox="0 0 16 16" fill="none" aria-hidden="true"><path d="M10 12L6 8L10 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
      </button>
      {#each pages as p}
        {#if p === '...'}
          <span class="px-2 text-xs" style="color: var(--text-secondary);">\u2026</span>
        {:else}
          <button
            class="flex items-center justify-center min-w-[36px] h-9 px-2 rounded-xl text-sm font-medium transition-all duration-150 hover:bg-[var(--bg-hover)]"
            style="background-color: {p === currentPage ? 'var(--accent-muted)' : 'transparent'}; color: {p === currentPage ? 'var(--accent)' : 'var(--text-secondary)'};"
            onclick={() => onpage((p - 1) * limit)}
            aria-label="Page {p}"
            aria-current={p === currentPage ? 'page' : undefined}
          >
            {p}
          </button>
        {/if}
      {/each}
      <button
        class="btn btn-ghost btn-sm"
        disabled={currentPage >= totalPages}
        onclick={() => onpage(currentPage * limit)}
        aria-label="Next page"
      >
        <svg width="16" height="16" viewBox="0 0 16 16" fill="none" aria-hidden="true"><path d="M6 4l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
      </button>
    </div>
  </div>
{/if}
