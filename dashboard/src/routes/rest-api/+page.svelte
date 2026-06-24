<script lang="ts">
	import Breadcrumbs from '$lib/components/Breadcrumbs.svelte';
	import Card from '$lib/components/Card.svelte';
	import CopyButton from '$lib/components/CopyButton.svelte';

	let baseUrl = 'http://localhost:8080';
	let activeLang = $state('curl');

	let langs = ['curl', 'js', 'go', 'python'];
	let langLabels: Record<string, string> = { curl: 'cURL', js: 'JavaScript', go: 'Go', python: 'Python' };

	let examples: Record<string, string> = {
		curl: `curl -H "apikey: YOUR_API_KEY" ${baseUrl}/api/v1/{app_slug}/users`,
		js: `fetch('${baseUrl}/api/v1/{app_slug}/users', {\n  headers: { apikey: 'YOUR_API_KEY' }\n}).then(r => r.json())`,
		go: `package main\n\nimport (\n\t"fmt"\n\t"io"\n\t"net/http"\n)\n\nfunc main() {\n\treq, _ := http.NewRequest("GET", "${baseUrl}/api/v1/{app_slug}/users", nil)\n\treq.Header.Set("apikey", "YOUR_API_KEY")\n\tresp, _ := http.DefaultClient.Do(req)\n\tbody, _ := io.ReadAll(resp.Body)\n\tfmt.Println(string(body))\n}`,
		python: `import requests\n\nresponse = requests.get(\n    '${baseUrl}/api/v1/{app_slug}/users',\n    headers={'apikey': 'YOUR_API_KEY'}\n)\nprint(response.json())`,
	};

	let currentExample = $derived(examples[activeLang]);
</script>

<Breadcrumbs items={[{ label: 'REST API' }]} />

<div class="max-w-7xl mx-auto">
	<div class="mb-8">
		<h1 class="text-2xl font-bold">REST API</h1>
		<p class="text-sm mt-1" style="color: var(--text-secondary)">Every application gets auto-generated RESTful endpoints</p>
	</div>

	<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
		<Card title="Endpoint Patterns">
			<div class="space-y-3">
				{#each [
					{ method: 'GET', pattern: '/api/v1/{app_slug}/{table}', desc: 'List all rows' },
					{ method: 'GET', pattern: '/api/v1/{app_slug}/{table}?id=eq.1', desc: 'Filter rows' },
					{ method: 'GET', pattern: '/api/v1/{app_slug}/{table}?order=id.desc', desc: 'Sorted results' },
					{ method: 'GET', pattern: '/api/v1/{app_slug}/{table}?limit=10&offset=0', desc: 'Paginated results' },
					{ method: 'POST', pattern: '/api/v1/{app_slug}/{table}', desc: 'Create a row' },
					{ method: 'PATCH', pattern: '/api/v1/{app_slug}/{table}?id=eq.1', desc: 'Update rows' },
					{ method: 'DELETE', pattern: '/api/v1/{app_slug}/{table}?id=eq.1', desc: 'Delete rows' },
				] as ep}
					<div class="flex items-center gap-3 p-3 rounded-lg" style="background-color: var(--bg-tertiary);">
						<span class="badge text-xs font-mono" style="background-color: {ep.method === 'GET' ? 'rgba(34,197,94,0.1)' : ep.method === 'POST' ? 'rgba(12,142,229,0.1)' : ep.method === 'PATCH' ? 'rgba(245,158,11,0.1)' : 'rgba(239,68,68,0.1)'}; color: {ep.method === 'GET' ? 'var(--success)' : ep.method === 'POST' ? 'var(--accent)' : ep.method === 'PATCH' ? 'var(--warning)' : 'var(--danger)'}">
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
			<p class="text-xs mt-3" style="color: var(--text-tertiary)">Replace <span class="font-mono">{'{app_slug}'}</span> with your application's slug and <span class="font-mono">YOUR_API_KEY</span> with a valid API key.</p>
		</Card>
	</div>
</div>
