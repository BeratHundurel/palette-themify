<script lang="ts">
	import { cn } from '$lib/utils';
	import ActionPillButton from '$lib/components/ui/ActionPillButton.svelte';
	import ColorSwatch from '$lib/components/ui/ColorSwatch.svelte';
	import DangerTextButton from '$lib/components/ui/DangerTextButton.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import IconDangerButton from '$lib/components/ui/IconDangerButton.svelte';
	import { appStore } from '$lib/stores/app/store.svelte';
	import { authStore } from '$lib/stores/auth.svelte';
	import { dialogStore } from '$lib/stores/dialog.svelte';
	import { popoverStore } from '$lib/stores/popovers.svelte';
	import { tutorialStore } from '$lib/stores/tutorial.svelte';
	import { TUTORIAL_APPLY_PALETTE, TUTORIAL_PALETTE_NAME } from '$lib/types/tutorial';
	import toast from 'svelte-french-toast';
	import type { Color } from '$lib/types/color';
	import type { PaletteData } from '$lib/types/palette';

	const tutorialPaletteItem: PaletteData = {
		id: 'tutorial_palette_preview',
		name: TUTORIAL_PALETTE_NAME,
		palette: TUTORIAL_APPLY_PALETTE,
		createdAt: new Date(0).toISOString(),
		isSystem: true,
		isShared: false
	};

	let displayPalettes = $derived.by(() => {
		if (tutorialStore.state.isActive && tutorialStore.getCurrentStep()?.id === 'apply-palette') {
			return [tutorialPaletteItem, ...appStore.state.savedPalettes];
		}

		return appStore.state.savedPalettes.filter((item) => item.id !== tutorialPaletteItem.id);
	});

	function handlePaletteLoad(palette: Color[]) {
		if (!appStore.state.imageLoaded) {
			toast.error('Load an image before applying a palette.');
			return;
		}

		appStore.applyPalette(palette);
		if (!(tutorialStore.state.isActive && tutorialStore.getCurrentStep()?.id === 'apply-palette')) {
			popoverStore.close('saved');
		}
		tutorialStore.setSavedPaletteApplied(true);
	}

	function handleCreateThemeFromPalette(palette: Color[]) {
		if (!palette || palette.length === 0) {
			toast.error('This palette has no colors to generate a theme.');
			return;
		}

		appStore.resetThemeExportSession();
		appStore.state.themeExport.backupColors = palette.map((color) => ({ hex: color.hex }));

		popoverStore.close('saved');
		popoverStore.state.current = 'themeExport';
	}

	async function handlePaletteDelete(paletteId: string, paletteName: string) {
		const shouldDelete = await dialogStore.confirm({
			title: 'Delete saved palette?',
			message: `Are you sure you want to delete "${paletteName}"?`,
			confirmLabel: 'Delete palette',
			variant: 'danger'
		});
		if (!shouldDelete) return;

		await appStore.deletePalette(paletteId);
	}

	async function handleDeleteAllPalettes() {
		const palettesToDelete = displayPalettes.filter((item) => !item.isSystem);
		if (palettesToDelete.length === 0) return;

		const shouldDeleteAll = await dialogStore.confirm({
			title: 'Delete all saved palettes?',
			message: `This will permanently delete ${palettesToDelete.length} saved palette${palettesToDelete.length === 1 ? '' : 's'}.`,
			confirmLabel: 'Delete all',
			variant: 'danger'
		});
		if (!shouldDeleteAll) return;

		await appStore.deletePalettes(palettesToDelete.map((palette) => palette.id));
	}

	function isLocalPalette(id: string): boolean {
		return id.startsWith('local_');
	}

	async function handlePaletteShareToggle(item: PaletteData) {
		if (item.isSystem) {
			toast.error('System palettes cannot be shared. Save a copy first.');
			return;
		}

		if (isLocalPalette(item.id)) {
			toast.error(
				authStore.state.isAuthenticated
					? 'Sync this palette first, then share it.'
					: 'Sign in to sync this palette before sharing.'
			);
			return;
		}

		await appStore.setPaletteShared(item.id, !item.isShared);
	}
</script>

<div
	class={cn(
		'palette-dropdown-base w-80 max-w-[90vw]',
		popoverStore.state.direction === 'right' ? 'left-full ml-2' : 'right-full mr-2'
	)}
	style={`min-width: 260px; ${popoverStore.state.direction === 'right' ? 'left: calc(100% + 0.5rem);' : 'right: calc(100% + 0.5rem);'}`}
	role="dialog"
	aria-labelledby="saved-palettes-title"
	tabindex="-1"
>
	<div class="mb-3 flex items-center justify-between">
		<h3 id="saved-palettes-title" class="text-brand text-xs font-medium">Saved Palettes</h3>

		<DangerTextButton
			onclick={handleDeleteAllPalettes}
			title="Delete all saved palettes"
			disabled={displayPalettes.filter((item) => !item.isSystem).length === 0}
		>
			Delete all
		</DangerTextButton>
	</div>
	<div class="scrollable-content custom-scrollbar max-h-72 overflow-y-auto">
		{#if displayPalettes.length === 0}
			<EmptyState>
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
			</EmptyState>
		{:else}
			<ul class="flex flex-col gap-3">
				{#each displayPalettes as item, i (item.id + '-' + i)}
					<li
						class="hover:border-brand/50 group relative overflow-hidden rounded-lg border border-zinc-600 bg-zinc-800/50 transition-[background-color,border-color,box-shadow] duration-300 hover:bg-white/5 {item.id ===
						tutorialPaletteItem.id
							? 'tutorial-palette-apply border-brand/60 bg-brand/10'
							: ''}"
					>
						<div class="p-3">
							<div class="mb-3 min-w-0">
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

							<div class="mb-3 flex items-center justify-around gap-1">
								<ActionPillButton
									onclick={() => handlePaletteLoad(item.palette)}
									class="gap-1 px-2"
									title="Apply palette"
								>
									<svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
									</svg>
									Apply
								</ActionPillButton>

								<ActionPillButton
									onclick={() => handleCreateThemeFromPalette(item.palette)}
									class="gap-1 px-2"
									title="Create theme from palette"
								>
									<svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M13 10V3L4 14h7v7l9-11h-7z"
										/>
									</svg>
									Theme
								</ActionPillButton>

								<ActionPillButton
									onclick={() => handlePaletteShareToggle(item)}
									class="px-2"
									title={item.isSystem || isLocalPalette(item.id)
										? 'Sign in synced palette required to share'
										: item.isShared
											? 'Remove from shared list'
											: 'Share palette publicly'}
								>
									{item.isShared ? 'Unshare' : 'Share'}
								</ActionPillButton>

								{#if !item.isSystem}
									<IconDangerButton
										onclick={() => handlePaletteDelete(item.id, item.name)}
										class="shrink-0 p-1"
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
									</IconDangerButton>
								{/if}
							</div>

							<div class="relative mb-2">
								<div class="flex flex-wrap gap-1.5">
									{#each item.palette as color, index (item.id + '-' + index)}
										{#if index < 12}
											<div class="group/swatch relative">
												<ColorSwatch color={color.hex} title={color.hex} />
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
							<div class="flex items-center justify-between gap-2">
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
								{#if item.isShared}
									<span
										class="rounded-full border border-emerald-400/40 bg-emerald-500/15 px-2 py-0.5 text-[10px] font-semibold text-emerald-300"
									>
										Shared
									</span>
								{/if}
							</div>
						</div>
					</li>
				{/each}
			</ul>
		{/if}
	</div>
</div>
