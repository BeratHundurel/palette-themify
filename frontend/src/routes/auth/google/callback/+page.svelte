<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { authStore } from '$lib/stores/auth.svelte';
	import { appStore } from '$lib/stores/app.svelte';
	import { setAuthToken, getCurrentUser } from '$lib/api/auth';
	import toast from 'svelte-french-toast';

	let error = $state<string | null>(null);

	onMount(async () => {
		const token = page.url.searchParams.get('token');
		const errorParam = page.url.searchParams.get('error');

		if (errorParam) {
			error = page.url.searchParams.get('error_description') || 'Authorization was denied';
			return;
		}

		if (!token) {
			error = 'No token received';
			return;
		}

		try {
			setAuthToken(token);
			const { user } = await getCurrentUser();
			authStore.setUser(user);
			await appStore.syncPalettesOnAuth();
			toast.success('Successfully signed in with Google!');
			window.location.href = '/';
		} catch {
			error = 'Failed to complete Google sign in. Please try again.';
		}
	});
</script>

<div class="flex min-h-screen items-center justify-center bg-zinc-900">
	<div class="text-center">
		{#if error}
			<div class="mb-4 rounded-lg border border-red-500/50 bg-red-500/10 p-4">
				<p class="text-red-400">{error}</p>
			</div>
			<button
				class="rounded-md bg-zinc-800 px-4 py-2 text-zinc-300 transition-colors hover:bg-zinc-700"
				onclick={() => (window.location.href = '/')}
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
