import { buildURL, ensureOk } from './base';
import { getAuthHeaders } from './auth';
import type { SavedThemeItem } from '$lib/types/theme';

export type ThemesResponse = {
	themes: SavedThemeItem[];
};

export type SavedThemeResponse = {
	message: string;
	theme: SavedThemeItem;
};

export type SavedThemesBatchResponse = {
	message: string;
	themes: SavedThemeItem[];
};

export async function getThemes(): Promise<ThemesResponse> {
	const response = await fetch(buildURL('/themes'), {
		method: 'GET',
		headers: getAuthHeaders()
	});

	await ensureOk(response);
	return response.json();
}

export async function saveTheme(theme: SavedThemeItem): Promise<SavedThemeResponse> {
	const response = await fetch(buildURL('/themes'), {
		method: 'POST',
		headers: getAuthHeaders(),
		body: JSON.stringify(theme)
	});

	await ensureOk(response);
	return response.json();
}

export async function saveThemes(themes: SavedThemeItem[]): Promise<SavedThemesBatchResponse> {
	const response = await fetch(buildURL('/themes/batch'), {
		method: 'POST',
		headers: getAuthHeaders(),
		body: JSON.stringify({ themes })
	});

	await ensureOk(response);
	return response.json();
}

export async function updateTheme(themeId: string, theme: SavedThemeItem): Promise<SavedThemeResponse> {
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

export async function deleteThemes(themeIds: string[]): Promise<{ message: string; deleted: number }> {
	const response = await fetch(buildURL('/themes'), {
		method: 'DELETE',
		headers: getAuthHeaders(),
		body: JSON.stringify({ ids: themeIds })
	});

	await ensureOk(response);
	return response.json();
}

export async function shareTheme(themeId: string): Promise<SavedThemeResponse> {
	const response = await fetch(buildURL(`/themes/${themeId}/share`), {
		method: 'POST',
		headers: getAuthHeaders()
	});

	await ensureOk(response);
	return response.json();
}

export async function unshareTheme(themeId: string): Promise<SavedThemeResponse> {
	const response = await fetch(buildURL(`/themes/${themeId}/share`), {
		method: 'DELETE',
		headers: getAuthHeaders()
	});

	await ensureOk(response);
	return response.json();
}
