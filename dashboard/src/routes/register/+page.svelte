<script lang="ts">
	import { register } from '$lib/stores/auth';
	import { goto } from '$app/navigation';

	let name = $state('');
	let email = $state('');
	let password = $state('');
	let error = $state('');
	let loading = $state(false);

	async function handleSubmit() {
		error = '';
		loading = true;

		try {
			await register(email, password, name || undefined);
			goto('/');
		} catch (e) {
			error = e instanceof Error ? e.message : 'Registration failed';
		} finally {
			loading = false;
		}
	}
</script>

<div class="max-w-md mx-auto mt-20">
	<div class="bg-white rounded-lg shadow p-8">
		<h1 class="text-2xl font-bold mb-6 text-center">Create Account</h1>

		<form onsubmit={handleSubmit} class="space-y-4">
			{#if error}
				<div class="bg-red-50 text-red-700 p-3 rounded text-sm">{error}</div>
			{/if}

			<div>
				<label for="name" class="block text-sm font-medium text-gray-700 mb-1">Name</label>
				<input
					id="name"
					type="text"
					bind:value={name}
					class="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-nexbic-500"
				/>
			</div>

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
					minlength="8"
					class="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-nexbic-500"
				/>
			</div>

			<button
				type="submit"
				disabled={loading}
				class="w-full bg-nexbic-600 text-white py-2 rounded hover:bg-nexbic-700 disabled:opacity-50 transition-colors"
			>
				{loading ? 'Creating account...' : 'Create Account'}
			</button>
		</form>

		<p class="mt-4 text-center text-sm text-gray-500">
			Already have an account?
			<a href="/login" class="text-nexbic-600 hover:underline">Sign In</a>
		</p>
	</div>
</div>
