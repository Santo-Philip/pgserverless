<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { api } from '$lib/api/client';
	import Breadcrumbs from '$lib/components/Breadcrumbs.svelte';
	import Card from '$lib/components/Card.svelte';
	import CopyButton from '$lib/components/CopyButton.svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import type { App } from '$lib/types';

	let apps = $state<App[]>([]);
	let loading = $state(true);
	let selectedSlug = $state('');
	let activeLang = $state('curl');

	let origin = $derived(browser ? window.location.origin : '');
	let slug = $derived(selectedSlug || apps[0]?.slug || '{app_slug}');

	let langs = ['curl', 'js', 'go', 'python'];
	let langLabels: Record<string, string> = { curl: 'cURL', js: 'JavaScript', go: 'Go', python: 'Python' };

	let examples = $derived({
		curl: `curl -H "apikey: YOUR_API_KEY" ${origin}/api/v1/${slug}/users`,
		js: `const response = await fetch('${origin}/api/v1/${slug}/users', {\n  headers: { apikey: 'YOUR_API_KEY' }\n});\nconst data = await response.json();`,
		go: `req, _ := http.NewRequest("GET", "${origin}/api/v1/${slug}/users", nil)\nreq.Header.Set("apikey", "YOUR_API_KEY")\nresp, _ := http.DefaultClient.Do(req)\nbody, _ := io.ReadAll(resp.Body)\nfmt.Println(string(body))`,
		python: `import requests\n\nresponse = requests.get(\n    '${origin}/api/v1/${slug}/users',\n    headers={'apikey': 'YOUR_API_KEY'}\n)\nprint(response.json())`,
	});

	let currentExample = $derived(examples[activeLang as keyof typeof examples]);

	onMount(async () => {
		try {
			const result = await api.listApps();
			apps = result.data;
			if (apps.length > 0) selectedSlug = apps[0].slug;
		} catch {}
		loading = false;
	});
</script>

<Breadcrumbs items={[{ label: 'REST API' }]} />

<div class="max-w-7xl mx-auto">
	<div class="mb-8">
		<h1 class="text-2xl font-bold">REST API</h1>
		<p class="text-sm mt-1" style="color: var(--text-secondary)">Every application gets auto-generated RESTful endpoints</p>
	</div>

	{#if loading}
		<Skeleton rows={4} />
	{:else if apps.length === 0}
		<EmptyState icon="↗" title="No applications" description="Create an application to get REST API endpoints.">
			<a href="/apps" class="btn btn-primary">Create Application</a>
		</EmptyState>
	{:else}
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
			<Card title="Endpoint Patterns">
				<div class="space-y-3">
					{#each [
						{ method: 'GET', pattern: `/api/v1/${slug}/{table}`, desc: 'List all rows' },
						{ method: 'GET', pattern: `/api/v1/${slug}/{table}?id=eq.1`, desc: 'Filter rows' },
						{ method: 'GET', pattern: `/api/v1/${slug}/{table}?order=id.desc`, desc: 'Sorted results' },
						{ method: 'GET', pattern: `/api/v1/${slug}/{table}?limit=10&offset=0`, desc: 'Paginated results' },
						{ method: 'POST', pattern: `/api/v1/${slug}/{table}`, desc: 'Create a row' },
						{ method: 'PATCH', pattern: `/api/v1/${slug}/{table}?id=eq.1`, desc: 'Update rows' },
						{ method: 'DELETE', pattern: `/api/v1/${slug}/{table}?id=eq.1`, desc: 'Delete rows' },
					] as ep}
						<div class="flex items-center gap-3 p-3 rounded-lg" style="background-color: var(--bg-tertiary);">
							<span class="badge text-xs font-mono" style="background-color: {ep.method === 'GET' ? 'rgba(34,197,94,0.1)' : ep.method === 'POST' ? 'rgba(59,130,246,0.1)' : ep.method === 'PATCH' ? 'rgba(245,158,11,0.1)' : 'rgba(239,68,68,0.1)'}; color: {ep.method === 'GET' ? 'var(--success)' : ep.method === 'POST' ? 'var(--accent)' : ep.method === 'PATCH' ? 'var(--warning)' : 'var(--danger)'}">
								{ep.method}
							</span>
							<div class="flex-1 min-w-0">
								<p class="text-xs font-mono truncate">{ep.pattern}</p>
								<p class="text-xs mt-0.5" style="color: var(--text-tertiary)">{ep.desc}</p>
							</div>
						</div>
					{/each}
				</div>
			</Card>

			<Card title="Code Examples">
				{#if apps.length > 1}
					<div class="mb-4">
						<label class="block text-xs font-medium mb-1.5" style="color: var(--text-secondary)">Application</label>
						<select bind:value={selectedSlug} class="input">
							{#each apps as app}
								<option value={app.slug}>{app.name} ({app.slug})</option>
							{/each}
						</select>
					</div>
				{/if}

				<div class="flex gap-2 mb-4 flex-wrap">
					{#each langs as lang}
						<button onclick={() => activeLang = lang} class="btn btn-sm {activeLang === lang ? 'btn-primary' : 'btn-ghost'}">
							{langLabels[lang]}
						</button>
					{/each}
				</div>
				<div class="relative">
					<pre class="p-4 rounded-lg text-sm font-mono overflow-x-auto whitespace-pre-wrap" style="background-color: var(--bg-tertiary); color: var(--text-primary);">{currentExample}</pre>
					<div class="absolute top-3 right-3"><CopyButton text={currentExample} /></div>
				</div>
				<p class="text-xs mt-3" style="color: var(--text-tertiary)">Replace <span class="font-mono">YOUR_API_KEY</span> with a valid API key from your application settings.</p>
			</Card>
		</div>
	{/if}
</div>
