<script lang="ts">
	import { tutorialStore } from '$lib/stores/tutorial.svelte';
	import { fade, fly } from 'svelte/transition';
	import { onMount } from 'svelte';

	let showPrompt = $state(false);

	onMount(() => {
		if (tutorialStore.shouldShowTutorial() && !tutorialStore.state.isActive) {
			showPrompt = true;
		}
	});

	$effect(() => {
		if (tutorialStore.state.isActive && showPrompt) {
			showPrompt = false;
		}
	});

	function startTutorial() {
		showPrompt = false;
		setTimeout(() => {
			tutorialStore.start();
		}, 300);
	}

	function dismissPrompt() {
		showPrompt = false;
		localStorage.setItem('tutorial-skipped', 'true');
	}
</script>

{#if showPrompt}
	<div class="fixed inset-0 z-9998 flex items-center justify-center p-6" transition:fade={{ duration: 300 }}>
		<div
			class="absolute inset-0 cursor-pointer bg-black/70"
			onclick={dismissPrompt}
			onkeydown={(e) => e.key === 'Escape' && dismissPrompt()}
			role="button"
			tabindex="0"
			aria-label="Close tutorial prompt"
		></div>
		<div class="relative w-full max-w-120" transition:fly={{ y: 30, duration: 400 }}>
			<div class="border-brand/50 rounded-xl border bg-zinc-900 p-6">
				<div class="max-h-[90vh] px-4">
					<div class="mb-8 text-center">
						<div class="mb-4 text-5xl max-sm:text-4xl">🎨</div>
						<h2 class="text-brand mb-2 text-2xl leading-tight font-bold max-sm:text-xl">
							Welcome to Image to Palette!
						</h2>
						<p class="m-0 text-base leading-relaxed text-zinc-400">Extract beautiful color palettes from any image</p>
					</div>

					<div class="mb-8 flex flex-col gap-4">
						<div
							class="feature-item hover:border-brand/50 flex items-center gap-4 rounded-md border border-zinc-600 bg-zinc-800/50 p-4 transition-[background-color,border-color] duration-300 hover:bg-white/5 max-sm:gap-3 max-sm:p-3"
						>
							<div class="shrink-0 text-2xl max-sm:text-xl">📸</div>
							<div class="flex flex-col gap-0.5">
								<strong class="text-sm font-semibold text-zinc-300">Upload Images</strong>
								<span class="text-xs leading-tight text-zinc-400">Drag & drop or click to upload</span>
							</div>
						</div>

						<div
							class="feature-item hover:border-brand/50 flex items-center gap-4 rounded-md border border-zinc-600 bg-zinc-800/50 p-4 transition-[background-color,border-color] duration-300 hover:bg-white/5 max-sm:gap-3 max-sm:p-3"
						>
							<div class="shrink-0 text-2xl max-sm:text-xl">🎯</div>
							<div class="flex flex-col gap-0.5">
								<strong class="text-sm font-semibold text-zinc-300">Select Areas</strong>
								<span class="text-xs leading-tight text-zinc-400">Choose specific regions for color extraction</span>
							</div>
						</div>

						<div
							class="feature-item hover:border-brand/50 flex items-center gap-4 rounded-md border border-zinc-600 bg-zinc-800/50 p-4 transition-[background-color,border-color] duration-300 hover:bg-white/5 max-sm:gap-3 max-sm:p-3"
						>
							<div class="shrink-0 text-2xl max-sm:text-xl">🎨</div>
							<div class="flex flex-col gap-0.5">
								<strong class="text-sm font-semibold text-zinc-300">Extract Colors</strong>
								<span class="text-xs leading-tight text-zinc-400">Get perfect palettes instantly</span>
							</div>
						</div>

						<div
							class="feature-item hover:border-brand/50 flex items-center gap-4 rounded-md border border-zinc-600 bg-zinc-800/50 p-4 transition-[background-color,border-color] duration-300 hover:bg-white/5 max-sm:gap-3 max-sm:p-3"
						>
							<div class="shrink-0 text-2xl max-sm:text-xl">💾</div>
							<div class="flex flex-col gap-0.5">
								<strong class="text-sm font-semibold text-zinc-300">Save & Apply</strong>
								<span class="text-xs leading-tight text-zinc-400">Save palettes and apply to new images</span>
							</div>
						</div>
					</div>

					<div class="mb-6 flex flex-col gap-3">
						<button
							class="hover:shadow-brand-lg bg-brand hover:bg-brand-hover flex cursor-pointer items-center justify-center gap-2 rounded-md border-0 px-6 py-3.5 font-semibold text-zinc-800 transition-[background-color,box-shadow] duration-300 focus:outline-none"
							onclick={startTutorial}
						>
							<span>Take the Tour</span>
							<span class="text-xs font-normal opacity-80">(2 min)</span>
						</button>

						<button
							class="hover:border-brand/50 cursor-pointer rounded-md border border-zinc-600 bg-transparent px-6 py-3.5 font-semibold text-zinc-400 outline-0 transition-[background-color,border-color,color] duration-300 hover:bg-white/5 hover:text-zinc-300"
							onclick={dismissPrompt}
						>
							Skip for now
						</button>
					</div>

					<div class="border-t border-zinc-600 pt-4 text-center">
						<p class="m-0 text-xs text-zinc-500">You can always restart the tutorial from the settings menu</p>
					</div>
				</div>
			</div>

			<button
				class="absolute top-4 right-11.5 flex h-8 w-8 cursor-pointer items-center justify-center rounded-md border border-zinc-600 text-zinc-400 outline-0 transition-[background-color,border-color,color] duration-300 hover:border-zinc-500 hover:bg-white/5 hover:text-zinc-300"
				onclick={dismissPrompt}
				aria-label="Close tutorial prompt"
			>
				<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<line x1="18" y1="6" x2="6" y2="18"></line>
					<line x1="6" y1="6" x2="18" y2="18"></line>
				</svg>
			</button>
		</div>
	</div>
{/if}

<style>
	/* Staggered animations for feature items */
	.feature-item {
		animation: slideIn 0.3s ease forwards;
	}

	.feature-item:nth-child(1) {
		animation-delay: 0.1s;
	}
	.feature-item:nth-child(2) {
		animation-delay: 0.3s;
	}
	.feature-item:nth-child(3) {
		animation-delay: 0.5s;
	}
	.feature-item:nth-child(4) {
		animation-delay: 0.7s;
	}

	@keyframes slideIn {
		from {
			opacity: 0;
			transform: translateY(20px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}
</style>
