<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import Card from '$lib/components/Card.svelte';
	import Breadcrumbs from '$lib/components/Breadcrumbs.svelte';
	import Alert from '$lib/components/Alert.svelte';

	let settings = $state<PlatformSettings | null>(null);
	let loading = $state(true);
	let saving = $state(false);
	let message = $state('');

	onMount(async () => {
		try {
			settings = await api.getSettings();
		} catch {}
		loading = false;
	});

	async function handleSave() {
		if (!settings) return;
		saving = true;
		message = '';
		try {
			await api.updateSettings(settings);
			message = 'Settings saved successfully.';
		} catch { message = 'Failed to save settings.'; }
		saving = false;
	}
</script>

<Breadcrumbs items={[{ label: 'Settings' }]} />

<div class="max-w-4xl mx-auto">
	<div class="mb-8">
		<h1 class="text-2xl font-bold">Platform Settings</h1>
		<p class="text-sm mt-1" style="color: var(--text-secondary)">Configure platform-wide settings</p>
	</div>

	<Alert message={message} type={message.includes('successfully') ? 'success' : message ? 'error' : undefined} />

	{#if loading}
		<div class="card p-12"><div class="skeleton h-8 w-full mb-4"></div><div class="skeleton h-8 w-full"></div></div>
	{:else if settings}
		<Card title="Platform Configuration">
			<form onsubmit={handleSave} class="space-y-5">
				<div>
					<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Default Region</label>
					<select bind:value={settings.region} class="input">
						<option value="us-east">US East (us-east)</option>
						<option value="us-west">US West (us-west)</option>
						<option value="eu-west">EU West (eu-west)</option>
					</select>
				</div>

				<div>
					<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Default Visibility</label>
					<select bind:value={settings.default_visibility} class="input">
						<option value="public">Public</option>
						<option value="private">Private</option>
					</select>
				</div>

				<button type="submit" disabled={saving} class="btn btn-primary">
					{saving ? 'Saving...' : 'Save Settings'}
				</button>
			</form>
		</Card>
	{/if}
</div>
