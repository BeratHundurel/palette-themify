import { buildURL, ensureOk } from './base';
import type { QueryParamValue } from './base';
import type { WallhavenSearchResponse, WallhavenSettings } from '$lib/types/wallhaven';

export async function searchWallhaven(
	settings: WallhavenSettings,
	query: string,
	page = 1
): Promise<WallhavenSearchResponse> {
	const params: Record<string, QueryParamValue> = {
		q: query,
		categories: settings.categories,
		purity: settings.purity,
		sorting: settings.sorting,
		order: settings.order,
		topRange: settings.topRange,
		page
	};

	if (settings.ratios.length > 0) {
		params.ratios = settings.ratios.join(',');
	}

	const url = buildURL('/wallhaven/search', params);

	// Add API key header if provided
	const headers: HeadersInit = {};
	if (settings.apikey) {
		headers['X-API-Key'] = settings.apikey;
	}

	const res = await fetch(url, { headers });
	await ensureOk(res);
	return res.json() as Promise<WallhavenSearchResponse>;
}

export async function getWallpaper(id: string) {
	const res = await fetch(buildURL(`/wallhaven/w/${encodeURIComponent(id)}`));
	await ensureOk(res);
	return res.json();
}

export async function downloadImage(imageUrl: string): Promise<Blob> {
	const res = await fetch(buildURL('/wallhaven/download', { url: imageUrl }));
	await ensureOk(res);
	return res.blob();
}
