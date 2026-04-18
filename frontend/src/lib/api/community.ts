import { buildURL, ensureOk } from './base';
import type { CommunityItemSort, CommunityItemsResponse } from '$lib/types/community';

export async function getCommunityItems(params?: {
	q?: string;
	sort?: CommunityItemSort;
	limit?: number;
	fetchFn?: typeof fetch;
}): Promise<CommunityItemsResponse> {
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
