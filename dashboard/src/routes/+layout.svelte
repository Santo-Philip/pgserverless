<script lang="ts">
	import '../app.css';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import { isAuthenticated } from '$lib/stores/auth';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	let { children } = $props();
	let authChecked = $state(false);

	onMount(() => {
		if (!$isAuthenticated && !$page.url.pathname.startsWith('/login') && !$page.url.pathname.startsWith('/register')) {
			goto('/login');
		}
		authChecked = true;
	});
</script>

{#if $isAuthenticated}
	<Sidebar>
		{@render children()}
	</Sidebar>
{:else if $page.url.pathname.startsWith('/login') || $page.url.pathname.startsWith('/register')}
	{@render children()}
{/if}
