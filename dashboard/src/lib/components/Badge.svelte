<script lang="ts">
	let { status }: { status: string } = $props();

	function color() {
		switch (status) {
			case 'active': return 'var(--success)';
			case 'suspended':
			case 'inactive':
			case 'deleted':
			case 'failed': return 'var(--danger)';
			case 'pending':
			case 'running': return 'var(--warning)';
			default: return 'var(--text-tertiary)';
		}
	}

	function bg() {
		switch (status) {
			case 'active': return 'rgba(34, 197, 94, 0.1)';
			case 'suspended':
			case 'deleted':
			case 'failed': return 'rgba(239, 68, 68, 0.1)';
			case 'inactive':
			case 'pending':
			case 'running': return 'rgba(245, 158, 11, 0.1)';
			default: return 'rgba(107, 114, 128, 0.1)';
		}
	}
</script>

<span class="badge" style="background-color: {bg()}; color: {color()}">{status}</span>
