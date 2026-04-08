import type { WallhavenResult, WallhavenSettings } from '$lib/types/wallhaven';

export type PageResult = {
	page: number;
	items: WallhavenResult[];
};

export type SearchState = {
	pages: PageResult[];
	hasMore: boolean;
	page: number;
	lastQuery: string;
	isSearching: boolean;
	loadingMore: boolean;
	latestSearchRequestId: number;
};

export const SEARCH_DEBOUNCE_MS = 400;
export const SEARCH_SCROLL_THRESHOLD_PX = 300;

export function normalizeSearchQuery(query: string): string {
	return query.trim();
}

export function createInitialSearchState(): SearchState {
	return {
		pages: [],
		hasMore: true,
		page: 1,
		lastQuery: '',
		isSearching: false,
		loadingMore: false,
		latestSearchRequestId: 0
	};
}

export function handleSearchQuery(query: string, state: SearchState): SearchState {
	const normalizedQuery = normalizeSearchQuery(query);

	if (!normalizedQuery) {
		return {
			...state,
			latestSearchRequestId: state.latestSearchRequestId + 1,
			pages: [],
			hasMore: true,
			page: 1,
			lastQuery: '',
			isSearching: false,
			loadingMore: false
		};
	}

	return {
		...state,
		lastQuery: normalizedQuery,
		isSearching: true
	};
}

export function getSearchLastPage(meta: { last_page?: number; per_page?: number; total?: number }): number | null {
	if (typeof meta.last_page === 'number') return meta.last_page;
	if (typeof meta.per_page === 'number' && typeof meta.total === 'number') {
		return Math.ceil(meta.total / meta.per_page);
	}
	return null;
}

export function dedupeResults(existingPages: PageResult[], nextItems: WallhavenResult[]): WallhavenResult[] {
	const seenIds = new Set(existingPages.flatMap((resultPage) => resultPage.items).map((result) => result.id));
	return nextItems.filter((item) => !seenIds.has(item.id));
}

export function getTotalResults(pages: PageResult[]): number {
	return pages.reduce((sum, resultPage) => sum + resultPage.items.length, 0);
}

export function hasAnyResults(pages: PageResult[]): boolean {
	return pages.some((resultPage) => resultPage.items.length > 0);
}

export function getWallhavenSettingsSignature(settings: WallhavenSettings): string {
	return [
		settings.categories,
		settings.purity,
		settings.sorting,
		settings.order,
		settings.topRange,
		settings.apikey,
		settings.ratios.join(',')
	].join('|');
}
