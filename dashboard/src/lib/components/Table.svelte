<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Column {
    key: string;
    label: string;
    width?: string;
    sortable?: boolean;
    render?: (value: unknown, row: Record<string, unknown>) => Snippet;
  }

  let {
    columns,
    rows,
    sortKey = $bindable(''),
    sortOrder = $bindable<'asc' | 'desc'>('asc'),
    onSort,
    onRowClick,
    emptyMessage = 'No data',
    loading = false,
  }: {
    columns: Column[];
    rows: Record<string, unknown>[];
    sortKey?: string;
    sortOrder?: 'asc' | 'desc';
    onSort?: (key: string) => void;
    onRowClick?: (row: Record<string, unknown>) => void;
    emptyMessage?: string;
    loading?: boolean;
  } = $props();

  function handleSort(key: string) {
    if (!onSort) return;
    onSort(key);
  }

  function sortIcon(key: string) {
    if (sortKey !== key) return '';
    return sortOrder === 'asc' ? ' \u25B2' : ' \u25BC';
  }
</script>

<div class="table-wrap overflow-x-auto rounded-xl border" style="border-color: var(--border);">
  <table class="w-full">
    <thead>
      <tr>
        {#each columns as col}
          <th
            class:cursor-pointer={col.sortable}
            style={col.width ? `width: ${col.width}` : ''}
            onclick={() => col.sortable && handleSort(col.key)}
          >
            {col.label}{col.sortable ? sortIcon(col.key) : ''}
          </th>
        {/each}
      </tr>
    </thead>
    <tbody>
      {#if loading}
        {#each Array(5) as _}
          <tr>
            {#each columns as col}
              <td>
                <div class="skeleton h-4 w-3/4"></div>
              </td>
            {/each}
          </tr>
        {/each}
      {:else if rows.length === 0}
        <tr>
          <td colspan={columns.length} class="text-center py-10 text-sm" style="color: var(--text-secondary);">
            {emptyMessage}
          </td>
        </tr>
      {:else}
        {#each rows as row}
          <tr
            class:cursor-pointer={!!onRowClick}
            class:hover:bg-[var(--bg-hover)]={true}
            onclick={() => onRowClick?.(row)}
            onkeydown={(e) => e.key === 'Enter' && onRowClick?.(row)}
            tabindex={onRowClick ? 0 : -1}
          >
            {#each columns as col}
              <td>
                {#if col.render}
                  {col.render(row[col.key], row)}
                {:else}
                  <span class="truncate block max-w-xs">{String(row[col.key] ?? '')}</span>
                {/if}
              </td>
            {/each}
          </tr>
        {/each}
      {/if}
    </tbody>
  </table>
</div>
