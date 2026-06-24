<script lang="ts">
	let { href, icon, children, active = false }: { href: string; icon?: string; children: import('svelte').Snippet; active?: boolean } = $props();
</script>

<a {href} class="nav-item {active ? 'active' : ''}">
	{#if icon}<span class="nav-icon">{icon}</span>{/if}
	<span>{@render children()}</span>
</a>

<style>
	.nav-item {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.5rem 0.75rem;
		border-radius: 0.5rem;
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--text-secondary);
		text-decoration: none;
		transition: all 0.15s;
		cursor: pointer;
	}
	.nav-item:hover {
		background-color: var(--bg-hover);
		color: var(--text-primary);
	}
	.nav-item.active {
		background-color: var(--accent-muted);
		color: var(--accent);
	}
	.nav-icon {
		font-size: 1.125rem;
		width: 1.25rem;
		text-align: center;
		flex-shrink: 0;
	}
</style>
