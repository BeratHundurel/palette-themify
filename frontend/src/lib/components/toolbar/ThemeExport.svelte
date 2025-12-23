<script lang="ts">
	import { appStore } from '$lib/stores/app.svelte';
	import { popoverStore } from '$lib/stores/popovers.svelte';

	function handleClick(e: MouseEvent) {
		if (!appStore.state.colors || appStore.state.colors.length === 0) {
			return;
		}
		popoverStore.toggle('themeExport', e);
	}

	const hasColors = $derived(appStore.state.colors.length > 0);
</script>

<button onclick={handleClick} disabled={!hasColors} class="toolbar-button-base" title="Export Theme">
	<svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
		<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17h.01" />
	</svg>
</button>
