<script lang="ts">
	import { appStore } from '$lib/stores/app.svelte';
	import { popoverStore } from '$lib/stores/popovers.svelte';
	import { cn } from '$lib/utils';
	import { AVAILABLE_RATIOS } from '$lib/types/wallhaven';

	let localSettings = $state(JSON.parse(JSON.stringify(appStore.state.wallhavenSettings)));

	function applySettings() {
		appStore.state.wallhavenSettings = { ...localSettings };
		popoverStore.close();
	}

	function resetToDefaults() {
		localSettings = {
			categories: '111',
			purity: '100',
			sorting: 'relevance',
			order: 'desc',
			topRange: '1M',
			resolutions: [],
			ratios: [],
			colors: [],
			apikey: ''
		};
	}

	function toggleCategory(index: number) {
		const chars = localSettings.categories.split('');
		chars[index] = chars[index] === '1' ? '0' : '1';
		localSettings.categories = chars.join('');
	}

	function togglePurity(index: number) {
		const chars = localSettings.purity.split('');
		chars[index] = chars[index] === '1' ? '0' : '1';
		localSettings.purity = chars.join('');
	}
</script>

<div
	class={cn(
		'palette-dropdown-base animate-scale-in border-brand/50 shadow-brand/20 z-50 max-h-[80vh] w-80 overflow-hidden rounded-xl border bg-zinc-900 shadow-2xl',
		popoverStore.state.direction === 'right' ? 'left-full ml-2' : 'right-full mr-2'
	)}
	style={popoverStore.state.direction === 'right' ? 'left: calc(100% + 0.5rem);' : 'right: calc(100% + 0.5rem);'}
