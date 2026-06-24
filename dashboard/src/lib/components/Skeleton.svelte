<script lang="ts">
	let { rows = 3 }: { rows?: number } = $props();
</script>

{#each Array(rows) as _}
	<div class="flex gap-4 p-4">
		<div class="skeleton h-4 w-1/4"></div>
		<div class="skeleton h-4 w-1/3"></div>
		<div class="skeleton h-4 w-1/6"></div>
		<div class="skeleton h-4 w-1/6"></div>
	</div>
{/each}
