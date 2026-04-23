import type { Color } from '$lib/types/color';
import type {
	SavePaletteRequest,
	GetPalettesResponse,
	SavePaletteResult,
	ExtractPaletteResponse,
	PaletteData
} from '$lib/types/palette';

import { getAuthHeaders } from './auth';
import { buildURL, buildZigURL, ensureOk, ZIG_API_BASE } from './base';

type SavePalettesBatchRequest = {
	palettes: Array<{
		name: string;
		palette: Color[];
	}>;
};

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
			`Cannot connect to Zig API server at ${ZIG_API_BASE}. ${err instanceof Error ? err.message : 'Network error'}`,
			{ cause: err }
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

export async function savePalettes(
	palettes: Array<{ name: string; palette: Color[] }>
): Promise<{ message: string; saved: number }> {
	const payload: SavePalettesBatchRequest = { palettes };

	const res = await fetch(buildURL('/palettes/batch'), {
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

export async function deletePalettes(ids: string[]): Promise<{ message: string; deleted: number }> {
	const res = await fetch(buildURL('/palettes'), {
		method: 'DELETE',
		headers: getAuthHeaders(),
		body: JSON.stringify({ ids })
	});
	await ensureOk(res);
	return res.json();
}

export async function sharePalette(id: string): Promise<{ message: string; palette: PaletteData }> {
	const res = await fetch(buildURL(`/palettes/${id}/share`), {
		method: 'POST',
		headers: getAuthHeaders()
	});
	await ensureOk(res);
	return res.json();
}

export async function unsharePalette(id: string): Promise<{ message: string; palette: PaletteData }> {
	const res = await fetch(buildURL(`/palettes/${id}/share`), {
		method: 'DELETE',
		headers: getAuthHeaders()
	});
	await ensureOk(res);
	return res.json();
}
