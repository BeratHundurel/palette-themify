import { browser } from '$app/environment';
import type { WallhavenSettings } from '$lib/types/wallhaven';

const WALLHAVEN_SETTINGS_KEY = 'wallhavenSettings';
export const DEFAULT_WALLHAVEN_SETTINGS: WallhavenSettings = {
	categories: '111',
	purity: '100',
	sorting: 'relevance',
	order: 'desc',
	topRange: '1M',
	ratios: [],
	apikey: ''
};

export function loadWallhavenSettings(): WallhavenSettings {
	if (!browser) return DEFAULT_WALLHAVEN_SETTINGS;
	const stored = localStorage.getItem(WALLHAVEN_SETTINGS_KEY);
	if (stored) {
		const parsed = JSON.parse(stored);
		return {
			categories: parsed.categories ?? DEFAULT_WALLHAVEN_SETTINGS.categories,
			purity: parsed.purity ?? DEFAULT_WALLHAVEN_SETTINGS.purity,
			sorting: parsed.sorting ?? DEFAULT_WALLHAVEN_SETTINGS.sorting,
			order: parsed.order ?? DEFAULT_WALLHAVEN_SETTINGS.order,
			topRange: parsed.topRange ?? DEFAULT_WALLHAVEN_SETTINGS.topRange,
			ratios: parsed.ratios ?? DEFAULT_WALLHAVEN_SETTINGS.ratios,
			apikey: parsed.apikey ?? DEFAULT_WALLHAVEN_SETTINGS.apikey
		};
	}
	return DEFAULT_WALLHAVEN_SETTINGS;
}

export function saveWallhavenSettings(settings: WallhavenSettings) {
	if (!browser) return;
	localStorage.setItem(WALLHAVEN_SETTINGS_KEY, JSON.stringify(settings));
}

export function clearWallhavenSettings() {
	if (!browser) return;
	localStorage.removeItem(WALLHAVEN_SETTINGS_KEY);
}
