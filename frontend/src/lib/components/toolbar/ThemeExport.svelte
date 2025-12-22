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

<button onclick={handleClick} disabled={!hasColors} class="toolbar-button-base" title="Export VS Code Theme">
	<span class="text-xs">VSC</span>
</button>
