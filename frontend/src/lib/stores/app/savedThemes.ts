import { browser } from '$app/environment';
import type { SavedThemeItem } from '$lib/types/theme';

const SAVED_THEMES_STORAGE_KEY = 'savedThemes';

export function loadSavedThemes(): SavedThemeItem[] {
	if (!browser) return [];
	try {
		const stored = localStorage.getItem(SAVED_THEMES_STORAGE_KEY);
		if (!stored) return [];
		const parsed = JSON.parse(stored) as SavedThemeItem[];
		if (!Array.isArray(parsed)) return [];
		return parsed;
	} catch {
		return [];
	}
}

export function clearSavedThemes() {
	if (!browser) return;
	localStorage.removeItem(SAVED_THEMES_STORAGE_KEY);
}

export function saveSavedThemes(themes: SavedThemeItem[]) {
	if (!browser) return;
	localStorage.setItem(SAVED_THEMES_STORAGE_KEY, JSON.stringify(themes));
}
