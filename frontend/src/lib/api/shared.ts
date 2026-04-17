import { buildURL, ensureOk } from './base';
import type { SharedItemSort, SharedItemsResponse } from '$lib/types/shared';

export async function getSharedItems(params?: {
	q?: string;
	sort?: SharedItemSort;
	limit?: number;
	fetchFn?: typeof fetch;
}): Promise<SharedItemsResponse> {
	const doFetch = params?.fetchFn ?? fetch;
	const res = await doFetch(
		buildURL('/shared-items', {
			q: params?.q?.trim() || null,
			sort: params?.sort || 'newest',
			limit: params?.limit ?? 120
		}),
		{ method: 'GET' }
	);

	await ensureOk(res);
	return res.json();
}
