<script lang="ts">
	import { popoverStore } from '$lib/stores/popovers.svelte';
</script>

<button
	class="toolbar-button-base"
	onclick={(e) => popoverStore.toggle('saved', e)}
	aria-label="Saved palettes"
	title="Open saved palettes"
	type="button"
>
	<svg
		class="hover:text-brand h-5 w-5 text-zinc-400 transition-colors duration-300"
		viewBox="0 0 24 24"
		fill="none"
		stroke="currentColor"
		stroke-width="2"
		stroke-linecap="round"
		stroke-linejoin="round"
		aria-hidden="true"
	>
		<circle cx="7" cy="8" r="3"></circle>
		<circle cx="16.5" cy="8" r="3"></circle>
		<circle cx="12" cy="16" r="3"></circle>
	</svg>
</button>
