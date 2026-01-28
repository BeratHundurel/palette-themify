<script lang="ts">
	import Toolbar from '$lib/components/toolbar/Toolbar.svelte';
	import { Toaster } from 'svelte-french-toast';
	import { onMount } from 'svelte';
	import Canvas from '$lib/components/Canvas.svelte';
	import UploadOverlay from '$lib/components/UploadOverlay.svelte';
	import PaletteGrid from '$lib/components/PaletteGrid.svelte';
	import AuthModal from '$lib/components/AuthModal.svelte';
	import UserProfile from '$lib/components/UserProfile.svelte';
	import Tutorial from '$lib/components/tutorial/Tutorial.svelte';
	import TutorialStart from '$lib/components/tutorial/TutorialStart.svelte';
	import { authStore } from '$lib/stores/auth.svelte';
	import { appStore } from '$lib/stores/app.svelte';
	import TutorialButton from '$lib/components/tutorial/TutorialButton.svelte';
	import { getSharedWorkspace } from '$lib/api/workspace';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import toast from 'svelte-french-toast';
	import { tick } from 'svelte';
	import Search from '$lib/components/search/Search.svelte';
	import { popoverStore } from '$lib/stores/popovers.svelte';
	import ThemeExportPopover from '$lib/components/toolbar/popovers/ThemeExportPopover.svelte';

	let showAuthModal = $state(false);

	onMount(async () => {
		await authStore.init();
		await appStore.loadSavedPalettes();
		await appStore.loadSavedWorkspaces();

		const shareToken = page.url.searchParams.get('share');
		if (shareToken) {
			const toastId = toast.loading('Loading shared workspace...');
			await tick();

			try {
				const workspace = await getSharedWorkspace(shareToken);
				await appStore.loadWorkspace(workspace);
				toast.success(`Loaded: ${workspace.name}`, { id: toastId });

				await goto(resolve('/'), { replaceState: true });
			} catch {
				toast.error('Could not load the shared workspace. Please check the link and try again.', {
					id: toastId
				});
			}
		}
	});

	function openAuthModal() {
		showAuthModal = true;
	}
</script>

<Toaster />

<div class="relative flex h-svh flex-col bg-black text-zinc-300">
	<enhanced:img
		src="../lib/assets/palette.jpg"
		alt="Palette"
		class="absolute top-0 left-0 h-full w-full object-cover"
	/>

	<div class="absolute top-0 left-0 z-10 h-full w-full bg-black/60"></div>

	<div class="z-40 h-max w-full p-4">
		<div class="flex flex-row items-center justify-between gap-4">
			<div class="flex items-center gap-4">
				<TutorialButton />
			</div>

			<div class="flex flex-1 justify-center">
				<Search />
			</div>

			<div class="flex items-center gap-4">
				{#if authStore.state.isAuthenticated}
					<UserProfile />
				{:else if !authStore.state.isLoading}
					<button
						onclick={openAuthModal}
						class="border-brand/50 hover:shadow-brand-lg flex w-32 cursor-pointer items-center justify-center gap-2 rounded-md border bg-zinc-900 py-2 text-sm font-medium transition-[background-color,border-color,box-shadow,color] duration-300"
					>
						<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
							/>
						</svg>
						<span>Sign In</span>
					</button>
				{:else}
					<!-- Placeholder to maintain layout during loading -->
					<div class="w-56"></div>
				{/if}
			</div>
		</div>
	</div>

	<div class="relative z-30 flex h-full w-full flex-col items-center justify-center overflow-hidden">
		<UploadOverlay />

		<Canvas />

		<Toolbar />

		<PaletteGrid />
	</div>
</div>

<AuthModal bind:isOpen={showAuthModal} />

{#if popoverStore.state.current === 'themeExport'}
	<ThemeExportPopover />
{/if}

<TutorialStart />

<Tutorial />
