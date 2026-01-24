<script lang="ts">
	import { appStore } from '$lib/stores/app.svelte';
	import SearchModal from './SearchModal.svelte';

	let showSearchModal = $state(false);

	function onSearchSubmit(e?: Event) {
		e?.preventDefault();
		showSearchModal = true;
	}
	function clearSearch() {
		appStore.state.searchQuery = '';
	}

	function onInlineInput() {
		const q = String(appStore.state.searchQuery).trimStart();
		if (q.length > 0) {
			showSearchModal = true;
		}
	}

	function closeSearchModal() {
		showSearchModal = false;
	}

	function onKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape' && showSearchModal) {
			closeSearchModal();
		}
	}
</script>

<svelte:window onkeydown={onKeydown} />

<form class="group relative mx-auto flex max-w-md flex-1" onsubmit={(e) => onSearchSubmit(e)} aria-label="Search">
	<label for="search" class="sr-only">Search palettes</label>
	<input
		id="search"
		type="search"
		bind:value={appStore.state.searchQuery}
		oninput={onInlineInput}
		placeholder="Search Wallpapers"
		class="focus:border-brand/50 w-full rounded-lg border border-zinc-700 bg-zinc-900 py-2 pr-10 pl-10 text-sm text-zinc-300 placeholder-zinc-400 transition-[border-color,box-shadow,background-color] duration-300 focus:outline-none"
	/>

	<span class="group-focus-within:text-brand absolute top-1/2 left-3 -translate-y-1/2 text-zinc-400 transition-colors">
		<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				stroke-width="2"
				d="M21 21l-4.35-4.35M10.5 18a7.5 7.5 0 100-15 7.5 7.5 0 000 15z"
			/>
		</svg>
	</span>

	<div class="absolute top-1/2 right-2 flex -translate-y-1/2 items-center gap-1">
		{#if appStore.state.searchQuery}
			<button
				type="button"
				onclick={clearSearch}
				class="hover:text-brand rounded-lg p-1 text-zinc-400 transition-[background-color,color] duration-300 hover:bg-zinc-800/50"
			>
				<span class="sr-only">Clear search</span>
				<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
				</svg>
			</button>
		{/if}
		<button
			type="submit"
			class="bg-brand hover:shadow-brand/20 rounded-lg px-3 py-1 text-sm font-medium text-zinc-900 transition-[box-shadow,background-color] duration-300"
			>Search</button
		>
	</div>
</form>

<SearchModal bind:isOpen={showSearchModal} />
