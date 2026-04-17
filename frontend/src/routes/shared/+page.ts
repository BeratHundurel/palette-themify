import { getSharedItems } from '$lib/api/shared';
import type { PageLoad } from './$types';

export const load = (async ({ fetch, url }) => {
	const q = (url.searchParams.get('q') ?? '').trim();
	const sortParam = url.searchParams.get('sort');
	const sort = sortParam === 'oldest' || sortParam === 'name' ? sortParam : 'newest';

	try {
		const response = await getSharedItems({ q, sort, limit: 200, fetchFn: fetch });
		return {
			searchQuery: q,
			sortBy: sort,
			items: response.items,
			errorMessage: ''
		};
	} catch (error) {
		return {
			searchQuery: q,
			sortBy: sort,
			items: [],
			errorMessage: error instanceof Error ? error.message : 'Failed to load shared items.'
		};
	}
}) satisfies PageLoad;
