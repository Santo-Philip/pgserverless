<script lang="ts">
	import { page } from '$app/stores';
	import { browser } from '$app/environment';
	import Card from '$lib/components/Card.svelte';
	import CopyButton from '$lib/components/CopyButton.svelte';

	let appSlug = $derived($page.params.id);

	let endpoints = $derived([
		{ method: 'GET', path: `/api/v1/${appSlug}/users`, description: 'List all users' },
		{ method: 'GET', path: `/api/v1/${appSlug}/users/:id`, description: 'Get user by ID' },
		{ method: 'POST', path: `/api/v1/${appSlug}/users`, description: 'Create a user' },
		{ method: 'PATCH', path: `/api/v1/${appSlug}/users/:id`, description: 'Update a user' },
		{ method: 'DELETE', path: `/api/v1/${appSlug}/users/:id`, description: 'Delete a user' },
	]);

	let origin = $derived(browser ? window.location.origin : 'http://localhost:8080');

	let snippets = $derived({
		curl: `curl -H "apikey: YOUR_API_KEY" ${origin}/api/v1/${appSlug}/users`,
		js: `fetch('/api/v1/${appSlug}/users', { headers: { apikey: 'YOUR_API_KEY' } })`,
		go: `req, _ := http.NewRequest("GET", "/api/v1/${appSlug}/users", nil)\nreq.Header.Set("apikey", "YOUR_API_KEY")`,
		python: `import requests\nrequests.get('/api/v1/${appSlug}/users', headers={'apikey': 'YOUR_API_KEY'})`,
	});

	let activeLang = $state('curl');
	let langs = ['curl', 'js', 'go', 'python'];
	let langLabels: Record<string, string> = { curl: 'cURL', js: 'JavaScript', go: 'Go', python: 'Python' };
	let currentSnippet = $derived(snippets[activeLang]);
</script>

<div class="space-y-6">
	<Card title="Generated Endpoints">
		<p class="text-sm mb-4" style="color: var(--text-secondary)">All tables in your schema are exposed via RESTful endpoints. Below are examples for the <span class="font-mono text-xs">users</span> table.</p>

		<div class="space-y-2">
			{#each endpoints as ep}
				<div class="flex items-center gap-3 p-3 rounded-lg" style="background-color: var(--bg-tertiary);">
					<span class="badge text-xs font-mono" style="background-color: {ep.method === 'GET' ? 'rgba(34,197,94,0.1)' : ep.method === 'POST' ? 'rgba(12,142,229,0.1)' : ep.method === 'PATCH' ? 'rgba(245,158,11,0.1)' : 'rgba(239,68,68,0.1)'}; color: {ep.method === 'GET' ? 'var(--success)' : ep.method === 'POST' ? 'var(--accent)' : ep.method === 'PATCH' ? 'var(--warning)' : 'var(--danger)'}">
						{ep.method}
					</span>
					<span class="text-sm font-mono flex-1">{ep.path}</span>
					<span class="text-xs" style="color: var(--text-tertiary)">{ep.description}</span>
					<CopyButton text={`${ep.method} ${ep.path}`} />
				</div>
			{/each}
		</div>
	</Card>

	<Card title="Code Examples">
		<div class="flex gap-2 mb-4">
			{#each langs as lang}
				<button
					onclick={() => activeLang = lang}
					class="btn btn-sm {activeLang === lang ? 'btn-primary' : 'btn-ghost'}"
				>
					{langLabels[lang]}
				</button>
			{/each}
		</div>

		<div class="relative">
			<pre class="p-4 rounded-lg text-sm font-mono overflow-x-auto" style="background-color: var(--bg-tertiary); color: var(--text-primary);">{currentSnippet}</pre>
			<div class="absolute top-3 right-3">
				<CopyButton text={currentSnippet} />
			</div>
		</div>
	</Card>
</div>
