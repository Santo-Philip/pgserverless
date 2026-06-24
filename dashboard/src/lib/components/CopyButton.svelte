<script lang="ts">
	let { text }: { text: string } = $props();

	let copied = $state(false);

	async function copy() {
		try {
			await navigator.clipboard.writeText(text);
			copied = true;
			setTimeout(() => copied = false, 1500);
		} catch {}
	}
</script>

<button onclick={copy} class="btn btn-ghost btn-sm" title="Copy">
	{copied ? 'Copied' : 'Copy'}
</button>
