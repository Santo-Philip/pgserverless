<script lang="ts">
  let {
    value = $bindable(''),
    placeholder = 'Search...',
    onsearch,
    debounceMs = 300,
  }: {
    value?: string;
    placeholder?: string;
    onsearch?: (q: string) => void;
    debounceMs?: number;
  } = $props();

  let timer: ReturnType<typeof setTimeout> | undefined;

  function handleInput(e: Event) {
    const target = e.target as HTMLInputElement;
    value = target.value;
    clearTimeout(timer);
    if (onsearch) {
      timer = setTimeout(() => onsearch(value), debounceMs);
    }
  }

  function handleClear() {
    value = '';
    onsearch?.('');
  }
</script>

<div class="relative">
  <span class="absolute left-3.5 top-1/2 -translate-y-1/2 text-base pointer-events-none" style="color: var(--text-secondary);" aria-hidden="true">\u2315</span>
  <input
    type="text"
    {placeholder}
    value={value}
    oninput={handleInput}
    class="input pl-10 pr-10"
    aria-label={placeholder}
  />
  {#if value}
    <button
      onclick={handleClear}
      class="absolute right-3 top-1/2 -translate-y-1/2 flex items-center justify-center w-6 h-6 rounded-lg transition-colors hover:bg-[var(--bg-hover)]"
      style="color: var(--text-secondary);"
      aria-label="Clear search"
    >
      &times;
    </button>
  {/if}
</div>
