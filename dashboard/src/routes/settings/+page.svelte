<script lang="ts">
	import { onMount } from 'svelte';
	import { isAuthenticated } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api/client';

	let settings = $state<PlatformSettings | null>(null);
	let loading = $state(true);
	let saving = $state(false);
	let message = $state('');

	onMount(async () => {
		if (!$isAuthenticated) {
			goto('/login');
			return;
		}

		try {
			settings = await api.getSettings();
		} catch (e) {
			console.error('Failed to load settings', e);
		} finally {
			loading = false;
		}
	});

	async function handleSave() {
		if (!settings) return;
		saving = true;
		message = '';
		try {
			await api.updateSettings(settings);
			message = 'Settings saved successfully.';
		} catch (e) {
			message = 'Failed to save settings.';
		} finally {
			saving = false;
		}
	}
</script>

<div class="max-w-4xl mx-auto">
	<h1 class="text-3xl font-bold mb-8">Settings</h1>

	{#if loading}
		<p class="text-gray-500">Loading...</p>
	{:else if settings}
		<div class="bg-white rounded-lg shadow p-6">
			{#if message}
				<div class="mb-4 p-3 rounded text-sm {message.includes('success') ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'}">
					{message}
				</div>
			{/if}

			<form onsubmit={handleSave} class="space-y-6">
				<div>
					<label for="region" class="block text-sm font-medium text-gray-700 mb-1">Default Region</label>
					<select id="region" bind:value={settings.region} class="w-full px-3 py-2 border rounded">
						<option value="us-east">US East (us-east)</option>
						<option value="us-west">US West (us-west)</option>
						<option value="eu-west">EU West (eu-west)</option>
					</select>
				</div>

				<div>
					<label for="visibility" class="block text-sm font-medium text-gray-700 mb-1">Default Visibility</label>
					<select id="visibility" bind:value={settings.default_visibility} class="w-full px-3 py-2 border rounded">
						<option value="public">Public</option>
						<option value="private">Private</option>
					</select>
				</div>

				<button
					type="submit"
					disabled={saving}
					class="bg-nexbic-600 text-white px-6 py-2 rounded hover:bg-nexbic-700 disabled:opacity-50"
				>
					{saving ? 'Saving...' : 'Save Settings'}
				</button>
			</form>
		</div>
	{/if}
</div>
