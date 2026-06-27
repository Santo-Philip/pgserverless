	<script lang="ts">
	import '../app.css';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import { isAuthenticated } from '$lib/stores/auth';
	import { page } from '$app/stores';
	import { navigating } from '$app/stores';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	let { children } = $props();
	let authChecked = $state(false);

	onMount(() => {
		if (!$isAuthenticated && $page.url.pathname !== '/login') {
			goto('/login');
		}
		authChecked = true;
	});
</script>

{#if $navigating}
	<div class="fixed top-0 left-0 right-0 z-[100] h-0.5" style="background-color: var(--accent);">
		<div class="h-full w-full origin-left-left" style="background-color: var(--accent); animation: progress 30s ease-in-out infinite;"></div>
	</div>
{/if}

{#if $isAuthenticated}
	<Sidebar>
		{@render children()}
	</Sidebar>
{:else if $page.url.pathname.startsWith('/login')}
	{@render children()}
{/if}

<style>
	@keyframes progress {
		0% { transform: scaleX(0); }
		50% { transform: scaleX(0.5); }
		100% { transform: scaleX(0.8); }
	}
</style>
