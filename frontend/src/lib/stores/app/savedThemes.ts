import { browser } from '$app/environment';
import type { SavedThemeItem } from '$lib/types/theme';

const SAVED_THEMES_STORAGE_KEY = 'savedThemes';
const DEFAULT_BOOST_COEFFICIENT = 1;

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === 'object' && value !== null;
}

function normalizeSavedThemeItem(value: unknown): SavedThemeItem | null {
	if (!isRecord(value)) return null;

	const id = typeof value.id === 'string' ? value.id : null;
	const name = typeof value.name === 'string' ? value.name : null;
	const editorType = value.editorType === 'vscode' || value.editorType === 'zed' ? value.editorType : null;
	const createdAt = typeof value.createdAt === 'string' ? value.createdAt : null;
	const themeResult = isRecord(value.themeResult) ? value.themeResult : null;

	if (!id || !name || !editorType || !createdAt || !themeResult || !isRecord(themeResult.theme)) {
		return null;
	}

	const themeOverrides = isRecord(themeResult.themeOverrides) ? themeResult.themeOverrides : {};
	const rawThemeOverrides = isRecord(themeResult.rawThemeOverrides) ? themeResult.rawThemeOverrides : {};
	const colors = Array.isArray(themeResult.colors)
		? themeResult.colors.filter((color): color is { hex: string } => isRecord(color) && typeof color.hex === 'string')
		: [];

	const boostCoefficient =
		typeof themeResult.boostCoefficient === 'number' ? themeResult.boostCoefficient : DEFAULT_BOOST_COEFFICIENT;

	return {
		id,
		name,
		editorType,
		themeResult: {
			theme: themeResult.theme as SavedThemeItem['themeResult']['theme'],
			themeOverrides,
			rawThemeOverrides,
			colors,
			boostCoefficient
		},
		createdAt,
		signature: typeof value.signature === 'string' ? value.signature : undefined
	};
}

export function loadSavedThemes(): SavedThemeItem[] {
	if (!browser) return [];
	try {
		const stored = localStorage.getItem(SAVED_THEMES_STORAGE_KEY);
		if (!stored) return [];
		const parsed = JSON.parse(stored) as unknown;
		if (!Array.isArray(parsed)) return [];
		return parsed.map((item) => normalizeSavedThemeItem(item)).filter((item): item is SavedThemeItem => item !== null);
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
