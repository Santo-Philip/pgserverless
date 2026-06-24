<script lang="ts">
	import { page } from '$app/stores';

	let { tabs }: { tabs: { label: string; href: string }[] } = $props();
</script>

<div class="flex gap-0 border-b" style="border-color: var(--border-primary); overflow-x: auto;">
	{#each tabs as tab}
		<a
			href={tab.href}
			class="tab {$page.url.pathname === tab.href ? 'active' : ''}"
		>
			{tab.label}
		</a>
	{/each}
</div>
