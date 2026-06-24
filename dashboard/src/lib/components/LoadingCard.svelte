<script lang="ts">
	let { icon = true }: { icon?: boolean } = $props();
</script>

<div class="card p-5">
	<div class="flex items-center gap-3">
		{#if icon}
			<div class="skeleton w-10 h-10 rounded-lg"></div>
		{/if}
		<div class="flex-1 space-y-2">
			<div class="skeleton h-4 w-1/3"></div>
			<div class="skeleton h-3 w-1/2"></div>
		</div>
	</div>
</div>
