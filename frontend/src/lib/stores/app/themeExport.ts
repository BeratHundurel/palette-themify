import { browser } from '$app/environment';
import type { EditorThemeType } from '../../api/theme';

const THEME_EXPORT_STORAGE_KEY = 'themeExportPreferences';

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
