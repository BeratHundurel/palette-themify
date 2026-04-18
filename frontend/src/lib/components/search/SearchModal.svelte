<script lang="ts">
	import { searchWallhaven } from '$lib/api/wallhaven';
	import { appStore } from '$lib/stores/app/store.svelte';
	import toast from 'svelte-french-toast';
	import {
		createInitialSearchState,
		dedupeResults,
		getSearchLastPage,
		getWallhavenSettingsSignature,
		getTotalResults,
		handleSearchQuery,
		hasAnyResults,
		normalizeSearchQuery,
		SEARCH_DEBOUNCE_MS,
		SEARCH_SCROLL_THRESHOLD_PX,
		type SearchState,
		type PageResult
	} from './search';
	import { cn } from '$lib/utils';
	import { isDesktopApp } from '$lib/platform';

	let { isOpen = $bindable() } = $props();

	// modal ui state
	let lastQuery = $state('');
	let lastSettingsSignature = $state('');
	let lastScrollTop = $state(0);
	let scrollEl = $state<HTMLDivElement | null>(null);
	let inputEl = $state<HTMLInputElement | null>(null);
	let searchTimer: number | null = null;

	// search state
	const initialSearchState = createInitialSearchState();
	let pages = $state<PageResult[]>(initialSearchState.pages);
	let page = $state(initialSearchState.page);
	let hasMore = $state(initialSearchState.hasMore);
	let isSearching = $state(initialSearchState.isSearching);
	let loadingMore = $state(initialSearchState.loadingMore);
	let latestSearchRequestId = initialSearchState.latestSearchRequestId;

	const normalizedQuery = $derived(normalizeSearchQuery(appStore.state.searchQuery));
	const currentSettingsSignature = $derived(getWallhavenSettingsSignature(appStore.state.wallhavenSettings));
	const showEmptyState = $derived(pages.length === 0 || !hasAnyResults(pages));

	function snapshotSearchState(): SearchState {
		return {
			pages,
			hasMore,
			page,
			lastQuery,
			isSearching,
			loadingMore,
			latestSearchRequestId
		};
	}

	function restoreSearchState(nextState: SearchState) {
		pages = nextState.pages;
		hasMore = nextState.hasMore;
		page = nextState.page;
		lastQuery = nextState.lastQuery;
		isSearching = nextState.isSearching;
		loadingMore = nextState.loadingMore;
		latestSearchRequestId = nextState.latestSearchRequestId;
	}

	function cancelScheduledSearch() {
		if (searchTimer === null) return;
		clearTimeout(searchTimer);
		searchTimer = null;
	}

	function closeModal() {
		if (scrollEl) {
			lastScrollTop = scrollEl.scrollTop;
		}
		cancelScheduledSearch();
		isOpen = false;
	}

	function loadWallpaperForPalette(url: string | undefined) {
		if (!url) return;
		const toastId = toast.loading('Extracting palette...');
		appStore.loadWallhavenImage(url, toastId);
		closeModal();
	}

	$effect(() => {
		if (isOpen && inputEl) {
			inputEl.focus();
		}

		if (isOpen && scrollEl) {
			scrollEl.scrollTop = lastScrollTop;
		}
	});

	async function performSearch(query: string, requestedPage = 1, append = false) {
		const normalized = normalizeSearchQuery(query);
		if (!normalized) {
			restoreSearchState(handleSearchQuery(normalized, snapshotSearchState()));
			return;
		}

		const requestId = ++latestSearchRequestId;
		if (requestedPage <= 1) {
			isSearching = true;
		} else {
			loadingMore = true;
		}

		try {
			lastSettingsSignature = currentSettingsSignature;
			const res = await searchWallhaven(appStore.state.wallhavenSettings, normalized, requestedPage);
			if (requestId !== latestSearchRequestId) return;

			const data = res.data || [];
			if (append) {
				const newItems = dedupeResults(pages, data);
				if (newItems.length > 0) {
					pages = [...pages, { page: requestedPage, items: newItems }];
				} else {
					hasMore = false;
				}
			} else {
				pages = [{ page: 1, items: data }];
			}

			const lastPage = getSearchLastPage(res.meta ?? {});
			if (hasMore) {
				if (lastPage !== null) {
					hasMore = requestedPage < lastPage;
				} else {
					hasMore = data.length > 0;
				}
			}
			page = requestedPage;
			lastQuery = normalized;
		} catch {
			if (requestId !== latestSearchRequestId) return;
			if (!append) pages = [];
			hasMore = false;
		} finally {
			if (requestId === latestSearchRequestId) {
				isSearching = false;
				loadingMore = false;
			}
		}
	}

	function scheduleSearch(query: string) {
		if (!isOpen) return;
		cancelScheduledSearch();

		restoreSearchState(handleSearchQuery(query, snapshotSearchState()));
		const normalized = normalizeSearchQuery(query);
		if (!normalized) {
			return;
		}

		searchTimer = window.setTimeout(() => {
			page = 1;
			hasMore = true;
			void performSearch(normalized, 1, false);
			searchTimer = null;
		}, SEARCH_DEBOUNCE_MS);
	}

	async function loadMore() {
		if (isSearching || loadingMore || !hasMore || searchTimer !== null) return;
		const next = page + 1;
		await performSearch(lastQuery || appStore.state.searchQuery, next, true);
	}

	function handleScroll(e: Event) {
		const target = e.target as HTMLElement;
		if (!target) return;
		lastScrollTop = target.scrollTop;
		if (target.scrollHeight - target.scrollTop - target.clientHeight < SEARCH_SCROLL_THRESHOLD_PX) {
			void loadMore();
		}
	}

	function scheduleSearchIfNeeded() {
		const settingsChangedSinceLastSearch = lastSettingsSignature !== currentSettingsSignature;
		if (lastQuery !== normalizedQuery || (settingsChangedSinceLastSearch && normalizedQuery.length > 0)) {
			scheduleSearch(appStore.state.searchQuery);
		}
	}

	function handleEscapeKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			closeModal();
		}
	}
