import { buildURL, ensureOk } from './base';
import { getAuthHeaders } from './auth';
import type { SavedThemeItem } from '$lib/types/theme';

export type ThemesResponse = {
	themes: SavedThemeItem[];
};

export type ThemeResponse = {
	message: string;
	theme: SavedThemeItem;
};

export async function getThemes(): Promise<ThemesResponse> {
	const response = await fetch(buildURL('/themes'), {
		method: 'GET',
		headers: getAuthHeaders()
	});

	await ensureOk(response);
	return response.json();
}

export async function saveTheme(theme: SavedThemeItem): Promise<ThemeResponse> {
	const response = await fetch(buildURL('/themes'), {
		method: 'POST',
		headers: getAuthHeaders(),
		body: JSON.stringify(theme)
	});

	await ensureOk(response);
	return response.json();
}

export async function updateTheme(themeId: string, theme: SavedThemeItem): Promise<ThemeResponse> {
	const response = await fetch(buildURL(`/themes/${themeId}`), {
		method: 'PUT',
		headers: getAuthHeaders(),
		body: JSON.stringify(theme)
	});

	await ensureOk(response);
	return response.json();
}

export async function deleteTheme(themeId: string): Promise<{ message: string }> {
	const response = await fetch(buildURL(`/themes/${themeId}`), {
		method: 'DELETE',
		headers: getAuthHeaders()
	});

	await ensureOk(response);
	return response.json();
}
