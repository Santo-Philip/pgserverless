<script lang="ts">
	import type { Snippet } from 'svelte';

	let { href, icon, children, active = false }: { href: string; icon?: string; children: Snippet; active?: boolean } = $props();
</script>

<a
	{href}
	class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors no-underline"
	class:active
	aria-current={active ? 'page' : undefined}
	style="color: {active ? 'var(--accent)' : 'var(--text-secondary)'}; background-color: {active ? 'var(--accent-muted)' : 'transparent'};"
	onmouseenter={() => {}}
	onmouseleave={() => {}}
>
	{#if icon}
		<span class="w-5 h-5 flex items-center justify-center text-base" aria-hidden="true">{icon}</span>
	{/if}
	{@render children()}
</a>
