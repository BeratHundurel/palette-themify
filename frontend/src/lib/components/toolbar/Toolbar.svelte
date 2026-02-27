<script lang="ts">
	import { cn } from '$lib/utils';
	import CopyOptions from './CopyOptions.svelte';
	import SelectorButton from './SelectorButton.svelte';
	import SavedPalettes from './SavedPalettes.svelte';
	import ApplyPaletteSettings from './ApplyPaletteSettings.svelte';
	import WallhavenSettings from './WallhavenSettings.svelte';
	import ThemeExport from './ThemeExport.svelte';
	import SavedThemes from './SavedThemes.svelte';
	import ApplyPaletteSettingsPopover from './popovers/ApplyPaletteSettingsPopover.svelte';
	import WallhavenSettingsPopover from './popovers/WallhavenSettingsPopover.svelte';
	import CopyOptionsPopover from './popovers/CopyOptionsPopover.svelte';
	import SavedPalettesPopover from './popovers/SavedPalettesPopover.svelte';
	import SavedThemesPopover from './popovers/SavedThemesPopover.svelte';
	import Download from './Download.svelte';
	import { appStore } from '$lib/stores/app.svelte';
	import { popoverStore } from '$lib/stores/popovers.svelte';

	// === Drag State ===
	let right = $state(25);
	let top = $state(100);
	let moving = $state(false);
	let dragHandle = $state<HTMLElement | undefined>(undefined);

	function handleMouseDown(e: MouseEvent) {
		if (dragHandle && dragHandle.contains(e.target as Node)) {
			moving = true;
			e.preventDefault();
		}
	}

	function handleMouseMove(e: MouseEvent) {
		if (moving) {
			right -= e.movementX;
			top += e.movementY;
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
	style="right: {right}px; top: {top}px;"
	class={cn('fixed select-none', moving ? 'cursor-move **:pointer-events-none' : '')}
>
	<div
		class={cn(
			'border-brand/50 shadow-brand/20 rounded-xl border bg-zinc-900 shadow-2xl',
			'hover:shadow-brand/40 hover:border-brand/50 hover:has-[.palette-dropdown-base]:border-brand/50 hover:has-[.palette-dropdown-base]:shadow-none',
			'transition-[border-color,box-shadow] duration-300 ease-out'
		)}
	>
		<div
			bind:this={dragHandle}
			class={cn(
				'flex cursor-move items-center justify-center rounded-t-xl border-b border-zinc-700/50 bg-zinc-800/30 px-6 py-4',
				'hover:border-brand/50 hover:bg-zinc-800/50',
				'transition-[background-color,border-color] duration-300 ease-out'
			)}
		>
			<div class="flex flex-col items-center gap-2">
				<div
					class={cn(
						'h-0.5 w-8 rounded-full transition-[background-color,box-shadow] duration-300 ease-out',
						moving ? 'bg-brand shadow-brand' : 'bg-zinc-500'
					)}
				></div>
				<div
					class={cn(
						'h-0.5 w-6 rounded-full transition-[background-color] duration-300 ease-out',
						moving ? 'bg-brand/60' : 'bg-zinc-500/60'
					)}
				></div>
			</div>
		</div>

		<div class="px-6 py-6">
			<div class="flex flex-col space-y-8">
				{#if appStore.state.selectors.length > 0}
					<div class="space-y-3">
						<div class="flex items-center gap-3">
							<h3 class="text-brand text-sm font-semibold tracking-wide uppercase">Selection Tools</h3>
							<div class="from-brand/30 h-px flex-1 bg-linear-to-r to-transparent"></div>
						</div>
						<div class="flex flex-wrap justify-start gap-3">
							{#each appStore.state.selectors as selector, i (selector.id)}
								<SelectorButton {selector} index={i} />
							{/each}
						</div>
					</div>
				{/if}

				<div class="space-y-6">
					<div class="space-y-3">
						<div class="flex items-center gap-3">
							<h3 class="text-brand text-sm font-semibold tracking-wide uppercase">Themes & Processing</h3>
							<div class="from-brand/30 h-px flex-1 bg-linear-to-r to-transparent"></div>
						</div>
						<div class="flex gap-3">
							<ThemeExport />
							<SavedPalettes />
							<SavedThemes />
						</div>
					</div>

					<div class="space-y-3">
						<div class="flex items-center gap-3">
							<h3 class="text-brand text-sm font-semibold tracking-wide uppercase">Settings</h3>
							<div class="from-brand/30 h-px flex-1 bg-linear-to-r to-transparent"></div>
						</div>
						<div class="flex gap-3">
							<WallhavenSettings />
							<ApplyPaletteSettings />
						</div>
					</div>

					<div class="space-y-6">
						<div class="space-y-3">
							<div class="flex items-center gap-3">
								<h3 class="text-brand text-sm font-semibold tracking-wide uppercase">Copy & Export</h3>
								<div class="from-brand/30 h-px flex-1 bg-linear-to-r to-transparent"></div>
							</div>
							<div class="flex gap-3">
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
