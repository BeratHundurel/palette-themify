import { getCommunityItems } from '$lib/api/community';
import type { PageLoad } from './$types';

export const load = (async ({ fetch, url }) => {
	const q = (url.searchParams.get('q') ?? '').trim();
	const sortParam = url.searchParams.get('sort');
	const sort = sortParam === 'oldest' || sortParam === 'name' ? sortParam : 'newest';

	try {
		const response = await getCommunityItems({ q, sort, limit: 200, fetchFn: fetch });
		return {
			searchQuery: q,
			sortBy: sort,
			items: response.items,
			hasLoadError: false
		};
	} catch (error) {
		console.error('Failed to load community items:', error);
		return {
			searchQuery: q,
			sortBy: sort,
			items: [],
			hasLoadError: true
		};
	}
}) satisfies PageLoad;
