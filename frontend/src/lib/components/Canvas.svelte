<script lang="ts">
	import { appStore } from '$lib/stores/app.svelte';
	import { cn } from '$lib/utils';
</script>

<section class="mb-6 flex w-full justify-center px-4">
	<div class="flex w-full max-w-[90vw] justify-center lg:max-w-200">
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
