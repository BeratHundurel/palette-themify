import type { EditorThemeType, ThemeAppearance } from '$lib/api/theme';
import { COLOR_REGEX } from '$lib/types/color';
import type { ThemeGenerationResponse, ThemeOverrides } from '$lib/types/theme';

export const THEME_NAME_DEBOUNCE_MS = 300;

export const overrideFields: Array<{ key: keyof ThemeOverrides; label: string; hint: string }> = [
	{ key: 'background', label: 'Background', hint: 'Base background (c0)' },
	{ key: 'foreground', label: 'Foreground', hint: 'Primary text color' },
	{ key: 'c1', label: 'Properties', hint: 'Fields/member access' },
	{ key: 'c2', label: 'Functions', hint: 'Function/method names' },
	{ key: 'c3', label: 'Strings', hint: 'String literals' },
	{ key: 'c4', label: 'Special Strings', hint: 'Escapes and punctuation' },
	{ key: 'c5', label: 'Types', hint: 'Classes/namespaces/types' },
	{ key: 'c6', label: 'Keywords', hint: 'Language keywords' },
	{ key: 'c7', label: 'Parameters', hint: 'Params and enums' },
	{ key: 'c8', label: 'Constructors', hint: 'Constructors/operators' },
	{ key: 'c9', label: 'Builtins', hint: 'Builtins and variants' },
	{ key: 'constants', label: 'Constants', hint: 'Numbers/boolean/constants' }
];

export function getThemeVersionKey(type: EditorThemeType, appearance: ThemeAppearance): string {
	return `${type}:${appearance}`;
}

export function cloneThemeResponse(response: ThemeGenerationResponse): ThemeGenerationResponse {
	return JSON.parse(JSON.stringify(response)) as ThemeGenerationResponse;
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

export function normalizeHex(value: string | null | undefined): string | null {
	if (!value) return null;
	if (!COLOR_REGEX.test(value)) return null;
	return value.slice(0, 7).toUpperCase();
}
