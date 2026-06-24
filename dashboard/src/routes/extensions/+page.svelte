<script lang="ts">
	import Breadcrumbs from '$lib/components/Breadcrumbs.svelte';
	import Card from '$lib/components/Card.svelte';
	import Badge from '$lib/components/Badge.svelte';

	let extensions = [
		{ name: 'pgcrypto', description: 'Cryptographic functions — hashing, encryption, random data', category: 'Security' },
		{ name: 'uuid-ossp', description: 'Universally Unique Identifier (UUID) generation', category: 'Core' },
		{ name: 'pgvector', description: 'Vector similarity search for AI/embeddings workloads', category: 'AI' },
		{ name: 'citext', description: 'Case-insensitive character string type', category: 'Text' },
		{ name: 'pg_trgm', description: 'Trigram-based text search and similarity', category: 'Search' },
		{ name: 'hstore', description: 'Key-value store within a single PostgreSQL value', category: 'Data' },
		{ name: 'ltree', description: 'Hierarchical tree structure data type', category: 'Data' },
		{ name: 'pg_stat_statements', description: 'Track execution statistics of SQL statements', category: 'Monitoring' },
		{ name: 'postgis', description: 'Geospatial data storage and queries', category: 'Geo' },
		{ name: 'pg_partman', description: 'Automated table partitioning management', category: 'Performance' },
		{ name: 'pg_bigm', description: 'Full-text search engine for Japanese and other languages', category: 'Search' },
		{ name: 'pg_repack', description: 'Rebuild tables online to remove bloat', category: 'Maintenance' },
	];
</script>

<Breadcrumbs items={[{ label: 'Extensions' }]} />

<div class="max-w-7xl mx-auto">
	<div class="mb-8">
		<h1 class="text-2xl font-bold">Extensions</h1>
		<p class="text-sm mt-1" style="color: var(--text-secondary)">Available PostgreSQL extensions for your applications</p>
	</div>

	<Card>
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
			{#each extensions as ext}
				<div class="p-4 rounded-xl border" style="border-color: var(--border-primary); background-color: var(--bg-tertiary);">
					<div class="flex items-center justify-between mb-2">
						<span class="text-sm font-semibold font-mono">{ext.name}</span>
						<Badge status={ext.category.toLowerCase()} />
					</div>
					<p class="text-xs" style="color: var(--text-secondary)">{ext.description}</p>
				</div>
			{/each}
		</div>
	</Card>
</div>
