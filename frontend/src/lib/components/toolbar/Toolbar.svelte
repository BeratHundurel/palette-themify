<script lang="ts">
	import { onMount } from 'svelte';
	import { cn } from '$lib/utils';
	import CopyOptions from './copy-options/CopyOptions.svelte';
	import SelectorButton from './SelectorButton.svelte';
	import SavedPalettes from './saved-palettes/SavedPalettes.svelte';
	import ApplyPaletteSettings from './apply-palette-settings/ApplyPaletteSettings.svelte';
	import WallhavenSettings from './wallhaven-settings/WallhavenSettings.svelte';
	import ThemeExport from './theme-export/ThemeExport.svelte';
	import SavedThemes from './saved-themes/SavedThemes.svelte';
	import ApplyPaletteSettingsPopover from './apply-palette-settings/ApplyPaletteSettingsPopover.svelte';
	import WallhavenSettingsPopover from './wallhaven-settings/WallhavenSettingsPopover.svelte';
	import CopyOptionsPopover from './copy-options/CopyOptionsPopover.svelte';
	import SavedPalettesPopover from './saved-palettes/SavedPalettesPopover.svelte';
	import SavedThemesPopover from './saved-themes/SavedThemesPopover.svelte';
	import Download from './download/Download.svelte';
	import { appStore } from '$lib/stores/app/store.svelte';
	import { popoverStore } from '$lib/stores/popovers.svelte';

	// === Drag State ===
	let right = $state(20);
	let top = $state(80);
	let moving = $state(false);
	let dragHandle = $state<HTMLElement | undefined>(undefined);
	let toolbarSection = $state<HTMLElement | undefined>(undefined);
	let isPositioned = $state(false);

	function setInitialPosition() {
		if (!toolbarSection || isPositioned) {
			return;
		}

		const rect = toolbarSection.getBoundingClientRect();
		const viewportWidth = window.innerWidth;
		const viewportHeight = window.innerHeight;

		if (viewportWidth < 768) {
			right = Math.max(10, (viewportWidth - rect.width) / 2);
		} else {
			right = 20;
		}

		top = Math.max(80, (viewportHeight - rect.height) / 2);
		isPositioned = true;
	}

	onMount(() => {
		setInitialPosition();
	});

	function handleMouseDown(e: MouseEvent) {
		if (dragHandle && dragHandle.contains(e.target as Node)) {
			moving = true;
			e.preventDefault();
		}
	}

	function handleMouseMove(e: MouseEvent) {
		if (moving && toolbarSection) {
			const rect = toolbarSection.getBoundingClientRect();
			const newRight = right - e.movementX;
			const newTop = top + e.movementY;

			// Calculate boundaries (with 10px padding from edges)
			const minRight = 10;
			const maxRight = window.innerWidth - rect.width - 10;
			const minTop = 10;
			const maxTop = window.innerHeight - rect.height - 10;

			// Constrain to viewport boundaries
			right = Math.max(minRight, Math.min(newRight, maxRight));
			top = Math.max(minTop, Math.min(newTop, maxTop));
		}
	}

	$effect(() => {
		if (moving && popoverStore.state.current) {
			popoverStore.close();
		}
	});
</script>

<svelte:window on:mouseup={() => (moving = false)} on:mousemove={handleMouseMove} />

<section
	role="toolbar"
	tabindex="0"
	onmousedown={handleMouseDown}
	style="right: {right}px; top: {top}px; visibility: {isPositioned ? 'visible' : 'hidden'};"
	class={cn('fixed select-none', moving ? 'cursor-move **:pointer-events-none' : '')}
	bind:this={toolbarSection}
