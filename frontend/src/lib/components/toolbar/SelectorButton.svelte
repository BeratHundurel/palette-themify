<script lang="ts">
	import { appStore } from '$lib/stores/app.svelte';
	import { type Selector } from '$lib/types/palette';
	import { cn } from '$lib/utils';

	let { selector, index = 0 } = $props<{
		selector: Selector;
		index?: number;
	}>();

	function handleClick() {
		appStore.state.activeSelectorId = selector.id;
		appStore.state.selectors = appStore.state.selectors.map((s) => ({
			...s,
			selected: s.id === selector.id
		}));
	}

	function handleRightClick(event: MouseEvent) {
		event.preventDefault();
		appStore.state.selectors = appStore.state.selectors.map((s) =>
			s.id === selector.id
				? {
						...s,
						selection: undefined
					}
				: s
		);
		appStore.redrawCanvas();
	}
</script>

<button
	class={cn(
		'group relative transition-all duration-300 ease-out',
		'h-11 w-11 overflow-hidden rounded-md',
		selector.selected
			? 'border-brand/50 shadow-brand/20 -translate-y-1 scale-105 border-2 shadow-lg'
			: 'border border-zinc-700/50  hover:border-zinc-600 hover:shadow-md active:scale-95'
	)}
	onclick={handleClick}
	oncontextmenu={handleRightClick}
	aria-label="Selector {index + 1}"
	type="button"
>
	<div
		class="absolute inset-0 transition-all duration-300"
		style="background: linear-gradient(135deg, {selector.color} 0%, {selector.color} 100%)"
	>
		<div class="absolute inset-0 bg-gradient-to-br from-white/5 to-transparent"></div>

		<div class="absolute inset-0 bg-black/10 transition-all duration-300 group-hover:bg-black/5"></div>
	</div>

	{#if selector.selected}
		<div class="relative z-10 flex h-full w-full items-center justify-center">
			<div class="rounded-full bg-zinc-900/60 p-1.5 shadow-lg ring-1 ring-white/20 backdrop-blur-sm">
				<svg xmlns="http://www.w3.org/2000/svg" height="14px" viewBox="0 -960 960 960" width="14px" fill="white">
					<path
						d="m424-296 282-282-56-56-226 226-114-114-56 56 170 170Zm56 216q-83 0-156-31.5T197-197q-54-54-85.5-127T80-480q0-83 31.5-156T197-763q54-54 127-85.5T480-880q83 0 156 31.5T763-763q54 54 85.5 127T880-480q0 83-31.5 156T763-197q-54 54-127 85.5T480-80Zm0-80q134 0 227-93t93-227q0-134-93-227t-227-93q-134 0-227 93t-93 227q0 134 93 227t227 93Zm0-320Z"
					/>
				</svg>
			</div>
		</div>
	{/if}

	<div
		class={cn(
			'absolute -top-1 -right-1 flex h-5 w-5 items-center justify-center rounded-full',
			'border bg-zinc-900 text-xs font-semibold shadow-md transition-all duration-300',
			selector.selected
				? 'border-brand/50 text-brand'
				: 'border-zinc-700 text-zinc-400 group-hover:border-zinc-600 group-hover:text-zinc-300'
		)}
	>
		{index + 1}
	</div>

	<!-- Hover shimmer effect -->
	<div
		class="absolute inset-0 bg-gradient-to-br from-white/0 via-white/10 to-white/0 opacity-0 transition-opacity duration-300 group-hover:opacity-100"
		style="transform: translateX(-100%); animation: shimmer 2s infinite;"
	></div>
</button>

<style>
	@keyframes shimmer {
		0% {
			transform: translateX(-100%) translateY(-100%) rotate(45deg);
		}
		100% {
			transform: translateX(100%) translateY(100%) rotate(45deg);
		}
	}
</style>
