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

<div class="max-w-md mx-auto mt-20">
	<div class="bg-white rounded-lg shadow p-8">
		<h1 class="text-2xl font-bold mb-6 text-center">Nexbic Platform</h1>

		<form onsubmit={handleSubmit} class="space-y-4">
			{#if error}
				<div class="bg-red-50 text-red-700 p-3 rounded text-sm">{error}</div>
			{/if}

			<div>
				<label for="email" class="block text-sm font-medium text-gray-700 mb-1">Email</label>
				<input
					id="email"
					type="email"
					bind:value={email}
					required
					class="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-nexbic-500"
				/>
			</div>

			<div>
				<label for="password" class="block text-sm font-medium text-gray-700 mb-1">Password</label>
				<input
					id="password"
					type="password"
					bind:value={password}
					required
					class="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-nexbic-500"
				/>
			</div>

			<button
				type="submit"
				disabled={loading}
				class="w-full bg-nexbic-600 text-white py-2 rounded hover:bg-nexbic-700 disabled:opacity-50 transition-colors"
			>
				{loading ? 'Signing in...' : 'Sign In'}
			</button>
		</form>

		<p class="mt-4 text-center text-sm text-gray-500">
			Don't have an account?
			<a href="/register" class="text-nexbic-600 hover:underline">Register</a>
		</p>
	</div>
</div>
