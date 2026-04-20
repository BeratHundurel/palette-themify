<script lang="ts">
	import { tutorialStore } from '$lib/stores/tutorial.svelte';
	import { fade, fly } from 'svelte/transition';
	import { onMount } from 'svelte';

	const PROMPT_FADE_DURATION_MS = 200;
	const PANEL_FLY_DURATION_MS = 240;
	const START_DELAY_MS = 180;

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
		}, START_DELAY_MS);
	}

	function dismissPrompt() {
		showPrompt = false;
		localStorage.setItem('tutorial-skipped', 'true');
	}
</script>

{#if showPrompt}
	<div
		class="fixed inset-0 z-10003 flex items-center justify-center p-6"
		transition:fade={{ duration: PROMPT_FADE_DURATION_MS }}
	>
		<div
			class="absolute inset-0 cursor-pointer bg-black/70"
			onclick={dismissPrompt}
			onkeydown={(e) => e.key === 'Escape' && dismissPrompt()}
			role="button"
			tabindex="0"
			aria-label="Close tutorial prompt"
		></div>
		<div
			class="prompt-panel custom-scrollbar relative w-full max-w-120"
			transition:fly={{ y: 24, duration: PANEL_FLY_DURATION_MS }}
		>
			<div class="border-brand/50 rounded-xl border bg-zinc-900 p-6">
				<div class="max-h-[90svh] overflow-y-auto px-4">
					<div class="mb-8 text-center">
						<div class="mb-4 text-5xl">🎨</div>
						<h2 class="text-brand mb-2 text-2xl leading-tight font-bold">Welcome to ThemeSmith!</h2>
						<p class="m-0 text-base leading-relaxed text-zinc-400">Extract beautiful color palettes from any image</p>
					</div>

					<div class="mb-8 flex flex-col gap-4">
						<div
							class="feature-item hover:border-brand/50 flex items-center gap-4 rounded-md border border-zinc-600 bg-zinc-800/50 p-4 transition-[background-color,border-color] duration-300 hover:bg-white/5"
						>
							<div class="shrink-0 text-2xl">📸</div>
							<div class="flex flex-col gap-0.5">
								<strong class="text-sm font-semibold text-zinc-300">Upload Images</strong>
								<span class="text-xs leading-tight text-zinc-400">Drag & drop or click to upload</span>
							</div>
						</div>

						<div
							class="feature-item hover:border-brand/50 flex items-center gap-4 rounded-md border border-zinc-600 bg-zinc-800/50 p-4 transition-[background-color,border-color] duration-300 hover:bg-white/5"
						>
							<div class="shrink-0 text-2xl">🎯</div>
							<div class="flex flex-col gap-0.5">
								<strong class="text-sm font-semibold text-zinc-300">Select Areas</strong>
								<span class="text-xs leading-tight text-zinc-400">Choose specific regions for color extraction</span>
							</div>
						</div>

						<div
							class="feature-item hover:border-brand/50 flex items-center gap-4 rounded-md border border-zinc-600 bg-zinc-800/50 p-4 transition-[background-color,border-color] duration-300 hover:bg-white/5"
						>
							<div class="shrink-0 text-2xl">🎨</div>
							<div class="flex flex-col gap-0.5">
								<strong class="text-sm font-semibold text-zinc-300">Extract Colors</strong>
								<span class="text-xs leading-tight text-zinc-400">Get perfect palettes instantly</span>
							</div>
						</div>

						<div
							class="feature-item hover:border-brand/50 flex items-center gap-4 rounded-md border border-zinc-600 bg-zinc-800/50 p-4 transition-[background-color,border-color] duration-300 hover:bg-white/5"
						>
							<div class="shrink-0 text-2xl">🧩</div>
							<div class="flex flex-col gap-0.5">
								<strong class="text-sm font-semibold text-zinc-300">Theme Inspector</strong>
								<span class="text-xs leading-tight text-zinc-400"
									>Generate, tweak, and export VS Code or Zed themes from your palette</span
								>
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
	.prompt-panel {
		will-change: opacity, transform;
	}

	@media (prefers-reduced-motion: reduce) {
		.prompt-panel {
			will-change: auto;
		}
	}
</style>
