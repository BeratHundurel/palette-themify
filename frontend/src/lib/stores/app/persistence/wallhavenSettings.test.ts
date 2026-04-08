import { describe, expect, it } from 'vitest';

import { parseWallhavenSettings } from '$lib/stores/app/persistence/wallhavenSettings';
import { DEFAULT_WALLHAVEN_SETTINGS } from '$lib/types/wallhaven';

describe('parseWallhavenSettings', () => {
	it('returns defaults for empty payloads', () => {
		expect(parseWallhavenSettings(null)).toEqual(DEFAULT_WALLHAVEN_SETTINGS);
		expect(parseWallhavenSettings(undefined)).toEqual(DEFAULT_WALLHAVEN_SETTINGS);
	});

	it('parses valid user settings and filters invalid ratios', () => {
		const parsed = parseWallhavenSettings({
			categories: '010',
			purity: '111',
			sorting: 'favorites',
			order: 'asc',
			topRange: '3M',
			ratios: ['16x9', '9x16', 42, null],
			apikey: 'secret'
		});

		expect(parsed).toEqual({
			categories: '010',
			purity: '111',
			sorting: 'favorites',
			order: 'asc',
			topRange: '3M',
			ratios: ['16x9', '9x16'],
			apikey: 'secret'
		});
	});

	it('falls back to defaults for non-string primitive values', () => {
		const parsed = parseWallhavenSettings({
			categories: 101,
			purity: false,
			sorting: {},
			order: [],
			topRange: 42,
			ratios: '16x9',
			apikey: undefined
		});

		expect(parsed).toEqual(DEFAULT_WALLHAVEN_SETTINGS);
	});

	it('keeps empty apikey and resets ratios when value is not an array', () => {
		const parsed = parseWallhavenSettings({
			categories: '001',
			purity: '010',
			sorting: 'date_added',
			order: 'desc',
			topRange: '6M',
			ratios: null,
			apikey: ''
		});

		expect(parsed).toEqual({
			categories: '001',
			purity: '010',
			sorting: 'date_added',
			order: 'desc',
			topRange: '6M',
			ratios: [],
			apikey: ''
		});
	});
});
