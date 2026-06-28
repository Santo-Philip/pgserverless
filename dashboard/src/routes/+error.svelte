<script lang="ts">
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';

  let { status, error } = $props();

  let message = $derived(
    status === 404 ? 'Page not found' :
    status === 403 ? 'Access denied' :
    status === 500 ? 'Server error' :
    'An unexpected error occurred'
  );

  let description = $derived(
    status === 404 ? 'The page you are looking for does not exist.' :
    status === 403 ? 'You do not have permission to access this page.' :
    status === 500 ? 'Something went wrong on our end. Please try again.' :
    'Please try again or contact support.'
  );
</script>

<div class="min-h-screen flex items-center justify-center p-4" style="background-color: var(--bg-primary);">
  <div class="text-center max-w-md">
    <div class="w-16 h-16 rounded-2xl flex items-center justify-center text-2xl font-bold mx-auto mb-6" style="background-color: rgba(239,68,68,0.1); color: var(--danger);">
      {status}
    </div>
    <h1 class="text-2xl font-bold mb-2">{message}</h1>
    <p class="text-sm mb-8" style="color: var(--text-secondary)">{description}</p>
    {#if error}
      <p class="text-xs mb-6 font-mono p-3 rounded-lg" style="background-color: var(--bg-tertiary); color: var(--text-tertiary);">{String(error)}</p>
    {/if}
    <div class="flex items-center justify-center gap-3">
      <button onclick={() => goto('/')} class="btn btn-primary">Go Home</button>
      <button onclick={() => history.back()} class="btn btn-secondary">Go Back</button>
    </div>
  </div>
</div>
