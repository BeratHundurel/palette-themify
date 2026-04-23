import { exchangeGoogleAuthCode } from '$lib/api/auth';
import type { PageLoad } from './$types';

export const ssr = false;

export const load = (async ({ url }) => {
	const errorParam = url.searchParams.get('error');
	const errorDescription = url.searchParams.get('error_description');
	const exchangeCode = url.searchParams.get('exchange_code');

	if (errorParam) {
		return {
			auth: null,
			error: errorDescription || 'Authorization was denied'
		};
	}

	if (!exchangeCode) {
		return {
			auth: null,
			error: 'No exchange code received'
		};
	}

	try {
		const auth = await exchangeGoogleAuthCode(exchangeCode);

		return {
			auth,
			error: null
		};
	} catch {
		return {
			auth: null,
			error: 'Failed to complete Google sign in.'
		};
	}
}) satisfies PageLoad;
