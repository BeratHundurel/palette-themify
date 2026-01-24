<script lang="ts">
	import { appStore } from '$lib/stores/app.svelte';
	import { preventDefault, cn } from '$lib/utils';
</script>

<section
	class={cn(
		'absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2',
		appStore.state.imageLoaded ? 'pointer-events-none hidden' : 'pointer-events-auto'
	)}
	aria-hidden={appStore.state.imageLoaded ? 'true' : 'false'}
>
	<button
		type="button"
		ondrop={appStore.handleDrop}
		ondragover={preventDefault}
		ondragenter={preventDefault}
		ondragleave={preventDefault}
		onclick={appStore.triggerFileSelect}
		class={cn(
			'group flex cursor-pointer flex-col items-center justify-center rounded-xl border-2 border-dashed border-white/50 bg-white/10 p-12 transition-[background-color,border-color] duration-300',
			'hover:border-white hover:bg-white/20'
		)}
		aria-label="Upload an image or drag and drop it here"
	>
		<svg
			viewBox="0 0 24 24"
			width="4rem"
			height="4rem"
			fill="#fff"
			role="img"
			aria-label="Upload icon"
			class="mb-3 opacity-80 transition-opacity group-hover:opacity-100"
		>
			<path
				d="M19 7v3h-2V7h-3V5h3V2h2v3h3v2h-3zm-3 4V8h-3V5H5a2 2 0 00-2 2v12c0 1.1.9 2 2 2h12a2 2 0 002-2v-8h-3zM5 19l3-4 2 3 3-4 4 5H5z"
			/>
		</svg>
		<span class="text-sm text-zinc-300 group-hover:text-white"> Upload an image or drag and drop it here </span>
		<input
			type="file"
			name="file"
			bind:this={appStore.state.fileInput}
			class="hidden"
			accept="image/*"
			oninput={appStore.onFileChange}
			aria-label="Choose image to extract a palette from"
		/>
	</button>
</section>