>
	<div
		class="border-brand/50 shadow-brand/20 hover:shadow-brand/40 hover:border-brand/50 hover:has-[.palette-dropdown-base]:border-brand/50 max-w-lg rounded-xl
			border bg-zinc-900 shadow-2xl transition-[border-color,box-shadow]
			duration-300 ease-out hover:has-[.palette-dropdown-base]:shadow-none"
	>
		<div
			bind:this={dragHandle}
			class="hover:border-brand/50 flex cursor-move items-center justify-center rounded-t-xl border-b border-zinc-700/50 bg-zinc-800/30 px-4
				py-2.5 transition-[background-color,border-color]
				duration-300 ease-out hover:bg-zinc-800/50"
		>
			<div class="flex flex-col items-center gap-1.5">
				<div
					class={cn(
						'h-0.5 w-7 rounded-full transition-[background-color,box-shadow] duration-300 ease-out',
						moving ? 'bg-brand shadow-brand' : 'bg-zinc-500'
					)}
				></div>
				<div
					class={cn(
						'h-0.5 w-5 rounded-full transition-[background-color] duration-300 ease-out',
						moving ? 'bg-brand/60' : 'bg-zinc-500/60'
					)}
				></div>
			</div>
		</div>

		<div class="px-4 py-4">
			<div class="flex flex-col space-y-5">
				{#if appStore.state.selectors.length > 0}
					<div class="space-y-2">
						<div class="flex items-center gap-2">
							<h3 class="text-brand text-xs font-semibold tracking-wide uppercase">Selection Tools</h3>
							<div class="from-brand/30 h-px flex-1 bg-linear-to-r to-transparent"></div>
						</div>
						<div class="flex flex-wrap justify-start gap-2">
							{#each appStore.state.selectors as selector, i (selector.id)}
								<SelectorButton {selector} index={i} />
							{/each}
						</div>
					</div>
				{/if}

				<div class="space-y-4">
					<div class="space-y-2">
						<div class="flex items-center gap-2">
							<h3 class="text-brand text-xs font-semibold tracking-wide uppercase">Themes & Processing</h3>
							<div class="from-brand/30 h-px flex-1 bg-linear-to-r to-transparent"></div>
						</div>
						<div class="flex gap-2">
							<ThemeExport />
							<SavedPalettes />
							<SavedThemes />
						</div>
					</div>

					<div class="space-y-2">
						<div class="flex items-center gap-2">
							<h3 class="text-brand text-xs font-semibold tracking-wide uppercase">Settings</h3>
							<div class="from-brand/30 h-px flex-1 bg-linear-to-r to-transparent"></div>
						</div>
						<div class="flex gap-2">
							<WallhavenSettings />
							<ApplyPaletteSettings />
						</div>
					</div>

					<div class="space-y-4">
						<div class="space-y-2">
							<div class="flex items-center gap-2">
								<h3 class="text-brand text-xs font-semibold tracking-wide uppercase">Copy & Export</h3>
								<div class="from-brand/30 h-px flex-1 bg-linear-to-r to-transparent"></div>
							</div>
							<div class="flex gap-2">
								<CopyOptions />
								<Download />
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>

		{#if popoverStore.state.current === 'application'}
			<ApplyPaletteSettingsPopover />
		{/if}

		{#if popoverStore.state.current === 'wallhaven'}
			<WallhavenSettingsPopover />
		{/if}

		{#if popoverStore.state.current === 'copy'}
			<CopyOptionsPopover />
		{/if}

		{#if popoverStore.state.current === 'saved'}
			<SavedPalettesPopover />
		{/if}

		{#if popoverStore.state.current === 'themes'}
			<SavedThemesPopover />
		{/if}
	</div>
</section>

<style>
	:global(.toolbar-button-base::before) {
		content: '';
		position: absolute;
		inset: 0;
		background: linear-gradient(45deg, transparent, rgba(255, 255, 255, 0.1), transparent);
		transform: translateX(-100%);
		transition: transform 0.6s;
	}

	:global(.toolbar-button-base:hover) {
		border-color: rgba(238, 179, 143, 0.5);
		background-color: rgba(39, 39, 42, 0.9);
		box-shadow: 0 4px 12px rgba(238, 179, 143, 0.15);
		transform: translateY(-1px);
	}

	:global(.toolbar-button-base:hover::before) {
		transform: translateX(100%);
	}

	:global(.toolbar-button-base:active) {
		transform: scale(0.95) translateY(0px);
		transition: transform 0.1s;
		box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.2);
	}

	@keyframes fadeInZoom {
		from {
			opacity: 0;
			transform: scale(0.95) translateY(-8px);
		}
		to {
			opacity: 1;
			transform: scale(1) translateY(0);
		}
	}

	.space-y-3:hover .from-brand\/30 {
		background: linear-gradient(to right, rgba(238, 179, 143, 0.6), transparent);
	}
</style>
