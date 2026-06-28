<script lang="ts">
  interface SpeedDialAction {
    label: string;
    icon: string;
    action: () => void;
  }

  let {
    icon = '+',
    speedDial = [] as SpeedDialAction[],
    bottom = 80,
    right = 20,
  }: {
    icon?: string;
    speedDial?: SpeedDialAction[];
    bottom?: number;
    right?: number;
  } = $props();

  let open = $state(false);
</script>

<div class="fixed z-40" style="bottom: {bottom}px; right: {right}px;" role="group" aria-label="Floating actions">
  {#if speedDial.length > 0}
    <div class="absolute bottom-full right-0 mb-3 flex flex-col items-end gap-2">
      {#each speedDial as action, i}
        <button
          class="flex items-center gap-2 px-4 py-2.5 rounded-xl shadow-lg transition-all duration-200 whitespace-nowrap"
          style="background-color: var(--bg-card); color: var(--text-primary); border: 1px solid var(--border); transform: {open ? 'translateY(0) scale(1)' : 'translateY(10px) scale(0.8)'}; opacity: {open ? 1 : 0}; pointer-events: {open ? 'auto' : 'none'}; transition-delay: {i * 30}ms;"
          onclick={() => { action.action(); open = false; }}
        >
          <span class="text-lg">{action.icon}</span>
          <span class="text-sm font-medium">{action.label}</span>
        </button>
      {/each}
    </div>
  {/if}

  <button
    class="fab"
    onclick={() => { if (speedDial.length > 0) { open = !open; } }}
    aria-label={open ? 'Close actions' : 'Open actions'}
    style="transform: {open ? 'rotate(45deg)' : 'rotate(0deg)'};"
  >
    {icon}
  </button>
</div>
