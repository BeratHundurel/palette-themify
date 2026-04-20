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
	import { appStore } from '$lib/stores/app/store.svelte';
	import TutorialButton from '$lib/components/tutorial/TutorialButton.svelte';
	import Search from '$lib/components/search/Search.svelte';
	import { popoverStore } from '$lib/stores/popovers.svelte';
	import ThemeExportPopover from '$lib/components/toolbar/theme-export/ThemeExportPopover.svelte';
	import { getDesktopAppDownloadUrl, isDesktopApp } from '$lib/platform';
	import BrandButton from '$lib/components/ui/BrandButton.svelte';
	import BrandLinks from '$lib/components/ui/BrandLinks.svelte';
	import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';
	import PromptDialog from '$lib/components/ui/PromptDialog.svelte';

	let showAuthModal = $state(false);
	const desktopAppDownloadUrl = getDesktopAppDownloadUrl();

	onMount(async () => {
		await authStore.init();
		appStore.loadSavedPalettes();
		await appStore.syncPreferencesOnAuth();
		await appStore.syncSavedThemesOnAuth();
	});
</script>

<Toaster />

<div class={`relative flex flex-col bg-black text-zinc-300 ${isDesktopApp ? 'h-full' : 'h-svh'}`}>
	<enhanced:img
		src="../lib/assets/palette.jpg"
		alt="Palette"
		class="absolute top-0 left-0 h-full w-full object-cover"
	/>

	<div class="absolute top-0 left-0 z-10 h-full w-full bg-black/60"></div>

	<div class="z-40 h-max w-full p-4">
		<div class="grid grid-cols-3 items-center gap-4">
			<div class="flex min-w-0 items-center gap-4 justify-self-start">
				<TutorialButton />
				{#if !isDesktopApp}
					<a
						href={desktopAppDownloadUrl}
						download
						class="border-brand/50 hover:shadow-brand-lg inline-flex w-40 cursor-pointer items-center justify-center rounded-md border bg-zinc-900 py-2 text-sm font-medium transition-[background-color,border-color,box-shadow,color] duration-300"
					>
						Download App
					</a>
				{/if}
			</div>

			<div class="flex min-w-0 justify-center">
				<Search />
			</div>

			<div class="flex min-w-0 items-center justify-end gap-4 justify-self-end">
				<BrandLinks href="/community">Community</BrandLinks>
				{#if authStore.state.isAuthenticated}
					<UserProfile />
				{:else}
					<BrandButton class="flex w-32 items-center justify-center gap-2" onclick={() => (showAuthModal = true)}>
						<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
							/>
						</svg>
						<span>Sign In</span>
					</BrandButton>
				{/if}
			</div>
		</div>
	</div>

	<Toolbar />

	<div class="relative z-30 flex h-full w-full flex-col items-center overflow-hidden">
		<UploadOverlay />

		<div class="flex flex-1 flex-col items-center justify-center">
			<Canvas />
		</div>

		<div class="w-full pb-4">
			<div class="flex w-full justify-center">
				<PaletteGrid />
			</div>
		</div>
	</div>
</div>

<AuthModal bind:isOpen={showAuthModal} />

{#if popoverStore.state.current === 'themeExport'}
	<ThemeExportPopover />
{/if}

<TutorialStart />

<Tutorial />

<ConfirmDialog />
<PromptDialog />
