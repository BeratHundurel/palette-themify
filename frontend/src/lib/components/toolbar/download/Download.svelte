<script lang="ts">
	import { appStore } from '$lib/stores/app.svelte';
</script>

<button
	id="download-button"
	onclick={() => appStore.downloadImage()}
	class="toolbar-button-base"
	disabled={appStore.state.isExtracting}
	aria-label="Download image"
	title="Download the current image"
>
	<svg
		class="hover:text-brand h-5 w-5 text-zinc-400 transition-colors duration-300"
		fill="none"
		stroke="currentColor"
		viewBox="0 0 24 24"
	>
		<path
			stroke-linecap="round"
			stroke-linejoin="round"
			stroke-width="2"
			d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
		/>
	</svg>
</button>
