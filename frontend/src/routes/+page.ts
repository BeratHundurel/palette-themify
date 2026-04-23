import type { PageLoad } from './$types';

import { appStore } from '$lib/stores/app/store.svelte';
import { authStore } from '$lib/stores/auth.svelte';

export const ssr = false;

export const load = (async () => {
	await authStore.init();

	await Promise.all([appStore.loadSavedPalettes(), appStore.syncPreferencesOnAuth(), appStore.syncSavedThemesOnAuth()]);

	return {};
}) satisfies PageLoad;
