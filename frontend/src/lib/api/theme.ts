import type { Color } from '$lib/types/color';
import type { Theme, ThemeGenerationResponse, ThemeOverrides } from '$lib/types/theme';
import type { ApplyParams, EditorThemeType, GenerateThemeRequest, ThemeAppearance } from '$lib/types/themeApi';

import { buildURL, buildZigURL, ensureOk } from './base';

export async function generateTheme(
	colors: Color[],
	type: EditorThemeType,
	name?: string,
	overrides?: ThemeOverrides | null,
	appearance?: ThemeAppearance | null,
	boostCoefficient?: number | null
): Promise<ThemeGenerationResponse> {
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
): Promise<ThemeGenerationResponse> {
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
