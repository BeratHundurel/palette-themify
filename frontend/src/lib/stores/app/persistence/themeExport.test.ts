import { describe, expect, it } from 'vitest';

import { parseThemeExportPreferences } from '$lib/stores/app/persistence/themeExport';
import { DEFAULT_THEME_EXPORT_PREFERENCES } from '$lib/types/theme';

describe('parseThemeExportPreferences', () => {
	it('returns defaults for invalid payloads', () => {
		expect(parseThemeExportPreferences(null)).toEqual(DEFAULT_THEME_EXPORT_PREFERENCES);
		expect(parseThemeExportPreferences('bad')).toEqual(DEFAULT_THEME_EXPORT_PREFERENCES);
	});

	it('keeps valid values', () => {
		const parsed = parseThemeExportPreferences({
			editorType: 'zed',
			appearance: 'light',
			saveOnCopy: false,
			boostCoefficient: 2.5
		});

		expect(parsed).toEqual({
			editorType: 'zed',
			appearance: 'light',
			saveOnCopy: false,
			boostCoefficient: 2.5
		});
	});

	it('clamps boost coefficient and falls back for invalid fields', () => {
		const parsed = parseThemeExportPreferences({
			editorType: 'unknown',
			appearance: 'other',
			saveOnCopy: 'yes',
			boostCoefficient: 300
		});

		expect(parsed).toEqual({
			editorType: DEFAULT_THEME_EXPORT_PREFERENCES.editorType,
			appearance: DEFAULT_THEME_EXPORT_PREFERENCES.appearance,
			saveOnCopy: DEFAULT_THEME_EXPORT_PREFERENCES.saveOnCopy,
			boostCoefficient: 3
		});
	});

	it('uses default boost coefficient for negative and non-finite values', () => {
		expect(parseThemeExportPreferences({ boostCoefficient: -0.1 }).boostCoefficient).toBe(
			DEFAULT_THEME_EXPORT_PREFERENCES.boostCoefficient
		);
		expect(parseThemeExportPreferences({ boostCoefficient: Number.NaN }).boostCoefficient).toBe(
			DEFAULT_THEME_EXPORT_PREFERENCES.boostCoefficient
		);
		expect(parseThemeExportPreferences({ boostCoefficient: Number.POSITIVE_INFINITY }).boostCoefficient).toBe(
			DEFAULT_THEME_EXPORT_PREFERENCES.boostCoefficient
		);
	});
});
