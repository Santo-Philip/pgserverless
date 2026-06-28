<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	let { open = false, onclose }: { open?: boolean; onclose?: () => void } = $props();

	let query = $state('');
	let results = $state<{ label: string; href: string; category: string }[]>([]);
	let selectedIndex = $state(0);

	const pages = [
		{ label: 'Dashboard', href: '/dashboard', category: 'Pages' },
		{ label: 'Applications', href: '/apps', category: 'Pages' },
		{ label: 'Database', href: '/database', category: 'Pages' },
		{ label: 'API Keys', href: '/api-keys', category: 'Pages' },
		{ label: 'REST API', href: '/rest-api', category: 'Pages' },
		{ label: 'Extensions', href: '/extensions', category: 'Pages' },
		{ label: 'Logs', href: '/logs', category: 'Pages' },
		{ label: 'Settings', href: '/settings', category: 'Pages' },
		{ label: 'Users', href: '/users', category: 'Admin' },
	];

	$effect(() => {
		if (!query.trim()) {
			results = pages;
		} else {
			const q = query.toLowerCase();
			results = pages.filter(p => p.label.toLowerCase().includes(q));
		}
		selectedIndex = 0;
	});

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'ArrowDown') { e.preventDefault(); selectedIndex = Math.min(selectedIndex + 1, results.length - 1); }
		if (e.key === 'ArrowUp') { e.preventDefault(); selectedIndex = Math.max(selectedIndex - 1, 0); }
		if (e.key === 'Enter' && results[selectedIndex]) {
			goto(results[selectedIndex].href);
			onclose?.();
		}
		if (e.key === 'Escape') onclose?.();
	}
</script>

{#if open}
	<div class="fixed inset-0 z-50" style="background-color: rgba(0,0,0,0.5)" onclick={() => onclose?.()} onkeydown={handleKeydown} role="dialog" tabindex="-1">
		<div class="mx-auto mt-[15vh] max-w-xl" onclick={(e) => e.stopPropagation()}>
			<div class="card p-0" style="box-shadow: 0 25px 50px rgba(0,0,0,0.4);">
				<div class="flex items-center gap-3 px-4 border-b" style="border-color: var(--border-primary);">
					<span class="text-sm" style="color: var(--text-tertiary)">⌘</span>
					<input
						type="text"
						bind:value={query}
						placeholder="Search pages..."
						class="input !border-0 !bg-transparent !shadow-none !px-0"
						autofocus
					/>
				</div>
				<div class="max-h-80 overflow-y-auto p-2">
					{#each results as item, i}
						<button
							class="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition-colors"
							class:active={i === selectedIndex}
							style={i === selectedIndex ? 'background-color: var(--bg-hover); color: var(--text-primary)' : 'color: var(--text-secondary)'}
							onclick={() => { goto(item.href); onclose?.(); }}
						>
							<span>{item.label}</span>
							<span class="ml-auto text-xs" style="color: var(--text-tertiary)">{item.category}</span>
						</button>
					{/each}
				</div>
			</div>
		</div>
	</div>
{/if}
