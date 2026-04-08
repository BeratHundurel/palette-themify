<script lang="ts">
	import { DEFAULT_SELECTOR_ID } from '$lib/types/selector';
	import { appStore } from '$lib/stores/app/store.svelte';
	import { tutorialStore } from '$lib/stores/tutorial.svelte';
	import { cn } from '$lib/utils';
	import { sortColorsByMethod, type SortMethod } from '$lib/colorUtils';
	import { tick } from 'svelte';
	import toast from 'svelte-french-toast';
	import { scale } from 'svelte/transition';

	let hoverX = $state(0);
	let buttonWidth = $state(0);
	let showHover = $state(false);
	let sortButtonGroup: HTMLDivElement | null = $state(null);

	function handleButtonHover(e: MouseEvent) {
		if (!sortButtonGroup) return;
		const button = e.currentTarget as HTMLButtonElement;
		const groupRect = sortButtonGroup.getBoundingClientRect();
		const buttonRect = button.getBoundingClientRect();
		hoverX = buttonRect.left - groupRect.left;
		buttonWidth = buttonRect.width - 5;
		showHover = true;
	}

	function sortColors(method: SortMethod) {
		if (appStore.state.sortMethod === method) return;
		appStore.state.sortMethod = method;
		const result = sortColorsByMethod(appStore.state.colors, method);
		appStore.resetThemeExportSession();
		appStore.state.colors = result.colors;

		if (result.hadNoChange) {
			toast('Colors have same order for this sorting method', {
				icon: '⚠️',
				duration: 3000
			});
		}
	}

	const sortOptions: Array<{ value: SortMethod; label: string }> = [
		{ value: 'hue', label: 'Hue' },
		{ value: 'saturation', label: 'Saturation' },
		{ value: 'lightness', label: 'Lightness' },
		{ value: 'luminance', label: 'Luminance' }
	];

	async function handleCopy(hex: string) {
		try {
			await navigator.clipboard.writeText(hex);
			toast.success('Copied to clipboard');
			tutorialStore.setColorCopied(true);
		} catch {
			toast.error('Could not copy to clipboard. Please try again.');
		}
	}

	async function returnToUpload() {
		await tick();
		appStore.resetThemeExportSession();
		appStore.state.image = null;
		appStore.state.imageLoaded = false;
		appStore.state.colors = [];
		appStore.state.activeSelectorId = DEFAULT_SELECTOR_ID;
		appStore.state.selectors.forEach((s) => {
			s.selection = undefined;
			s.selected = s.id === DEFAULT_SELECTOR_ID;
		});
		appStore.state.sortMethod = 'none';
	}
</script>

<section
	class={cn('flex w-full max-w-4xl flex-col gap-3 xl:max-w-5xl', appStore.state.imageLoaded ? 'min-h-50' : 'h-0 w-0')}
>
	{#if appStore.state.imageLoaded}
		<div class="flex min-h-9 flex-wrap items-center gap-2">
			<span class="text-sm font-medium text-zinc-300">Sort:</span>
			<div
				bind:this={sortButtonGroup}
				onmouseleave={() => (showHover = false)}
				role="group"
				aria-label="Sort options"
				class="border-brand/50 relative flex flex-wrap gap-0.5 rounded-md border bg-zinc-900 p-1 xl:gap-1"
			>
				<div
					class="absolute inset-y-1 left-1 rounded bg-zinc-700 transition-[transform,opacity] duration-300 ease-out"
					style="transform: translateX({hoverX}px); width: {buttonWidth}px; opacity: {showHover ? 1 : 0};"
				></div>
				{#each sortOptions as option (option.value)}
					<button
						type="button"
						onclick={() => sortColors(option.value)}
						onmouseenter={handleButtonHover}
						class={cn(
							'relative z-10 rounded px-2 py-1 text-xs font-medium transition-colors duration-300 xl:px-3 xl:py-1.5',
							appStore.state.sortMethod === option.value
								? 'bg-brand text-zinc-900'
								: 'text-zinc-300 hover:text-zinc-300'
						)}
					>
						{option.label}
					</button>
				{/each}
			</div>
		</div>
	{/if}

	<div class="grid min-h-24 grid-cols-10 items-center gap-4 transition-[opacity,transform] duration-300">
		{#each appStore.state.colors as color, i (`${color.hex}-${i}`)}
			<div
				id={i === 0 ? 'tutorial-color-swatch' : undefined}
				role="button"
				tabindex="0"
				onkeyup={(e) => (e.key === 'Enter' || e.key === ' ') && handleCopy(color.hex)}
				onclick={() => handleCopy(color.hex)}
				in:scale={{ delay: i * 60, duration: 200, start: 0.6 }}
				class="flex h-8 cursor-pointer items-center justify-center rounded-md p-1.5 shadow-md xl:h-9 xl:p-2"
				style="background-color: {color.hex}"
			>
				<span class="rounded bg-black/50 px-1.5 py-1 font-mono text-[11px] xl:px-2 xl:text-xs">{color.hex}</span>
			</div>
		{/each}
	</div>

	{#if appStore.state.imageLoaded}
		<div class="flex flex-row justify-between">
			<button
				class="border-brand/50 hover:shadow-brand-lg w-30 cursor-pointer rounded-md border bg-zinc-900 py-2 text-sm font-medium text-zinc-300 transition-[background-color,border-color,box-shadow,color] duration-300 xl:w-36"
				onclick={returnToUpload}>Back</button
			>

			<div class="flex items-center gap-4">
				<button
					id="save-palette"
					class="border-brand/50 hover:shadow-brand-lg w-30 cursor-pointer rounded-md border bg-zinc-900 py-2 text-sm font-medium text-zinc-300 transition-[background-color,border-color,box-shadow,color] duration-300 xl:w-36"
					onclick={appStore.savePalette}
				>
					Save Palette

					<span>🎨</span>
				</button>
			</div>
		</div>
	{/if}
</section>
