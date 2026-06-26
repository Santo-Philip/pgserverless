<script lang="ts">
	let { open = false, title, onclose, children }: { open?: boolean; title?: string; onclose?: () => void; children: import('svelte').Snippet } = $props();

	let modalRef: HTMLDivElement | undefined = $state(undefined);
	let previousFocus: HTMLElement | null = null;

	$effect(() => {
		if (open) {
			previousFocus = document.activeElement as HTMLElement;
			requestAnimationFrame(() => {
				modalRef?.focus();
			});
		} else if (previousFocus) {
			previousFocus.focus();
			previousFocus = null;
		}
	});

	function handleOverlayClick(e: MouseEvent) {
		if (e.target === e.currentTarget) {
			onclose?.();
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			onclose?.();
			return;
		}
		if (e.key === 'Tab' && modalRef) {
			const focusable = modalRef.querySelectorAll<HTMLElement>(
				'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
			);
			if (focusable.length === 0) return;
			const first = focusable[0];
			const last = focusable[focusable.length - 1];
			if (e.shiftKey && document.activeElement === first) {
				e.preventDefault();
				last.focus();
			} else if (!e.shiftKey && document.activeElement === last) {
				e.preventDefault();
				first.focus();
			}
		}
	}
</script>

{#if open}
	<div class="modal-overlay" onclick={handleOverlayClick} onkeydown={handleKeydown} role="dialog" aria-modal="true" aria-label={title || 'Dialog'} tabindex="-1" bind:this={modalRef}>
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
