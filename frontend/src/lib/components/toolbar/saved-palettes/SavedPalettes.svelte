<script lang="ts">
	import { popoverStore } from '$lib/stores/popovers.svelte';
</script>

<button
	class="toolbar-button-base"
	onclick={(e) => popoverStore.toggle('saved', e)}
	aria-label="Show saved palettes"
	title="View and load saved palettes"
	type="button"
>
	<span>🎨</span>
</button>
