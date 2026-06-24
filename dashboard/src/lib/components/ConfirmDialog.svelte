<script lang="ts">
	let { open = false, title, description, confirmLabel = 'Confirm', variant = 'danger', onconfirm, oncancel, loading = false }: { open?: boolean; title: string; description?: string; confirmLabel?: string; variant?: 'danger' | 'primary'; onconfirm: () => void; oncancel?: () => void; loading?: boolean } = $props();
</script>

{#if open}
	<div class="modal-overlay" role="dialog" tabindex="-1">
		<div class="modal-content max-w-md">
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
