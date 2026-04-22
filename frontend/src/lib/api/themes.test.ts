import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { SavedThemeItem } from '$lib/types/theme';

vi.mock('./auth', () => ({
	getAuthHeaders: vi.fn(() => ({
		'Content-Type': 'application/json',
		Authorization: 'Bearer test-token'
	}))
}));

import { saveThemes } from './themes';

function makeTheme(id: string): SavedThemeItem {
	return {
		id,
		name: `Theme ${id}`,
		editorType: 'vscode',
		createdAt: '2026-01-01T00:00:00.000Z',
		themeResult: {
			theme: { name: `Theme ${id}` } as SavedThemeItem['themeResult']['theme'],
			themeOverrides: {},
			rawThemeOverrides: {},
			colors: [{ hex: '#112233' }],
			boostCoefficient: 1
		},
		signature: `sig-${id}`
	};
}

describe('themes api', () => {
	beforeEach(() => {
		vi.restoreAllMocks();
	});

	it('sends one batch request with all themes', async () => {
		const themes = [makeTheme('one'), makeTheme('two')];
		const responsePayload = { message: 'ok', themes };
		const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue({
			ok: true,
			json: async () => responsePayload
		} as Response);

		const result = await saveThemes(themes);

		expect(fetchMock).toHaveBeenCalledWith('http://localhost:8088/themes/batch', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
				Authorization: 'Bearer test-token'
			},
			body: JSON.stringify({ themes })
		});
		expect(result).toEqual(responsePayload);
	});

	it('throws normalized API errors for failed responses', async () => {
		vi.spyOn(globalThis, 'fetch').mockResolvedValue({
			ok: false,
			status: 400,
			json: async () => ({ error: 'Themes are required' })
		} as Response);

		await expect(saveThemes([])).rejects.toThrow('Themes are required');
	});
});
