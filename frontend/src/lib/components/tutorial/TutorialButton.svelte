<script lang="ts">
	import { tutorialStore } from '$lib/stores/tutorial.svelte';
	import { popoverStore } from '$lib/stores/popovers.svelte';

	function restartTutorial() {
		popoverStore.close();
		tutorialStore.reset();
		tutorialStore.start();
	}
</script>

<button
	class="border-brand/50 hover:shadow-brand-lg flex w-32 cursor-pointer items-center justify-center gap-2 rounded-md border bg-zinc-900 py-2 text-sm font-medium transition-[background-color,border-color,box-shadow,color] duration-300"
	onclick={restartTutorial}
	title="Restart the interactive tutorial"
	type="button"
>
	<span>🎓</span>
	<span>Tutorial</span>
</button>
