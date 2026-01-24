<script lang="ts">
	import { cn } from '$lib/utils';
	import { appStore } from '$lib/stores/app.svelte';
	import { popoverStore } from '$lib/stores/popovers.svelte';

	let luminosity = $derived(appStore.state.luminosity);
	let nearest = $derived(appStore.state.nearest);
	let power = $derived(appStore.state.power);
	let maxDistance = $derived(appStore.state.maxDistance);

	let showTooltip = $state('');

	const settings = $derived([
		{
			id: 'luminosity',
			label: 'Luminosity',
			value: luminosity,
			min: 0.1,
			max: 3.0,
			step: 0.1,
			tooltip: 'Brightness adjustment factor. 1.0 = no change, >1.0 = brighter, <1.0 = darker',
			action: (val: number) => (appStore.state.luminosity = val)
		},
		{
			id: 'nearest',
			label: 'Nearest Count',
			value: nearest,
			min: 1,
			max: 20,
			step: 1,
			tooltip: 'Number of palette colors to use for interpolation. Higher = smoother blending',
			action: (val: number) => (appStore.state.nearest = val)
		},
		{
			id: 'power',
			label: 'Power',
			value: power,
			min: 0.5,
			max: 10.0,
			step: 0.5,
			tooltip: 'Distance weighting power. Lower = soft blending, higher = sharp transitions',
			action: (val: number) => (appStore.state.power = val)
		},
		{
			id: 'maxDistance',
			label: 'Max Distance',
			value: maxDistance,
			min: 0,
			max: 200,
			step: 5,
			tooltip: 'Maximum distance threshold. Only pixels within this range get recolored (0 = all pixels)',
			action: (val: number) => (appStore.state.maxDistance = val)
		}
	]);

	function handleSliderChange(value: number, action: (val: number) => void) {
		action(value);
	}
</script>

<div
	class={cn(
		'palette-dropdown-base flex min-w-80 flex-col gap-4',
		popoverStore.state.direction === 'right' ? 'left-full ml-2' : 'right-full mr-2'
	)}
	style={popoverStore.state.direction === 'right' ? 'left: calc(100% + 0.5rem);' : 'right: calc(100% + 0.5rem);'}
>
	<h3 class="text-brand mb-1 text-xs font-medium">How your saved palettes are applied to current image</h3>

	{#each settings as setting (setting.id)}
		<div class="flex flex-col gap-2">
			<div class="flex items-center justify-between">
				<label for={setting.id} class="text-xs font-medium text-zinc-300">
					{setting.label}
				</label>
				<span
					class="relative cursor-help"
					onmouseenter={() => (showTooltip = setting.id)}
					onmouseleave={() => (showTooltip = '')}
					role="tooltip"
				>
					<svg xmlns="http://www.w3.org/2000/svg" height="16px" viewBox="0 -960 960 960" width="16px" fill="#fff">
						<path
							d="M440-280h80v-240h-80v240Zm40-320q17 0 28.5-11.5T520-640q0-17-11.5-28.5T480-680q-17 0-28.5 11.5T440-640q0 17 11.5 28.5T480-600Zm0 520q-83 0-156-31.5T197-197q-54-54-85.5-127T80-480q0-83 31.5-156T197-763q54-54 127-85.5T480-880q83 0 156 31.5T763-763q54 54 85.5 127T880-480q0 83-31.5 156T763-197q-54 54-127 85.5T480-80Zm0-80q134 0 227-93t93-227q0-134-93-227t-227-93q-134 0-227 93t-93 227q0 134 93 227t227 93Zm0-320Z"
						/>
					</svg>
					{#if showTooltip === setting.id}
						<div
							class="absolute bottom-full left-1/2 z-10 mb-2 w-64 -translate-x-1/2 rounded border border-zinc-700 bg-zinc-800 px-3 py-2 text-xs whitespace-normal text-zinc-300 shadow-lg"
						>
							{setting.tooltip}
						</div>
					{/if}
				</span>
			</div>

			<div class="flex items-center gap-3">
				<input
					id={setting.id}
					type="range"
					min={setting.min}
					max={setting.max}
					step={setting.step}
					value={setting.value}
					oninput={(e) => handleSliderChange(parseFloat(e.currentTarget.value), setting.action)}
					class="slider flex-1 cursor-pointer appearance-none bg-transparent"
				/>
				<input
					type="number"
					min={setting.min}
					max={setting.max}
					step={setting.step}
					value={setting.value}
					onchange={(e) => handleSliderChange(parseFloat(e.currentTarget.value), setting.action)}
					class="focus:border-brand/50 w-16 rounded border border-zinc-700 bg-zinc-800 px-2 py-1 text-center text-xs text-zinc-300 focus:outline-none"
				/>
			</div>
		</div>
	{/each}

	<div class="mt-2 space-y-3 border-t border-zinc-700 pt-3">
		<p class="text-xs text-zinc-400">Adjust these settings to fine-tune how the palette is applied to your image.</p>
	</div>
</div>

<style>
	.slider {
		height: 6px;
		border-radius: 3px;
		background: linear-gradient(to right, #3f3f46 0%, #d4a574 100%);
		outline: none;
	}

	.slider::-webkit-slider-thumb {
		appearance: none;
		height: 16px;
		width: 16px;
		border-radius: 50%;
		background: #d4a574;
		cursor: pointer;
		border: 2px solid #27272a;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
	}

	.slider::-moz-range-thumb {
		height: 16px;
		width: 16px;
		border-radius: 50%;
		background: #d4a574;
		cursor: pointer;
		border: 2px solid #27272a;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
	}

	.slider::-webkit-slider-thumb:hover {
		background: #e6c29a;
		transform: scale(1.1);
	}
</style>
