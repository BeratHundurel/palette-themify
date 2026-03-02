import { browser } from '$app/environment';
import type { EditorThemeType, ThemeAppearance } from '../../api/theme';

const THEME_EXPORT_STORAGE_KEY = 'themeExportPreferences';

export type ThemeExportPreferences = {
	editorType: EditorThemeType;
	appearance: ThemeAppearance;
	saveOnCopy: boolean;
};

export const DEFAULT_THEME_EXPORT_PREFERENCES: ThemeExportPreferences = {
	editorType: 'vscode',
	appearance: 'dark',
	saveOnCopy: true
};

export function loadThemeExportPreferences(): ThemeExportPreferences {
	if (!browser) {
		return { ...DEFAULT_THEME_EXPORT_PREFERENCES };
	}
	const stored = localStorage.getItem(THEME_EXPORT_STORAGE_KEY);
	if (stored) {
		const parsed = JSON.parse(stored);
		return {
			editorType: parsed.editorType || DEFAULT_THEME_EXPORT_PREFERENCES.editorType,
			appearance: parsed.appearance || DEFAULT_THEME_EXPORT_PREFERENCES.appearance,
			saveOnCopy: parsed.saveOnCopy ?? DEFAULT_THEME_EXPORT_PREFERENCES.saveOnCopy
		};
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
