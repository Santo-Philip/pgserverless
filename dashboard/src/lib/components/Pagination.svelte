<script lang="ts">
	let { total, limit, offset, onpage }: { total: number; limit: number; offset: number; onpage: (offset: number) => void } = $props();

	let totalPages = $derived(Math.max(1, Math.ceil(total / limit)));
	let currentPage = $derived(Math.floor(offset / limit) + 1);
</script>

{#if totalPages > 1}
	<div class="flex items-center justify-between px-4 py-3 border-t" style="border-color: var(--border-primary);">
		<span class="text-xs" style="color: var(--text-tertiary)">
			{offset + 1}–{Math.min(offset + limit, total)} of {total}
		</span>
		<div class="flex gap-1">
			<button
				class="btn btn-ghost btn-sm"
				disabled={currentPage <= 1}
				onclick={() => onpage((currentPage - 2) * limit)}
			>
				Prev
			</button>
			<button
				class="btn btn-ghost btn-sm"
				disabled={currentPage >= totalPages}
				onclick={() => onpage(currentPage * limit)}
			>
				Next
			</button>
		</div>
	</div>
{/if}
