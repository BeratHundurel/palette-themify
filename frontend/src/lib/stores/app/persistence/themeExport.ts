import { browser } from '$app/environment';
import type { EditorThemeType, ThemeAppearance } from '$lib/api/theme';
import { DEFAULT_THEME_EXPORT_PREFERENCES, type ThemeExportPreferences } from '$lib/types/theme';

const THEME_EXPORT_STORAGE_KEY = 'themeExportPreferences';

function asNonNegativeFiniteNumber(value: unknown, fallback: number): number {
	if (typeof value !== 'number' || !Number.isFinite(value) || value < 0) return fallback;
	return Math.min(3, value);
}

function asEditorThemeType(value: unknown): EditorThemeType {
	return value === 'vscode' || value === 'zed' ? value : DEFAULT_THEME_EXPORT_PREFERENCES.editorType;
}

function asThemeAppearance(value: unknown): ThemeAppearance {
	return value === 'light' || value === 'dark' ? value : DEFAULT_THEME_EXPORT_PREFERENCES.appearance;
}

export function parseThemeExportPreferences(value: unknown): ThemeExportPreferences {
	if (!value || typeof value !== 'object') {
		return { ...DEFAULT_THEME_EXPORT_PREFERENCES };
	}

	const parsed = value as Partial<ThemeExportPreferences>;
	return {
		editorType: asEditorThemeType(parsed.editorType),
		appearance: asThemeAppearance(parsed.appearance),
		saveOnCopy:
			typeof parsed.saveOnCopy === 'boolean' ? parsed.saveOnCopy : DEFAULT_THEME_EXPORT_PREFERENCES.saveOnCopy,
		boostCoefficient: asNonNegativeFiniteNumber(
			parsed.boostCoefficient,
			DEFAULT_THEME_EXPORT_PREFERENCES.boostCoefficient
		)
	};
}

export function loadThemeExportPreferences(): ThemeExportPreferences {
	if (!browser) {
		return { ...DEFAULT_THEME_EXPORT_PREFERENCES };
	}
	const stored = localStorage.getItem(THEME_EXPORT_STORAGE_KEY);
	if (stored) {
		try {
			return parseThemeExportPreferences(JSON.parse(stored));
		} catch {
			return { ...DEFAULT_THEME_EXPORT_PREFERENCES };
		}
	}
	return { ...DEFAULT_THEME_EXPORT_PREFERENCES };
}

export function saveThemeExportPreferences(preferences: ThemeExportPreferences) {
	if (!browser) return;
	localStorage.setItem(THEME_EXPORT_STORAGE_KEY, JSON.stringify(preferences));
}

export function clearThemeExportPreferences() {
	if (!browser) return;
	localStorage.removeItem(THEME_EXPORT_STORAGE_KEY);
}
