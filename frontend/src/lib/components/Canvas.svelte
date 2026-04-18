<script lang="ts">
	import { appStore } from '$lib/stores/app/store.svelte';
	import { cn } from '$lib/utils';

	$effect(() => {
		if (!appStore.state.canvas || !appStore.state.imageLoaded || !appStore.state.image) return;
		appStore.restoreCanvasImage();
	});
</script>

<section class="flex w-full justify-center px-4">
	<div class="flex w-full max-w-200 justify-center">
		<canvas
			bind:this={appStore.state.canvas}
			onmousedown={appStore.handleMouseDown}
			onmousemove={appStore.handleMouseMove}
			onmouseup={appStore.handleMouseUp}
			class={cn(
				'mx-auto rounded-xl border border-zinc-600 shadow-lg',
				appStore.state.imageLoaded ? 'pointer-events-auto block' : 'pointer-events-none hidden'
			)}
		></canvas>
	</div>
</section>
