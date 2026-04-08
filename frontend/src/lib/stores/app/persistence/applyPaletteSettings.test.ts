import { describe, expect, it } from 'vitest';

import { parseApplyPaletteSettings } from '$lib/stores/app/persistence/applyPaletteSettings';
import { DEFAULT_APPLY_PALETTE_SETTINGS } from '$lib/types/applyPaletteSettings';

describe('parseApplyPaletteSettings', () => {
	it('returns defaults for non-object payloads', () => {
		expect(parseApplyPaletteSettings(null)).toEqual(DEFAULT_APPLY_PALETTE_SETTINGS);
		expect(parseApplyPaletteSettings(undefined)).toEqual(DEFAULT_APPLY_PALETTE_SETTINGS);
		expect(parseApplyPaletteSettings('invalid')).toEqual(DEFAULT_APPLY_PALETTE_SETTINGS);
	});

	it('keeps valid numeric values', () => {
		const parsed = parseApplyPaletteSettings({
			luminosity: 1.2,
			nearest: 9,
			power: 3.5,
			maxDistance: 120
		});

		expect(parsed).toEqual({
			luminosity: 1.2,
			nearest: 9,
			power: 3.5,
			maxDistance: 120
		});
	});

	it('falls back field-by-field when value types are invalid', () => {
		const parsed = parseApplyPaletteSettings({
			luminosity: '1.2',
			nearest: NaN,
			power: false,
			maxDistance: '120'
		});

		expect(parsed).toEqual({
			luminosity: DEFAULT_APPLY_PALETTE_SETTINGS.luminosity,
			nearest: DEFAULT_APPLY_PALETTE_SETTINGS.nearest,
			power: DEFAULT_APPLY_PALETTE_SETTINGS.power,
			maxDistance: DEFAULT_APPLY_PALETTE_SETTINGS.maxDistance
		});
	});

	it('clamps values to supported ranges', () => {
		const parsed = parseApplyPaletteSettings({
			luminosity: 99,
			nearest: 100,
			power: 0.1,
			maxDistance: -5
		});

		expect(parsed).toEqual({
			luminosity: 3,
			nearest: 30,
			power: 0.5,
			maxDistance: 0
		});
	});

	it('rounds nearest to the closest integer', () => {
		const parsed = parseApplyPaletteSettings({ nearest: 9.6 });

		expect(parsed.nearest).toBe(10);
	});
});
