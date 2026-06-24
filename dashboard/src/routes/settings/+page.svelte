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
		<form onsubmit={handleSave} class="space-y-6">
			<Card title="General">
				<div class="space-y-4">
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">App Name</label>
						<input type="text" bind:value={settings.app_name} class="input" placeholder="Nexbic Platform" />
					</div>
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Log Level</label>
						<select bind:value={settings.log_level} class="input">
							<option value="debug">Debug</option>
							<option value="info">Info</option>
							<option value="warn">Warn</option>
							<option value="error">Error</option>
						</select>
					</div>
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
				</div>
			</Card>

			<Card title="Authentication">
				<div class="space-y-4">
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">JWT Access TTL</label>
						<input type="text" bind:value={settings.jwt_access_ttl} class="input font-mono" placeholder="15m" />
					</div>
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">JWT Refresh TTL</label>
						<input type="text" bind:value={settings.jwt_refresh_ttl} class="input font-mono" placeholder="168h" />
					</div>
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">OTP Expiry</label>
						<input type="text" bind:value={settings.otp_expiry} class="input font-mono" placeholder="5m" />
					</div>
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Default User Role</label>
						<select bind:value={settings.default_user_role} class="input">
							<option value="authenticated">Authenticated</option>
							<option value="admin">Admin</option>
						</select>
					</div>
					<div class="flex items-center gap-3">
						<input type="checkbox" bind:checked={settings.registration_enabled} id="reg_enabled" class="checkbox" />
						<label for="reg_enabled" class="text-sm font-medium">Registration Enabled</label>
					</div>
					<div class="flex items-center gap-3">
						<input type="checkbox" bind:checked={settings.email_verification_required} id="email_verify" class="checkbox" />
						<label for="email_verify" class="text-sm font-medium">Email Verification Required</label>
					</div>
				</div>
			</Card>

			<Card title="Security">
				<div class="space-y-4">
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">CORS Origins</label>
						<input type="text" bind:value={settings.cors_origins} class="input" placeholder="*" />
					</div>
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Password Min Length</label>
						<input type="number" bind:value={settings.password_min_length} class="input" min="6" max="128" />
					</div>
					<div class="flex items-center gap-3">
						<input type="checkbox" bind:checked={settings.password_require_special} id="pwd_special" class="checkbox" />
						<label for="pwd_special" class="text-sm font-medium">Password Requires Special Characters</label>
					</div>
					<div class="flex items-center gap-3">
						<input type="checkbox" bind:checked={settings.password_require_numbers} id="pwd_numbers" class="checkbox" />
						<label for="pwd_numbers" class="text-sm font-medium">Password Requires Numbers</label>
					</div>
				</div>
			</Card>

			<Card title="Database">
				<div class="space-y-4">
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Max DB Connections</label>
						<input type="number" bind:value={settings.max_db_connections} class="input" min="1" max="500" />
					</div>
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Min DB Connections</label>
						<input type="number" bind:value={settings.min_db_connections} class="input" min="1" max="100" />
					</div>
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Health Check Period</label>
						<input type="text" bind:value={settings.health_check_period} class="input font-mono" placeholder="30s" />
					</div>
				</div>
			</Card>

			<Card title="Limits">
				<div class="space-y-4">
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">API Rate Limits</label>
						<input type="text" bind:value={settings.api_rate_limits} class="input" placeholder="1000/h" />
					</div>
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Storage Limit (MB)</label>
						<input type="number" bind:value={settings.storage_limit_mb} class="input" min="0" max="100000" />
					</div>
				</div>
			</Card>

			<Card title="Feature Flags">
				<div class="space-y-4">
					<div class="flex items-center gap-3">
						<input type="checkbox" bind:checked={settings.monitoring_enabled} id="monitoring" class="checkbox" />
						<label for="monitoring" class="text-sm font-medium">Monitoring Enabled</label>
					</div>
					<div class="flex items-center gap-3">
						<input type="checkbox" bind:checked={settings.maintenance_mode} id="maintenance" class="checkbox" />
						<label for="maintenance" class="text-sm font-medium">Maintenance Mode</label>
					</div>
					<div>
						<label class="block text-sm font-medium mb-1" style="color: var(--text-secondary)">Feature Flags</label>
						<textarea bind:value={settings.feature_flags} class="input" rows="3" placeholder="JSON feature flags"></textarea>
					</div>
				</div>
			</Card>

			<div class="flex justify-end">
				<button type="submit" disabled={saving} class="btn btn-primary">
					{saving ? 'Saving...' : 'Save Settings'}
				</button>
			</div>
		</form>
	{/if}
</div>
