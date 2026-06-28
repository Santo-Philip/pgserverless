<script lang="ts">
  import { login } from '$lib/stores/auth';
  import { goto } from '$app/navigation';
  import { APP_NAME, APP_LOGO_LETTER } from '$lib/config/brand';

  let email = $state('');
  let password = $state('');
  let error = $state('');
  let loading = $state(false);

  async function handleSubmit(e: Event) {
    e.preventDefault();
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
      <div class="w-12 h-12 rounded-2xl flex items-center justify-center text-xl font-bold mx-auto mb-4" style="background-color: var(--accent); color: #fff;">{APP_LOGO_LETTER}</div>
      <h1 class="text-xl font-bold">{APP_NAME}</h1>
      <p class="text-sm mt-1" style="color: var(--text-secondary)">PostgreSQL Admin Dashboard</p>
    </div>

    <div class="card p-6">
      <form onsubmit={handleSubmit} class="space-y-4">
        {#if error}
          <div class="px-4 py-3 rounded-lg text-sm" style="background-color: rgba(239,68,68,0.1); color: var(--danger)">{error}</div>
        {/if}

        <div>
          <label for="email" class="block text-sm font-medium mb-1.5" style="color: var(--text-secondary)">Email</label>
          <input id="email" type="email" bind:value={email} required class="input text-base py-3" placeholder="you@example.com" />
        </div>

        <div>
          <label for="password" class="block text-sm font-medium mb-1.5" style="color: var(--text-secondary)">Password</label>
          <input id="password" type="password" bind:value={password} required class="input text-base py-3" placeholder="Enter your password" />
        </div>

        <button type="submit" disabled={loading} class="btn btn-primary w-full py-3 text-base rounded-xl">
          {loading ? 'Signing in...' : 'Sign In'}
        </button>
      </form>
    </div>
  </div>
</div>
