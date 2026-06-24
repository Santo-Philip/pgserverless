<script lang="ts">
	let { open = false, title, onclose, children }: { open?: boolean; title?: string; onclose?: () => void; children: import('svelte').Snippet } = $props();

	function handleOverlayClick(e: MouseEvent) {
		if (e.target === e.currentTarget) {
			onclose?.();
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') onclose?.();
	}
</script>

{#if open}
	<div class="modal-overlay" onclick={handleOverlayClick} onkeydown={handleKeydown} role="dialog" tabindex="-1">
		<div class="modal-content">
			{#if title}
				<div class="flex items-center justify-between px-6 py-4 border-b" style="border-color: var(--border-primary);">
					<h2 class="text-lg font-semibold">{title}</h2>
					<button onclick={() => onclose?.()} class="btn btn-ghost btn-sm" aria-label="Close">&times;</button>
				</div>
			{/if}
			<div class="p-6">
				{@render children()}
			</div>
		</div>
	</div>
{/if}
