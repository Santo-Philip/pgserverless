<script lang="ts">
	let { columns, rows, onrowclick }: { columns: { key: string; label: string; width?: string; render?: (value: unknown, row: Record<string, unknown>) => import('svelte').Snippet }[]; rows: Record<string, unknown>[]; onrowclick?: (row: Record<string, unknown>) => void } = $props();
</script>

<div class="table-wrap overflow-x-auto">
	<table class="w-full">
		<thead>
			<tr>
				{#each columns as col}
					<th style={col.width ? `width: ${col.width}` : ''}>{col.label}</th>
				{/each}
			</tr>
		</thead>
		<tbody>
			{#each rows as row}
				<tr
					class={onrowclick ? 'cursor-pointer' : ''}
					onclick={() => onrowclick?.(row)}
					onkeydown={(e) => e.key === 'Enter' && onrowclick?.(row)}
					tabindex={onrowclick ? 0 : -1}
				>
					{#each columns as col}
						<td class="truncate max-w-xs">
							{col.render ? col.render(row[col.key], row) : row[col.key]}
						</td>
					{/each}
				</tr>
			{/each}
		</tbody>
	</table>
</div>
