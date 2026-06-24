<script lang="ts">
	let { message = '', type = 'info' }: { message?: string; type?: 'info' | 'success' | 'error' } = $props();

	let colors: Record<string, string> = {
		info: 'var(--accent)',
		success: 'var(--success)',
		error: 'var(--danger)',
	};

	let bgColors: Record<string, string> = {
		info: 'rgba(12, 142, 229, 0.1)',
		success: 'rgba(34, 197, 94, 0.1)',
		error: 'rgba(239, 68, 68, 0.1)',
	};
</script>

{#if message}
	<div class="flex items-center gap-2 px-4 py-3 rounded-lg text-sm" style="background-color: {bgColors[type]}; color: {colors[type]};">
		{message}
	</div>
{/if}
