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
		<circle cx="12" cy="12" r="3"></circle>
		<path d="M12 1v6m0 6v6m4.22-13.22l4.24 4.24M1.54 1.54l4.24 4.24M20.46 20.46l-4.24-4.24M1.54 20.46l4.24-4.24"></path>
	</svg>
</button>