</script>

{#if isOpen}
	<div
		class={cn('fixed inset-0 flex items-start justify-center bg-black/80 px-6 pb-6', isDesktopApp ? 'pt-16' : 'pt-6')}
		role="button"
		tabindex="0"
		onclick={closeModal}
		onkeydown={handleEscapeKeydown}
	>
		<div
			class="animate-scale-in border-brand/50 shadow-brand/20 relative w-full max-w-4xl rounded-xl border bg-zinc-900 shadow-2xl xl:max-w-5xl"
			role="dialog"
			tabindex="-1"
			onclick={(e) => e.stopPropagation()}
			onkeydown={handleEscapeKeydown}
		>
			<!-- Close Button Overlay -->
			<button
				onclick={closeModal}
				class="hover:text-brand absolute top-4 right-4 z-10 rounded-lg p-2 text-zinc-400 transition-[background-color,color] duration-300 hover:bg-zinc-800/50"
				aria-label="Close search"
			>
				<svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
				</svg>
			</button>

			<!-- Search Input Section -->
			<div class="px-6 py-6">
				<div class="relative mx-auto max-w-2xl">
					<input
						id="modal-search"
						type="text"
						bind:this={inputEl}
						bind:value={appStore.state.searchQuery}
						onchange={scheduleSearchIfNeeded}
						onfocus={scheduleSearchIfNeeded}
						oninput={() => scheduleSearch(appStore.state.searchQuery)}
						placeholder="Search wallpapers..."
						class="text-md w-full rounded border-none bg-zinc-800/75 p-4 font-light text-zinc-100 placeholder-zinc-500 transition-[background-color,box-shadow,color] duration-300 outline-none"
					/>
					<div
						class="focus-within:bg-brand absolute bottom-0 left-0 h-px w-full bg-black transition-colors duration-300"
					></div>
				</div>
			</div>

			<div
				class={cn('custom-scrollbar overflow-auto px-6 pb-6', isDesktopApp ? 'max-h-[77svh]' : 'max-h-[80svh]')}
				bind:this={scrollEl}
				onscroll={handleScroll}
			>
				{#if isSearching}
					<div class="flex flex-col items-center py-12">
						<div
							class="text-brand border-t-brand mb-4 h-8 w-8 animate-spin rounded-full border-2 border-zinc-600"
						></div>
						<p class="text-sm font-medium text-zinc-300">Searching wallpapers...</p>
						<p class="text-xs text-zinc-500">Finding the perfect images for you</p>
					</div>
				{:else if showEmptyState}
					{#if appStore.state.searchQuery}
						<div class="flex flex-col items-center py-12">
							<svg class="mb-4 h-12 w-12 text-zinc-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
								/>
							</svg>
							<p class="mb-2 text-sm font-medium text-zinc-300">No results found</p>
							<p class="text-xs text-zinc-500">Try adjusting your search terms</p>
						</div>
					{:else}
						<div class="flex flex-col items-center py-12">
							<svg class="mb-4 h-12 w-12 text-zinc-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
								/>
							</svg>
							<p class="mb-2 text-sm font-medium text-zinc-300">Start searching</p>
							<p class="text-xs text-zinc-500">Type keywords to find wallpapers</p>
						</div>
					{/if}
				{:else}
					{@const totalResults = getTotalResults(pages)}
					<div class="mb-6">
						<p class="text-sm font-medium text-zinc-300">
							{totalResults}
							{totalResults === 1 ? 'result' : 'results'} for "{appStore.state.searchQuery}"
						</p>
					</div>

					{#each pages as pageResult (pageResult.page)}
						{#if pageResult.page > 1}
							<div class="page-separator">
								<span class="bg-zinc-900 px-3 text-xs font-medium text-zinc-500">Page {pageResult.page}</span>
							</div>
						{/if}

						<div class="masonry pe-3">
							{#each pageResult.items as result, idx (result.id + '-' + idx)}
								<div
									class="masonry-item group hover:border-brand/50 hover:shadow-brand/10 relative overflow-hidden rounded-lg border border-zinc-700/50 transition-[border-color,box-shadow] duration-300"
								>
									{#if result.path && result.thumbs && result.thumbs.original}
										<button
											class="relative block h-full w-full overflow-hidden"
											onclick={() => loadWallpaperForPalette(result.path)}
											onkeydown={(e) => {
												if (e.key === 'Enter' || e.key === ' ') {
													loadWallpaperForPalette(result.path);
												}
											}}
										>
											<img
												src={result.thumbs.original}
												alt="wallpaper thumb"
												class="h-full w-full object-cover transition-[transform,opacity] duration-300 group-hover:scale-105 group-hover:opacity-90"
											/>
											<div
												class="absolute inset-0 bg-linear-to-t from-black/60 via-transparent to-transparent opacity-0 transition-opacity duration-300 group-hover:opacity-100"
											>
												<div class="absolute right-2 bottom-2 left-2">
													<p class="text-xs font-medium text-white">Click to extract palette</p>
												</div>
											</div>
										</button>
									{:else}
										<div
											class="flex aspect-video w-full items-center justify-center bg-linear-to-br from-zinc-700 to-zinc-900"
										>
											<svg class="h-8 w-8 text-zinc-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path
													stroke-linecap="round"
													stroke-linejoin="round"
													stroke-width="2"
													d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
												/>
											</svg>
										</div>
									{/if}
								</div>
							{/each}
						</div>
					{/each}

					{#if loadingMore}
						<div class="flex items-center justify-center py-6">
							<div class="text-brand border-t-brand h-6 w-6 animate-spin rounded-full border-2 border-zinc-600"></div>
						</div>
					{:else if !hasMore && totalResults > 0}
						<div class="mt-6 text-center">
							<p class="text-xs font-medium text-zinc-500">You've reached the end of the results</p>
						</div>
					{/if}
				{/if}
			</div>
		</div>
	</div>
{/if}

<style>
	.masonry {
		column-gap: 16px;
		columns: 1;
		margin: 16px auto 0 auto;
	}

	@media (min-width: 768px) {
		.masonry {
			columns: 2;
		}
	}

	@media (min-width: 1280px) {
		.masonry {
			columns: 3;
		}
	}

	.masonry-item {
		display: inline-block;
		width: 100%;
		height: 100%;
		margin: 0 0 8px 0;
		break-inside: avoid;
		-webkit-column-break-inside: avoid;
		page-break-inside: avoid;
		overflow: hidden;
	}

	.masonry-item img {
		width: 100%;
		height: auto;
		display: block;
		object-fit: cover;
	}

	/* Custom scrollbar styling */
	.custom-scrollbar::-webkit-scrollbar {
		width: 8px;
	}

	.custom-scrollbar::-webkit-scrollbar-track {
		background: rgba(0, 0, 0, 0.1);
		border-radius: 4px;
	}

	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: rgba(168, 85, 247, 0.3);
		border-radius: 4px;
		transition: background-color 0.3s;
	}

	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: rgba(168, 85, 247, 0.5);
	}

	.custom-scrollbar::-webkit-scrollbar-thumb:active {
		background: rgba(168, 85, 247, 0.7);
	}

	.page-separator {
		display: flex;
		align-items: center;
		justify-content: center;
		margin: 1.5rem 0;
		position: relative;
		break-inside: avoid;
		-webkit-column-break-inside: avoid;
		page-break-inside: avoid;
	}

	.page-separator::before,
	.page-separator::after {
		content: '';
		flex: 1;
		height: 1px;
		background: linear-gradient(to right, transparent, rgba(168, 85, 247, 0.1), transparent);
	}

	.page-separator span {
		font-size: 0.75rem;
		color: #71717a;
		background: #18181b;
		padding: 0 0.75rem;
		position: relative;
		z-index: 1;
	}
</style>