>
	<!-- Header -->
	<header class="border-b border-zinc-700 bg-zinc-800/50 px-6 py-5">
		<h2 class="text-brand text-xl font-semibold">Wallhaven Settings</h2>
		<p class="mt-1 text-xs text-zinc-400">Configure search parameters for Wallhaven API</p>
	</header>

	<!-- Content -->
	<main class="custom-scrollbar max-h-96 overflow-y-auto p-4">
		<!-- Search Options Section -->
		<section class="mb-6">
			<h3 class="text-brand mb-3 text-sm font-semibold tracking-wide uppercase">Search Options</h3>

			<!-- Categories -->
			<div class="mb-5">
				<span class="mb-2 block text-xs font-medium text-zinc-300">Categories</span>
				<div class="flex gap-4">
					<label for="cat-general" class="flex cursor-pointer items-center gap-2">
						<input
							id="cat-general"
							type="checkbox"
							checked={localSettings.categories[0] === '1'}
							onchange={() => toggleCategory(0)}
							class="text-brand focus:ring-brand h-3.5 w-3.5 rounded border-zinc-600 bg-zinc-800 focus:ring-2"
						/>
						<span class="text-xs text-zinc-300">General</span>
					</label>
					<label for="cat-anime" class="flex cursor-pointer items-center gap-2">
						<input
							id="cat-anime"
							type="checkbox"
							checked={localSettings.categories[1] === '1'}
							onchange={() => toggleCategory(1)}
							class="text-brand focus:ring-brand h-3.5 w-3.5 rounded border-zinc-600 bg-zinc-800 focus:ring-2"
						/>
						<span class="text-xs text-zinc-300">Anime</span>
					</label>
					<label for="cat-people" class="flex cursor-pointer items-center gap-2">
						<input
							id="cat-people"
							type="checkbox"
							checked={localSettings.categories[2] === '1'}
							onchange={() => toggleCategory(2)}
							class="text-brand focus:ring-brand h-3.5 w-3.5 rounded border-zinc-600 bg-zinc-800 focus:ring-2"
						/>
						<span class="text-xs text-zinc-300">People</span>
					</label>
				</div>
			</div>

			<!-- Purity -->
			<div class="mb-5">
				<span class="mb-2 block text-xs font-medium text-zinc-300">Content Purity</span>
				<div class="flex gap-3">
					<label for="pur-sfw" class="flex cursor-pointer items-center gap-2">
						<input
							id="pur-sfw"
							type="checkbox"
							checked={localSettings.purity[0] === '1'}
							onchange={() => togglePurity(0)}
							class="text-brand focus:ring-brand h-3.5 w-3.5 rounded border-zinc-600 bg-zinc-800 focus:ring-2"
						/>
						<span class="text-xs text-zinc-300">SFW</span>
					</label>
					<label for="pur-sketchy" class="flex cursor-pointer items-center gap-2">
						<input
							id="pur-sketchy"
							type="checkbox"
							checked={localSettings.purity[1] === '1'}
							onchange={() => togglePurity(1)}
							class="text-brand focus:ring-brand h-3.5 w-3.5 rounded border-zinc-600 bg-zinc-800 focus:ring-2"
						/>
						<span class="text-xs text-zinc-300">Sketchy</span>
					</label>
					<label for="pur-nsfw" class="flex cursor-pointer items-center gap-2">
						<input
							id="pur-nsfw"
							type="checkbox"
							checked={localSettings.purity[2] === '1'}
							onchange={() => togglePurity(2)}
							class="text-brand focus:ring-brand h-3.5 w-3.5 rounded border-zinc-600 bg-zinc-800 focus:ring-2"
						/>
						<span class="text-xs text-zinc-300">NSFW</span>
					</label>
				</div>
			</div>

			<!-- Sorting -->
			<div class="mb-5">
				<label for="sort-method" class="mb-2 block text-xs font-medium text-zinc-300">Sort By</label>
				<select
					id="sort-method"
					bind:value={localSettings.sorting}
					class="focus:border-brand/50 w-full rounded-lg border border-zinc-700 bg-zinc-900 px-2.5 py-1.5 text-sm text-zinc-300 transition-all duration-300 focus:outline-none"
				>
					<option value="relevance">Relevance</option>
					<option value="date_added">Date Added</option>
					<option value="random">Random</option>
					<option value="views">Views</option>
					<option value="favorites">Favorites</option>
					<option value="toplist">Toplist</option>
					<option value="hot">Hot</option>
				</select>
			</div>

			<!-- Order -->
			<div class="mb-5">
				<label for="sort-order" class="mb-2 block text-xs font-medium text-zinc-300">Sort Order</label>
				<select
					id="sort-order"
					bind:value={localSettings.order}
					class="focus:border-brand/50 w-full rounded-lg border border-zinc-700 bg-zinc-900 px-2.5 py-1.5 text-sm text-zinc-300 transition-all duration-300 focus:outline-none"
				>
					<option value="desc">Descending</option>
					<option value="asc">Ascending</option>
				</select>
			</div>

			<!-- Top Range (only for toplist sorting) -->
			{#if localSettings.sorting === 'toplist'}
				<div class="mb-5">
					<label for="time-range" class="mb-2 block text-xs font-medium text-zinc-300">Time Range</label>
					<select
						id="time-range"
						bind:value={localSettings.topRange}
						class="focus:border-brand/50 w-full rounded-lg border border-zinc-700 bg-zinc-900 px-2.5 py-1.5 text-sm text-zinc-300 transition-all duration-300 focus:outline-none"
					>
						<option value="1d">1 Day</option>
						<option value="3d">3 Days</option>
						<option value="1w">1 Week</option>
						<option value="1M">1 Month</option>
						<option value="3M">3 Months</option>
						<option value="6M">6 Months</option>
						<option value="1y">1 Year</option>
					</select>
				</div>
			{/if}
		</section>

		<!-- Advanced Section -->
		<section class="mb-4">
			<h3 class="text-brand mb-3 text-sm font-semibold tracking-wide uppercase">Advanced</h3>

			<!-- API Key -->
			<div class="mb-5">
				<label for="api-key" class="mb-2 block text-xs font-medium text-zinc-300">
					API Key
					<span class="ml-1 font-normal text-zinc-500">(optional, for NSFW content)</span>
				</label>
				<input
					id="api-key"
					type="password"
					bind:value={localSettings.apikey}
					placeholder="Enter your Wallhaven API key"
					class="focus:border-brand/50 w-full rounded-lg border border-zinc-700 bg-zinc-900 px-2.5 py-1.5 text-sm text-zinc-300 placeholder-zinc-500 transition-all duration-300 focus:outline-none"
				/>
				<p class="mt-1 text-xs text-zinc-500">Get your API key from Wallhaven account settings</p>
			</div>

			<!-- Aspect Ratios -->
			<div class="mb-5">
				<label for="ratio-filter" class="mb-2 block text-xs font-medium text-zinc-300">
					Aspect Ratios
					<span class="ml-1 font-normal text-zinc-500">(optional)</span>
				</label>
				<select
					id="ratio-filter"
					multiple
					bind:value={localSettings.ratios}
					class="focus:border-brand/50 w-full rounded-lg border border-zinc-700 bg-zinc-900 px-2.5 py-1.5 text-sm text-zinc-300 transition-all duration-300 focus:outline-none"
					size="4"
				>
					{#each AVAILABLE_RATIOS as ratio (ratio)}
						<option value={ratio}>{ratio}</option>
					{/each}
				</select>
				<p class="mt-1 text-xs text-zinc-500">Hold Ctrl/Cmd to select multiple ratios</p>
			</div>
		</section>
	</main>

	<!-- Actions -->
	<footer class="flex justify-center gap-2 border-t border-zinc-700 bg-zinc-800/50 px-2 py-3">
		<button
			onclick={resetToDefaults}
			class="hover:border-brand/50 rounded-lg border border-zinc-600 px-2 py-3 text-xs font-medium text-zinc-300 transition-all duration-300 hover:bg-zinc-800/50"
		>
			Reset to Defaults
		</button>
		<button
			onclick={() => popoverStore.close()}
			class="hover:border-brand/50 rounded-lg border border-zinc-600 px-2 py-3 text-xs font-medium text-zinc-300 transition-all duration-300 hover:bg-zinc-800/50"
		>
			Cancel
		</button>
		<button
			onclick={applySettings}
			class="bg-brand shadow-brand/20 hover:shadow-brand/40 rounded-lg px-2 py-3 text-xs font-semibold text-zinc-900 transition-all duration-300 hover:scale-105 disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:scale-100"
		>
			Apply Settings
		</button>
	</footer>
</div>
