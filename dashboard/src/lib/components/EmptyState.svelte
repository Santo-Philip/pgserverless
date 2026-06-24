<script lang="ts">
	let { icon, title, description, children }: { icon?: string; title: string; description?: string; children?: import('svelte').Snippet } = $props();
</script>

<div class="card p-12 text-center">
	{#if icon}<div class="text-3xl mb-3">{icon}</div>{/if}
	<h3 class="text-base font-semibold mb-1" style="color: var(--text-secondary)">{title}</h3>
	{#if description}
		<p class="text-sm mb-4" style="color: var(--text-tertiary)">{description}</p>
	{/if}
	{#if children}
		{@render children()}
	{/if}
</div>
