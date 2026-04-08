import { beforeEach, describe, expect, it } from 'vitest';

import { clearSavedThemes, loadSavedThemes, saveSavedThemes } from '$lib/stores/app/persistence/savedThemes';

const STORAGE_KEY = 'savedThemes';

describe('savedThemes storage helpers', () => {
	beforeEach(() => {
		localStorage.clear();
	});

	it('returns empty array for missing or malformed storage values', () => {
		expect(loadSavedThemes()).toEqual([]);

		localStorage.setItem(STORAGE_KEY, '{broken json');
		expect(loadSavedThemes()).toEqual([]);

		localStorage.setItem(STORAGE_KEY, JSON.stringify({ not: 'an array' }));
		expect(loadSavedThemes()).toEqual([]);
	});

	it('keeps valid entries and filters invalid ones', () => {
		const valid = {
			id: 'theme-1',
			name: 'Theme One',
			editorType: 'vscode',
			createdAt: '2026-01-01T00:00:00.000Z',
			themeResult: {
				theme: { name: 'Theme One' },
				themeOverrides: { background: '#101010' },
				rawThemeOverrides: { background: '#101010' },
				colors: [{ hex: '#101010' }, { hex: 42 }],
				boostCoefficient: 1
			},
			signature: 'sig-1'
		};

		const invalid = {
			id: 123,
			name: null,
			editorType: 'bad-type',
			createdAt: null,
			themeResult: {}
		};

		localStorage.setItem(STORAGE_KEY, JSON.stringify([valid, invalid]));

		const loaded = loadSavedThemes();
		expect(loaded).toHaveLength(1);
		expect(loaded[0].id).toBe('theme-1');
		expect(loaded[0].signature).toBe('sig-1');
		expect(loaded[0].themeResult.colors).toEqual([{ hex: '#101010' }]);
	});

	it('save and clear roundtrip persisted values', () => {
		const payload = [
			{
				id: 'theme-2',
				name: 'Theme Two',
				editorType: 'zed',
				createdAt: '2026-01-02T00:00:00.000Z',
				themeResult: {
					theme: { name: 'Theme Two' },
					themeOverrides: {},
					rawThemeOverrides: {},
					colors: [{ hex: '#AABBCC' }],
					boostCoefficient: 1
				}
			}
		] as never;

		saveSavedThemes(payload);
		const loaded = loadSavedThemes();
		expect(loaded).toHaveLength(1);
		expect(loaded[0].id).toBe('theme-2');

		clearSavedThemes();
		expect(loadSavedThemes()).toEqual([]);
	});

	it('normalizes nested fields for partially valid entries', () => {
		localStorage.setItem(
			STORAGE_KEY,
			JSON.stringify([
				{
					id: 'theme-3',
					name: 'Theme Three',
					editorType: 'vscode',
					createdAt: '2026-01-03T00:00:00.000Z',
					themeResult: {
						theme: {},
						themeOverrides: null,
						rawThemeOverrides: 'invalid',
						colors: [{ hex: '#ABCDEF' }, { hex: 12 }],
						boostCoefficient: 'invalid'
					},
					signature: 123
				}
			])
		);

		const [entry] = loadSavedThemes();
		expect(entry).toMatchObject({
			id: 'theme-3',
			themeResult: {
				themeOverrides: {},
				rawThemeOverrides: {},
				colors: [{ hex: '#ABCDEF' }],
				boostCoefficient: 1
			}
		});
		expect(entry.signature).toBeUndefined();
	});
});
