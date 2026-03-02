import { buildURL, ensureOk } from './base';
import { getAuthHeaders } from './auth';

export type PreferencesPayload = Record<string, unknown>;

export type PreferencesResponse<T = PreferencesPayload> = {
	preferences: T | null;
};

export async function getPreferences<T = PreferencesPayload>(): Promise<PreferencesResponse<T>> {
	const response = await fetch(buildURL('/auth/preferences'), {
		method: 'GET',
		headers: getAuthHeaders()
	});

	await ensureOk(response);
	return response.json();
}

export async function savePreferences<T = PreferencesPayload>(preferences: T): Promise<PreferencesResponse<T>> {
	const response = await fetch(buildURL('/auth/preferences'), {
		method: 'PUT',
		headers: getAuthHeaders(),
		body: JSON.stringify(preferences)
	});

	await ensureOk(response);
	return response.json();
}
