<script lang="ts">
  let {
    tabs,
    active,
    onchange,
  }: {
    tabs: { value: string; label: string }[];
    active: string;
    onchange: (value: string) => void;
  } = $props();
</script>

<div class="flex border-b overflow-x-auto scrollbar-hide" style="border-color: var(--border);">
  {#each tabs as tab}
    <button
      class="tab"
      class:active={active === tab.value}
      onclick={() => onchange(tab.value)}
      role="tab"
      aria-selected={active === tab.value}
    >
      {tab.label}
    </button>
  {/each}
</div>
