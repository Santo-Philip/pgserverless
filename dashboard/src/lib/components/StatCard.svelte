<script lang="ts">
	let { title, value, subtitle, icon, trend }: { title: string; value: string | number; subtitle?: string; icon?: string; trend?: { value: string; positive: boolean } } = $props();
</script>

<div class="card p-5 card-hover">
	<div class="flex items-start justify-between mb-3">
		<div class="flex items-center gap-2">
			{#if icon}<span class="text-lg">{icon}</span>{/if}
			<span class="text-xs font-medium uppercase tracking-wider" style="color: var(--text-tertiary)">{title}</span>
		</div>
	</div>
	<div class="flex items-baseline gap-2">
		<span class="text-2xl font-bold" style="color: var(--text-primary)">{value}</span>
		{#if trend}
			<span class="text-xs font-medium" class:positive={trend.positive} style="color: {trend.positive ? 'var(--success)' : 'var(--danger)'}">
				{trend.value}
			</span>
		{/if}
	</div>
	{#if subtitle}
		<p class="text-xs mt-1" style="color: var(--text-tertiary)">{subtitle}</p>
	{/if}
</div>
