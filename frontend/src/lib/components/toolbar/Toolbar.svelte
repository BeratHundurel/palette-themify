<script lang="ts">
	import { cn } from '$lib/utils';
	import CopyOptions from './CopyOptions.svelte';
	import OpenInCoolors from './OpenInCoolors.svelte';
	import SelectorButton from './SelectorButton.svelte';
	import SavedPalettes from './SavedPalettes.svelte';
	import SavedWorkspaces from './SavedWorkspaces.svelte';
	import ApplicationSettings from './ApplicationSettings.svelte';
	import WallhavenSettings from './WallhavenSettings.svelte';
	import ThemeExport from './ThemeExport.svelte';
	import ApplicationSettingsPopover from './popovers/ApplicationSettingsPopover.svelte';
	import WallhavenSettingsPopover from './popovers/WallhavenSettingsPopover.svelte';
	import CopyOptionsPopover from './popovers/CopyOptionsPopover.svelte';
	import SavedPalettesPopover from './popovers/SavedPalettesPopover.svelte';
	import SavedWorkspacesPopover from './popovers/SavedWorkspacesPopover.svelte';
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
	class={cn('fixed select-none', moving ? 'cursor-move [&_*]:pointer-events-none' : '')}
>
	<div
		class={cn(
			'border-brand/50 rounded-md border bg-zinc-900',
			'hover:shadow-brand hover:border-brand hover:has-[.palette-dropdown-base]:border-brand/50 hover:has-[.palette-dropdown-base]:shadow-none',
			'transition-all duration-300 ease-out'
		)}
	>
		<div
			bind:this={dragHandle}
			class={cn(
				'flex cursor-move items-center justify-center rounded-md border-b border-zinc-700 px-5 py-4',
				'hover:border-brand/50 hover:bg-zinc-800/50',
				'transition-all duration-300 ease-out'
			)}
		>
			<div class="flex flex-col items-center gap-1.5">
				<div
					class={cn(
						'h-0.5 w-8 rounded-full transition-all duration-300 ease-out',
						moving ? 'bg-brand/80 shadow-brand' : 'bg-zinc-400/80'
					)}
				></div>
				<div
					class={cn(
						'h-0.5 w-6 rounded-full transition-all duration-300 ease-out',
						moving ? 'bg-brand/50 shadow-brand' : 'bg-zinc-400/50'
					)}
				></div>
			</div>
		</div>

		<div class="relative p-5">
			<ul class="flex flex-col gap-3">
				{#if appStore.state.selectors.length > 0}
					<li class="relative mb-4">
						<div class="mb-3 flex items-center gap-2">
							<h3 class="text-brand text-xs font-medium uppercase">Selection Tools</h3>
							<div class="from-brand/50 h-px flex-1 bg-gradient-to-r to-transparent"></div>
						</div>
						<div class="flex flex-wrap justify-center gap-2">
							{#each appStore.state.selectors as selector, i (selector.id)}
								<SelectorButton {selector} index={i} />
							{/each}
						</div>
					</li>
				{/if}

				<li class="relative mb-4">
					<div class="mb-2 flex items-center gap-2">
						<h3 class="text-brand text-xs font-medium uppercase">Image Sources</h3>
						<div class="from-brand/50 h-px flex-1 bg-gradient-to-r to-transparent"></div>
					</div>
					<div class="flex justify-start gap-2">
						<WallhavenSettings />
					</div>
				</li>

				<li class="relative mb-4">
					<div class="mb-2 flex items-center gap-2">
						<h3 class="text-brand text-xs font-medium uppercase">Processing</h3>
						<div class="from-brand/50 h-px flex-1 bg-gradient-to-r to-transparent"></div>
					</div>
					<div class="flex justify-start gap-2">
						<SavedWorkspaces />
						<SavedPalettes />
						<ApplicationSettings />
					</div>
				</li>

				<li class="relative mb-4">
					<div class="mb-2 flex items-center gap-2">
						<h3 class="text-brand text-xs font-medium uppercase">Copy</h3>
						<div class="from-brand/50 h-px flex-1 bg-gradient-to-r to-transparent"></div>
					</div>
					<div class="flex justify-start gap-2">
						<CopyOptions />
						<OpenInCoolors />
					</div>
				</li>

				<li class="relative mb-4">
					<div class="mb-2 flex items-center gap-2">
						<h3 class="text-brand text-xs font-medium uppercase">Theme Generation</h3>
						<div class="from-brand/50 h-px flex-1 bg-gradient-to-r to-transparent"></div>
					</div>
					<ThemeExport />
				</li>

				<li class="relative">
					<div class="mb-2 flex items-center gap-2">
						<h3 class="text-brand text-xs font-medium uppercase">Export</h3>
						<div class="from-brand/50 h-px flex-1 bg-gradient-to-r to-transparent"></div>
					</div>
					<Download />
				</li>
			</ul>
		</div>

		{#if popoverStore.state.current === 'application'}
			<ApplicationSettingsPopover />
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

		{#if popoverStore.state.current === 'workspaces'}
			<SavedWorkspacesPopover />
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
		border-color: rgba(161, 161, 170, 0.6);
		background-color: rgba(39, 39, 42, 0.9);
		box-shadow:
			0 10px 15px -3px rgba(0, 0, 0, 0.1),
			0 4px 6px -2px rgba(0, 0, 0, 0.05);
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

	/* Enhanced visual feedback for sections */
	li:hover .from-brand\/50 {
		background: linear-gradient(to right, rgba(255, 175, 120, 0.785), transparent);
	}
</style>
