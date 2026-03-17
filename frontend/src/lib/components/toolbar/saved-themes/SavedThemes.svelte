<script lang="ts">
	import { popoverStore } from '$lib/stores/popovers.svelte';
</script>

<button
	class="toolbar-button-base"
	onclick={(e) => popoverStore.toggle('themes', e)}
	aria-label="Show saved themes"
	title="View and load saved themes"
	type="button"
>
	<span>🧩</span>
</button>
