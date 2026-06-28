<script lang="ts">
  let {
    title,
    value,
    subtitle,
    icon,
    trend,
    loading = false,
  }: {
    title: string;
    value: string | number;
    subtitle?: string;
    icon?: string;
    trend?: { value: string; positive: boolean };
    loading?: boolean;
  } = $props();
</script>

<div class="card p-5 transition-all duration-150 hover:border-[var(--accent)] hover:border-opacity-30">
  {#if loading}
    <div class="space-y-3">
      <div class="skeleton h-3 w-16"></div>
      <div class="skeleton h-7 w-24"></div>
      <div class="skeleton h-3 w-20"></div>
    </div>
  {:else}
    <div class="flex items-start justify-between mb-3">
      <span class="text-xs font-semibold uppercase tracking-wider" style="color: var(--text-secondary);">{title}</span>
      {#if icon}
        <span class="text-xl" aria-hidden="true">{icon}</span>
      {/if}
    </div>
    <div class="flex items-baseline gap-2.5">
      <span class="text-3xl font-bold tracking-tight" style="color: var(--text-primary);">{value}</span>
      {#if trend}
        <span
          class="text-xs font-semibold px-1.5 py-0.5 rounded-lg"
          style="background-color: {trend.positive ? 'rgba(102, 187, 106, 0.12)' : 'rgba(239, 83, 80, 0.12)'}; color: {trend.positive ? 'var(--success)' : 'var(--danger)'};"
        >
          {trend.value}
        </span>
      {/if}
    </div>
    {#if subtitle}
      <p class="text-xs mt-1.5" style="color: var(--text-secondary);">{subtitle}</p>
    {/if}
  {/if}
</div>
