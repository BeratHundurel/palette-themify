<script lang="ts">
	import { popoverStore } from '$lib/stores/popovers.svelte';
</script>

<button
	id="tutorial-open-saved-themes"
	class="toolbar-button-base"
	onclick={(e) => popoverStore.toggle('themes', e)}
	aria-label="Saved themes"
	title="Open saved themes"
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
		<path d="M7 4h10a1 1 0 0 1 1 1v15l-6-3-6 3V5a1 1 0 0 1 1-1Z"></path>
		<path d="M9.5 8.5h5"></path>
		<path d="M9.5 11.5h5"></path>
	</svg>
</button>
