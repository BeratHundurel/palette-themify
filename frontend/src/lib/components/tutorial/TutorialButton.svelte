<script lang="ts">
	import BrandButton from '$lib/components/ui/BrandButton.svelte';
	import { tutorialStore } from '$lib/stores/tutorial.svelte';
	import { popoverStore } from '$lib/stores/popovers.svelte';

	function restartTutorial() {
		popoverStore.close();
		tutorialStore.reset();
		tutorialStore.start();
	}
</script>

<BrandButton
	class="flex w-32 items-center justify-center gap-2"
	onclick={restartTutorial}
	title="Restart the interactive tutorial"
>
	<span>Tutorial</span>
</BrandButton>
