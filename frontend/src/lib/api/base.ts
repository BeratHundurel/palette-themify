import { API_BASE, ZIG_API_BASE } from '$lib/config';
import { reportError } from '$lib/telemetry';

export { API_BASE, ZIG_API_BASE };

export type QueryParamValue = string | number | boolean | null;

export function buildURL(path: string, params?: Record<string, QueryParamValue>): string {
	const url = new URL(path, API_BASE);
	if (params) {
		for (const [k, v] of Object.entries(params)) {
			if (v !== null) url.searchParams.set(k, String(v));
		}
	}
	return url.toString();
}

export function buildZigURL(path: string, params?: Record<string, QueryParamValue>): string {
	const url = new URL(path, ZIG_API_BASE);
	if (params) {
		for (const [k, v] of Object.entries(params)) {
			if (v !== null) url.searchParams.set(k, String(v));
		}
	}
	return url.toString();
}

export async function ensureOk(res: Response): Promise<Response> {
	if (!res.ok) {
		let msg = `HTTP ${res.status}`;
		try {
			const data = await res.json().catch(() => null);
			if (data && typeof data === 'object') {
				const obj = data as { error?: string; message?: string };
				if (typeof obj.error === 'string') {
					msg = obj.error;
				} else if (typeof obj.message === 'string') {
					msg = obj.message;
				}
			}
		} catch {
			// ignore JSON parse errors
		}
		const error = new Error(msg);
		reportError(error, { source: 'api.response', url: res.url });
		throw error;
	}
	return res;
}
