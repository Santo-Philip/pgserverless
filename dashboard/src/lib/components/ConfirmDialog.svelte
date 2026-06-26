<script lang="ts">
	let { open = false, title, description, confirmLabel = 'Confirm', variant = 'danger', onconfirm, oncancel, loading = false }: { open?: boolean; title: string; description?: string; confirmLabel?: string; variant?: 'danger' | 'primary'; onconfirm: () => void; oncancel?: () => void; loading?: boolean } = $props();

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
		if (e.key === 'Escape') {
			oncancel?.();
		}
	}
</script>

{#if open}
	<div class="modal-overlay" role="dialog" aria-modal="true" aria-label={title} tabindex="-1" bind:this={dialogRef} onkeydown={handleKeydown}>
		<div class="modal-content max-w-md" role="document">
			<div class="p-6">
				<h3 class="text-lg font-semibold mb-2">{title}</h3>
				{#if description}
					<p class="text-sm" style="color: var(--text-secondary)">{description}</p>
				{/if}
				<div class="flex justify-end gap-3 mt-6">
					<button onclick={() => oncancel?.()} class="btn btn-secondary">Cancel</button>
					<button onclick={onconfirm} disabled={loading} class="btn {variant === 'danger' ? 'btn-danger' : 'btn-primary'}" style="min-width: 80px;">
						{loading ? '...' : confirmLabel}
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}
