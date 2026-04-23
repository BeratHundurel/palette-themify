<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { authStore } from '$lib/stores/auth.svelte';
	import { appStore } from '$lib/stores/app/store.svelte';
	import toast from 'svelte-french-toast';
	import type { AuthResponse } from '$lib/api/auth';
	import * as authApi from '$lib/api/auth';
	import { resolve } from '$app/paths';

	let { data } = $props<{
		data: {
			auth: AuthResponse | null;
			error: string | null;
		};
	}>();

	let completed = $state(false);

	async function finalizeGoogleSignIn() {
		await Promise.all([
			appStore.syncPalettesOnAuth(),
			appStore.syncPreferencesOnAuth(),
			appStore.syncSavedThemesOnAuth()
		]);

		toast.success('Successfully signed in with Google!');
		await goto(resolve('/'), { replaceState: true });
	}

	onMount(() => {
		if (completed || data.error || !data.auth) return;

		completed = true;

		authApi.setAuthToken(data.auth.token);
		authStore.setUser(data.auth.user);
		finalizeGoogleSignIn();
	});
</script>

<div class="flex min-h-screen items-center justify-center bg-zinc-900">
	<div class="text-center">
		{#if data.error}
			<div class="mb-4 rounded-lg border border-red-500/50 bg-red-500/10 p-4">
				<p class="text-red-400">{data.error}</p>
			</div>
			<button
				class="rounded-md bg-zinc-800 px-4 py-2 text-zinc-300 transition-colors hover:bg-zinc-700"
				onclick={() => goto(resolve('/'))}
			>
				Go back home
			</button>
		{:else}
			<div class="flex items-center justify-center">
				<svg
					class="mr-3 -ml-1 h-8 w-8 animate-spin text-zinc-300"
					xmlns="http://www.w3.org/2000/svg"
					fill="none"
					viewBox="0 0 24 24"
				>
					<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
					<path
						class="opacity-75"
						fill="currentColor"
						d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
					></path>
				</svg>
				<span class="text-lg text-zinc-300">Completing sign in...</span>
			</div>
		{/if}
	</div>
</div>
