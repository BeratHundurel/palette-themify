import type { EditorThemeType, ThemeAppearance } from '$lib/api/theme';
import { COLOR_REGEX } from '$lib/constants';
import type { SavedThemeItem, ThemeOverrides, ThemeResponse } from '$lib/types/theme';

export const THEME_NAME_DEBOUNCE_MS = 300;

export const overrideFields: Array<{ key: keyof ThemeOverrides; label: string; hint: string }> = [
	{ key: 'background', label: 'Background', hint: 'Base background (c0)' },
	{ key: 'foreground', label: 'Foreground', hint: 'Primary text color' },
	{ key: 'c1', label: 'C1', hint: 'Properties/fields' },
	{ key: 'c2', label: 'C2', hint: 'Functions' },
	{ key: 'c3', label: 'C3', hint: 'Strings' },
	{ key: 'c4', label: 'C4', hint: 'Special strings/punctuation' },
	{ key: 'c5', label: 'C5', hint: 'Namespaces/types' },
	{ key: 'c6', label: 'C6', hint: 'Keywords' },
	{ key: 'c7', label: 'C7', hint: 'Parameters/enums' },
	{ key: 'c8', label: 'C8', hint: 'Constructors' },
	{ key: 'c9', label: 'C9', hint: 'Builtins/variants' },
	{ key: 'constants', label: 'Constants', hint: 'Numbers/boolean/constants' }
];

export function getThemeVersionKey(type: EditorThemeType, appearance: ThemeAppearance): string {
	return `${type}:${appearance}`;
}

export function cloneThemeResponse(response: ThemeResponse): ThemeResponse {
	return JSON.parse(JSON.stringify(response)) as ThemeResponse;
}

export function validateThemeName(name: string): string | null {
	const trimmedName = name.trim();

	if (trimmedName.length === 0) {
		return 'Theme name cannot be empty';
	}

	const invalidChars = /[<>:"/\\|?*]/;
	if (invalidChars.test(trimmedName)) {
		return 'Theme name contains invalid characters: < > : " / \\ | ? *';
	}

	return null;
}

export function normalizeThemeName(name: string): string {
	return name.trim().toLowerCase();
}

export function getThemeSignature(result: SavedThemeItem['themeResult'] | null): string {
	if (!result) return '';

	try {
		return JSON.stringify(result);
	} catch {
		return '';
	}
}

export function normalizeHex(value: string | null | undefined): string | null {
	if (!value) return null;
	if (!COLOR_REGEX.test(value)) return null;
	return value.slice(0, 7).toUpperCase();
}

export function getBaseColorLabel(color: string, baseOverrides: ThemeOverrides): string | null {
	const match = overrideFields.find((entry) => baseOverrides[entry.key] === color);
	return match?.label ?? null;
}
