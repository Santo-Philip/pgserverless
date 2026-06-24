<script lang="ts">
	import { page } from '$app/stores';

	let { items }: { items: { label: string; href?: string }[] } = $props();

	let path = $derived($page.url.pathname);
</script>

<nav class="flex items-center gap-2 text-sm mb-6" style="color: var(--text-tertiary)">
	{#each items as item, i}
		{#if i > 0}<span>/</span>{/if}
		{#if item.href}
			<a href={item.href} class="link">{item.label}</a>
		{:else}
			<span>{item.label}</span>
		{/if}
	{/each}
</nav>
