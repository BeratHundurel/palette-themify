import type {
	Color,
	SavePaletteRequest,
	GetPalettesResponse,
	SavePaletteResult,
	ThemeOverrides,
	ThemeResponse,
	ExtractPaletteResponse
} from '$lib/types/palette';

import { getAuthHeaders } from './auth';
import { buildURL, buildZigURL, ensureOk, ZIG_API_BASE } from './base';

export type ApplyParams = {
	luminosity: number;
	nearest: number;
	power: number;
	maxDistance: number;
};

export type GenerateThemeRequest = {
	colors: Color[];
	type: EditorThemeType;
	name?: string;
	overrides?: ThemeOverrides | null;
};

export type EditorThemeType = 'vscode' | 'zed';

export async function extractPalette(file: Blob | File): Promise<ExtractPaletteResponse> {
	if (!file) throw new Error('No files provided');

	let res: Response;
	try {
		res = await fetch(buildZigURL('/extract-palette'), {
			method: 'POST',
			body: file
		});
	} catch (err) {
		throw new Error(
			`Cannot connect to Zig API server at ${ZIG_API_BASE}. ${err instanceof Error ? err.message : 'Network error'}`
		);
	}

	await ensureOk(res);

	const text = await res.text();
	if (!text) {
		throw new Error('Empty response from Zig API server');
	}

	try {
		return JSON.parse(text);
	} catch {
		throw new Error(`Invalid JSON response: ${text.substring(0, 100)}`);
	}
}

export async function generateTheme(
	colors: Color[],
	type: EditorThemeType,
	name?: string,
	overrides?: ThemeOverrides | null
): Promise<ThemeResponse> {
	const payload: GenerateThemeRequest = { colors, type, name, overrides };

	const res = await fetch(buildZigURL('/generate-theme'), {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(payload)
	});
	await ensureOk(res);
	return res.json();
}

export async function downloadTheme(
	colors: Color[],
	type: EditorThemeType,
	name: string = 'Generated Theme'
): Promise<void> {
	const response = await generateTheme(colors, type, name);
	const json = JSON.stringify(response.theme, null, 2);
	const blob = new Blob([json], { type: 'application/json' });
	const url = URL.createObjectURL(blob);

	const link = document.createElement('a');
	link.href = url;
	link.download = `${name}.json`;
	document.body.appendChild(link);
	link.click();
	document.body.removeChild(link);

	URL.revokeObjectURL(url);
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

export async function savePalette(name: string, colors: Color[]): Promise<SavePaletteResult> {
	const payload: SavePaletteRequest = { name, palette: colors };

	const res = await fetch(buildURL('/palettes'), {
		method: 'POST',
		headers: getAuthHeaders(),
		body: JSON.stringify(payload)
	});
	await ensureOk(res);
	return res.json();
}

export async function getPalettes(): Promise<GetPalettesResponse> {
	const res = await fetch(buildURL('/palettes'), {
		method: 'GET',
		headers: getAuthHeaders()
	});
	await ensureOk(res);
	return res.json();
}

export async function deletePalette(id: string): Promise<{ message: string }> {
	const res = await fetch(buildURL(`/palettes/${id}`), {
		method: 'DELETE',
		headers: getAuthHeaders()
	});
	await ensureOk(res);
	return res.json();
}
