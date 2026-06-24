<script lang="ts">
	import type { Snippet } from 'svelte';

	let { title, description, children }: { title?: string; description?: string; children: Snippet } = $props();
</script>

<div class="card p-5">
	{#if title}
		<div class="mb-4">
			<h3 class="text-sm font-semibold" style="color: var(--text-primary)">{title}</h3>
			{#if description}
				<p class="text-xs mt-0.5" style="color: var(--text-tertiary)">{description}</p>
			{/if}
		</div>
	{/if}
	{@render children()}
</div>
