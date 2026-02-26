import { browser } from '$app/environment';
import type { EditorThemeType } from './api/theme';
import type { ApplyPaletteSettings } from './types/applyPaletteSettings';
import type { SavedThemeItem } from './types/theme';
import type { WallhavenSettings } from './types/wallhaven';

const THEME_EXPORT_STORAGE_KEY = 'themeExportPreferences';
const SAVED_THEMES_STORAGE_KEY = 'savedThemes';
const WALLHAVEN_SETTINGS_KEY = 'wallhavenSettings';
const APPLY_PALETTE_SETTINGS_KEY = 'applyPaletteSettings';

const DEFAULT_WALLHAVEN_SETTINGS: WallhavenSettings = {
	categories: '111',
	purity: '100',
	sorting: 'relevance',
	order: 'desc',
	topRange: '1M',
	ratios: [],
	apikey: ''
};

const DEFAULT_APPLY_PALETTE_SETTINGS: ApplyPaletteSettings = {
	luminosity: 1,
	nearest: 30,
	power: 4,
	maxDistance: 0
};

export function loadThemeExportPreferences(): EditorThemeType {
	if (!browser) return 'vscode';
	const stored = localStorage.getItem(THEME_EXPORT_STORAGE_KEY);
	if (stored) {
		const parsed = JSON.parse(stored);
		return parsed.editorType || 'vscode';
	}
	return 'vscode';
}

export function loadThemeExportSaveOnCopy(): boolean {
	if (!browser) return true;
	const stored = localStorage.getItem(THEME_EXPORT_STORAGE_KEY);
	if (stored) {
		const parsed = JSON.parse(stored);
		return parsed.saveOnCopy ?? true;
	}
	return true;
}

export function saveThemeExportPreferences(editorType: EditorThemeType, saveOnCopy: boolean) {
	if (!browser) return;
	localStorage.setItem(THEME_EXPORT_STORAGE_KEY, JSON.stringify({ editorType, saveOnCopy }));
}

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

export function saveSavedThemes(themes: SavedThemeItem[]) {
	if (!browser) return;
	localStorage.setItem(SAVED_THEMES_STORAGE_KEY, JSON.stringify(themes));
}

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

export function loadApplyPaletteSettings(): ApplyPaletteSettings {
	if (!browser) return DEFAULT_APPLY_PALETTE_SETTINGS;
	const stored = localStorage.getItem(APPLY_PALETTE_SETTINGS_KEY);
	if (stored) {
		const parsed = JSON.parse(stored);
		return {
			luminosity: parsed.luminosity ?? DEFAULT_APPLY_PALETTE_SETTINGS.luminosity,
			nearest: parsed.nearest ?? DEFAULT_APPLY_PALETTE_SETTINGS.nearest,
			power: parsed.power ?? DEFAULT_APPLY_PALETTE_SETTINGS.power,
			maxDistance: parsed.maxDistance ?? DEFAULT_APPLY_PALETTE_SETTINGS.maxDistance
		};
	}
	return DEFAULT_APPLY_PALETTE_SETTINGS;
}

export function saveApplyPaletteSettings(settings: ApplyPaletteSettings) {
	if (!browser) return;
	localStorage.setItem(APPLY_PALETTE_SETTINGS_KEY, JSON.stringify(settings));
}
