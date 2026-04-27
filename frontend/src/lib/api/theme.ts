import type { Color } from '$lib/types/color';
import type {
	Theme,
	ThemeGenerationResult,
	ThemeItem,
	ThemeOverrides,
	ThemeResponse,
	ThemesResponse
} from '$lib/types/theme';
import type { ApplyParams, EditorThemeType, GenerateThemeRequest, ThemeAppearance } from '$lib/types/theme';
import { getAuthHeaders } from './auth';

import { buildURL, buildZigURL, ensureOk } from './base';

export async function getThemes(): Promise<ThemesResponse<ThemeItem>> {
	const response = await fetch(buildURL('/themes'), {
		method: 'GET',
		headers: getAuthHeaders()
	});

	await ensureOk(response);
	return response.json();
}

export async function saveTheme(theme: ThemeItem): Promise<ThemeResponse<ThemeItem>> {
	const response = await fetch(buildURL('/themes'), {
		method: 'POST',
		headers: getAuthHeaders(),
		body: JSON.stringify(theme)
	});

	await ensureOk(response);
	return response.json();
}

export async function saveThemes(themes: ThemeItem[]): Promise<ThemesResponse<ThemeItem>> {
	const response = await fetch(buildURL('/themes/batch'), {
		method: 'POST',
		headers: getAuthHeaders(),
		body: JSON.stringify({ themes })
	});

	await ensureOk(response);
	return response.json();
}

export async function updateTheme(themeId: string, theme: ThemeItem): Promise<ThemeResponse<ThemeItem>> {
	const response = await fetch(buildURL(`/themes/${themeId}`), {
		method: 'PUT',
		headers: getAuthHeaders(),
		body: JSON.stringify(theme)
	});

	await ensureOk(response);
	return response.json();
}

export async function deleteTheme(themeId: string): Promise<void> {
	const response = await fetch(buildURL(`/themes/${themeId}`), {
		method: 'DELETE',
		headers: getAuthHeaders()
	});

	await ensureOk(response);
}

export async function deleteThemes(themeIds: string[]): Promise<void> {
	const response = await fetch(buildURL('/themes'), {
		method: 'DELETE',
		headers: getAuthHeaders(),
		body: JSON.stringify({ ids: themeIds })
	});

	await ensureOk(response);
}

export async function shareTheme(themeId: string): Promise<ThemeItem> {
	const response = await fetch(buildURL(`/themes/${themeId}/share`), {
		method: 'POST',
		headers: getAuthHeaders()
	});

	await ensureOk(response);
	return response.json();
}

export async function unshareTheme(themeId: string): Promise<ThemeItem> {
	const response = await fetch(buildURL(`/themes/${themeId}/share`), {
		method: 'DELETE',
		headers: getAuthHeaders()
	});

	await ensureOk(response);
	return response.json();
}

export async function generateTheme(
	colors: Color[],
	type: EditorThemeType,
	name?: string,
	overrides?: ThemeOverrides | null,
	appearance?: ThemeAppearance | null,
	boostCoefficient?: number | null
): Promise<ThemeGenerationResult> {
	const payload: GenerateThemeRequest = { colors, type, name, overrides, appearance, boostCoefficient };

	const res = await fetch(buildZigURL('/generate-theme'), {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(payload)
	});
	await ensureOk(res);
	return res.json();
}

export async function generateOverridable(
	theme: Theme,
	overrides?: ThemeOverrides | null,
	themeType: EditorThemeType = 'zed',
	appearance?: ThemeAppearance | null,
	boostCoefficient?: number | null
): Promise<ThemeGenerationResult> {
	const payload = {
		theme,
		themeType,
		appearance,
		boostCoefficient,
		...(overrides ? { ThemeOverrides: overrides } : {})
	};

	const res = await fetch(buildZigURL('/generate-overridable'), {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(payload)
	});
	await ensureOk(res);
	return res.json();
}

export async function applyPaletteBlob(imageBlob: Blob, colors: Color[], params: ApplyParams): Promise<Blob> {
	const formData = new FormData();
	formData.append('file', imageBlob, 'image.png');
	formData.append('palette', JSON.stringify(colors.map((c) => c.hex)));
	formData.append('luminosity', String(params.luminosity));
	formData.append('nearest', String(params.nearest));
	formData.append('power', String(params.power));
	formData.append('maxDistance', String(params.maxDistance));

	const res = await fetch(buildURL('/apply-palette'), {
		method: 'POST',
		body: formData
	});
	await ensureOk(res);
	return res.blob();
}
