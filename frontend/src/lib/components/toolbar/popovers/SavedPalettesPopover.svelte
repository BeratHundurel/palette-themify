<script lang="ts">
	import { cn } from '$lib/utils';
	import { appStore } from '$lib/stores/app.svelte';
	import { popoverStore } from '$lib/stores/popovers.svelte';
	import { tutorialStore } from '$lib/stores/tutorial.svelte';
	import toast from 'svelte-french-toast';
	import type { Color } from '$lib/types/color';

	function handlePaletteLoad(palette: Color[]) {
		if (!appStore.state.imageLoaded) {
			toast.error('Load an image before applying a palette.');
			return;
		}

		appStore.applyPalette(palette);
		popoverStore.close('saved');
		tutorialStore.setSavedPaletteApplied(true);
	}

	async function handlePaletteDelete(paletteId: string, paletteName: string) {
		if (confirm(`Are you sure you want to delete "${paletteName}"?`)) {
			await appStore.deletePalette(paletteId);
		}
	}
</script>

<div
	class={cn(
		'palette-dropdown-base w-80',
		popoverStore.state.direction === 'right' ? 'left-full ml-2' : 'right-full mr-2'
	)}
	style={`min-width: 260px; ${popoverStore.state.direction === 'right' ? 'left: calc(100% + 0.5rem);' : 'right: calc(100% + 0.5rem);'}`}
	role="dialog"
	aria-labelledby="saved-palettes-title"
	tabindex="-1"
>
	<h3 id="saved-palettes-title" class="text-brand mb-3 text-xs font-medium">Saved Palettes</h3>
	<div class="scrollable-content custom-scrollbar max-h-72 overflow-y-auto">
		{#if appStore.state.savedPalettes.length === 0}
			<div class="flex flex-col items-center justify-center py-12 text-center">
				<svg class="mb-3 h-12 w-12 text-zinc-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="1.5"
						d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17h.01"
					/>
				</svg>
				<p class="text-sm text-zinc-400">No saved palettes yet</p>
				<p class="mt-1 text-xs text-zinc-500">Save your current palette to see it here</p>
			</div>
		{:else}
			<ul class="flex flex-col gap-3">
				{#each appStore.state.savedPalettes as item, i (i)}
					<li
						class="hover:border-brand/50 group relative overflow-hidden rounded-lg border border-zinc-600 bg-zinc-800/50 transition-[background-color,border-color,box-shadow] duration-300 hover:bg-white/5"
					>
						<div class="p-3">
							<!-- Header -->
							<div class="mb-3 flex items-start justify-between">
								<div class="min-w-0 flex-1">
									<div class="flex items-center gap-2">
										<h4 class="text-brand truncate font-mono text-sm font-semibold" title={item.name}>
											{item.name}
										</h4>
										{#if item.isSystem}
											<span
												class="shrink-0 rounded-full bg-purple-500/20 px-2 py-0.5 text-xs font-medium text-purple-300"
											>
												System
											</span>
										{/if}
									</div>
									<div class="mt-1 flex items-center gap-1.5 text-xs text-zinc-400">
										<svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17h.01"
											/>
										</svg>
										<span>{item.palette.length} {item.palette.length === 1 ? 'color' : 'colors'}</span>
									</div>
								</div>

								<!-- Actions -->
								<div class="flex items-center gap-1.5">
									<button
										class="text-brand hover:bg-brand/10 flex items-center gap-1.5 rounded-md px-2.5 text-xs font-medium transition-[transform,background-color] hover:scale-105"
										onclick={() => handlePaletteLoad(item.palette)}
										type="button"
										title="Apply palette"
									>
										<svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
										</svg>
										Apply
									</button>

									{#if !item.isSystem}
										<button
											class="flex items-center gap-1 rounded-md p-1.5 text-zinc-500 transition-[transform,background-color,color] hover:scale-110 hover:bg-red-500/10 hover:text-red-400"
											onclick={() => handlePaletteDelete(item.id, item.name)}
											type="button"
											title="Delete palette"
										>
											<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path
													stroke-linecap="round"
													stroke-linejoin="round"
													stroke-width="2"
													d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
												/>
											</svg>
										</button>
									{/if}
								</div>
							</div>

							<div class="relative mb-2">
								<div class="flex flex-wrap gap-1.5">
									{#each item.palette as color, index (item.id + '-' + index)}
										{#if index < 12}
											<div class="group/swatch relative">
												<span
													class="inline-block h-8 w-8 cursor-pointer rounded-md border border-zinc-700/50 shadow-md transition-[transform,box-shadow,border-color] duration-300 hover:scale-105 hover:ring-2 hover:ring-white/50"
													style="background-color: {color.hex}"
													title={color.hex}
												></span>
											</div>
										{/if}
									{/each}
									{#if item.palette.length > 12}
										<div
											class="flex h-8 w-8 items-center justify-center rounded-md border border-zinc-600 bg-zinc-900/50 text-xs font-medium text-zinc-400"
										>
											+{item.palette.length - 12}
										</div>
									{/if}
								</div>
							</div>
						</div>

						<div class="group-hover:border-brand/50 group border-t border-zinc-600 bg-zinc-900/50 px-3 py-1.5">
							<span class="flex items-center gap-1.5 text-xs text-zinc-500">
								<svg class="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
									/>
								</svg>
								{new Date(item.createdAt).toLocaleDateString('en-US', {
									month: 'short',
									day: 'numeric',
									year: 'numeric'
								})}
							</span>
						</div>
					</li>
				{/each}
			</ul>
		{/if}
	</div>
</div>

<style>
</style>
