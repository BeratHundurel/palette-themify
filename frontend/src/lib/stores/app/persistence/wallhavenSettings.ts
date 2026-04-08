import { browser } from '$app/environment';
import { DEFAULT_WALLHAVEN_SETTINGS, type WallhavenSettings } from '$lib/types/wallhaven';

const WALLHAVEN_SETTINGS_KEY = 'wallhavenSettings';

function asString(value: unknown, fallback: string): string {
	return typeof value === 'string' ? value : fallback;
}

export function parseWallhavenSettings(value: unknown): WallhavenSettings {
	if (!value || typeof value !== 'object') return { ...DEFAULT_WALLHAVEN_SETTINGS };
	const parsed = value as Partial<WallhavenSettings>;
	return {
		categories: asString(parsed.categories, DEFAULT_WALLHAVEN_SETTINGS.categories),
		purity: asString(parsed.purity, DEFAULT_WALLHAVEN_SETTINGS.purity),
		sorting: asString(parsed.sorting, DEFAULT_WALLHAVEN_SETTINGS.sorting),
		order: asString(parsed.order, DEFAULT_WALLHAVEN_SETTINGS.order),
		topRange: asString(parsed.topRange, DEFAULT_WALLHAVEN_SETTINGS.topRange),
		ratios: Array.isArray(parsed.ratios) ? parsed.ratios.filter((ratio) => typeof ratio === 'string') : [],
		apikey: asString(parsed.apikey, DEFAULT_WALLHAVEN_SETTINGS.apikey ?? '')
	};
}

export function loadWallhavenSettings(): WallhavenSettings {
	if (!browser) return { ...DEFAULT_WALLHAVEN_SETTINGS };
	const stored = localStorage.getItem(WALLHAVEN_SETTINGS_KEY);
	if (stored) {
		try {
			return parseWallhavenSettings(JSON.parse(stored));
		} catch {
			return { ...DEFAULT_WALLHAVEN_SETTINGS };
		}
	}
	return { ...DEFAULT_WALLHAVEN_SETTINGS };
}

export function saveWallhavenSettings(settings: WallhavenSettings) {
	if (!browser) return;
	localStorage.setItem(WALLHAVEN_SETTINGS_KEY, JSON.stringify(settings));
}

export function clearWallhavenSettings() {
	if (!browser) return;
	localStorage.removeItem(WALLHAVEN_SETTINGS_KEY);
}
