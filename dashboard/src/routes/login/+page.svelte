<script lang="ts">
	import { login } from '$lib/stores/auth';
	import { goto } from '$app/navigation';

	let email = $state('');
	let password = $state('');
	let error = $state('');
	let loading = $state(false);

	async function handleSubmit() {
		error = '';
		loading = true;
		try {
			await login(email, password);
			goto('/');
		} catch (e) {
			error = e instanceof Error ? e.message : 'Login failed';
		} finally {
			loading = false;
		}
	}
</script>

<div class="min-h-screen flex items-center justify-center p-4" style="background-color: var(--bg-primary);">
	<div class="w-full max-w-sm">
		<div class="text-center mb-8">
			<div class="w-10 h-10 rounded-xl flex items-center justify-center text-lg font-bold mx-auto mb-4" style="background-color: var(--accent); color: #fff;">N</div>
			<h1 class="text-xl font-bold">Welcome back</h1>
			<p class="text-sm mt-1" style="color: var(--text-secondary)">Sign in to your account</p>
		</div>

		<div class="card p-6">
			<form onsubmit={handleSubmit} class="space-y-4">
				{#if error}
					<div class="px-4 py-3 rounded-lg text-sm" style="background-color: rgba(239,68,68,0.1); color: var(--danger)">{error}</div>
				{/if}

				<div>
					<label for="email" class="block text-sm font-medium mb-1.5" style="color: var(--text-secondary)">Email</label>
					<input id="email" type="email" bind:value={email} required class="input" placeholder="you@example.com" />
				</div>

				<div>
					<label for="password" class="block text-sm font-medium mb-1.5" style="color: var(--text-secondary)">Password</label>
					<input id="password" type="password" bind:value={password} required class="input" placeholder="Enter your password" />
				</div>

				<button type="submit" disabled={loading} class="btn btn-primary w-full">
					{loading ? 'Signing in...' : 'Sign In'}
				</button>
			</form>

			<p class="mt-6 text-center text-sm" style="color: var(--text-tertiary)">
				Don't have an account? <a href="/register" class="link">Register</a>
			</p>
		</div>
	</div>
</div>
