<script lang="ts">
	import { popoverStore } from '$lib/stores/popovers.svelte';
</script>

<button
	class="toolbar-button-base"
	aria-label="Configure Wallhaven search settings"
	title="Configure Wallhaven search settings"
	onclick={(e) => popoverStore.toggle('wallhaven', e)}
	type="button"
>
	<svg
		width="18px"
		height="18px"
		viewBox="0 0 24 24"
		fill="none"
		stroke="currentColor"
		stroke-width="2"
		stroke-linecap="round"
		stroke-linejoin="round"
		class="hover:text-brand text-zinc-400 transition-colors duration-300"
	>
		<circle cx="6" cy="12" r="3"></circle>
		<circle cx="18" cy="12" r="3"></circle>
		<path d="M9 12h6"></path>
		<path d="M6 9v6"></path>
		<path d="M18 9v6"></path>
	</svg>
</button>
