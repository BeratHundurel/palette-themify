<script lang="ts">
	import { authStore } from '$lib/stores/auth.svelte';
	import { appStore } from '$lib/stores/app.svelte';
	import toast from 'svelte-french-toast';

	let showDropdown = $state(false);

	async function handleLogout() {
		try {
			await authStore.logout();

			await appStore.loadSavedPalettes();
			await appStore.loadSavedWorkspaces();

			showDropdown = false;
			toast.success('Logged out successfully');
		} catch {
			toast.error('Could not sign out. Please try again.');
		}
	}

	function handleClickOutside(event: MouseEvent) {
		if (event.target instanceof Element && !event.target.closest('.user-profile')) {
			showDropdown = false;
		}
	}

	$effect(() => {
		if (showDropdown) {
			document.addEventListener('click', handleClickOutside);
			return () => document.removeEventListener('click', handleClickOutside);
		}
	});
</script>

{#if authStore.state.user}
	<div class="user-profile relative">
		<button
			onclick={() => (showDropdown = !showDropdown)}
			class="flex cursor-pointer items-center space-x-2 rounded-full bg-white/20 px-3 py-2 text-zinc-300 transition-[background-color] hover:bg-white/30"
		>
			<div class="bg-brand flex h-8 w-8 items-center justify-center rounded-full text-sm font-semibold text-zinc-300">
				{authStore.state.user.name.charAt(0).toUpperCase()}
			</div>
			<div class="hidden sm:block">
				<span class="text-sm font-medium">{authStore.state.user.name}</span>
				{#if authStore.state.user.email === 'demo@imagepalette.com'}
					<span class="ml-1 rounded bg-orange-500/30 px-1.5 py-0.5 text-xs font-medium text-orange-200">Demo</span>
				{/if}
			</div>
			<svg
				class="h-4 w-4 transition-transform duration-300"
				class:rotate-180={showDropdown}
				fill="none"
				stroke="currentColor"
				viewBox="0 0 24 24"
			>
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
			</svg>
		</button>

		{#if showDropdown}
			<div class="border-brand/50 absolute top-full right-0 z-50 mt-2 w-64 rounded-md border bg-zinc-900 shadow-lg">
				<div class="border-b border-zinc-700 p-4">
					<div class="flex items-center space-x-3">
						<div class="bg-brand flex h-10 w-10 items-center justify-center rounded-full text-zinc-300">
							{authStore.state.user.name.charAt(0).toUpperCase()}
						</div>
						<div class="min-w-0 flex-1">
							<div class="flex items-center space-x-2">
								<p class="truncate text-sm font-semibold text-zinc-300">{authStore.state.user.name}</p>
								{#if authStore.state.user.email === 'demo@imagepalette.com'}
									<span class="rounded bg-orange-500/20 px-2 py-0.5 text-xs font-medium text-orange-300"> Demo </span>
								{/if}
							</div>
							<p class="truncate text-xs text-zinc-400">{authStore.state.user.email}</p>
							{#if authStore.state.user.email === 'demo@imagepalette.com'}
								<p class="mt-1 text-xs text-zinc-500">Explore features without signing up</p>
							{/if}
						</div>
					</div>
				</div>

				<div class="py-2">
					<button
						type="button"
						onclick={handleLogout}
						class="flex w-full cursor-pointer items-center px-4 py-2 text-sm text-zinc-300 transition-colors hover:bg-zinc-800 hover:text-zinc-300"
					>
						<svg class="mr-3 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
							/>
						</svg>
						Sign Out
					</button>
				</div>
			</div>
		{/if}
	</div>
{/if}
