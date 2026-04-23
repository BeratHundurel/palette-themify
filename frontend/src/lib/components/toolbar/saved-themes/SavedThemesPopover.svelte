<script lang="ts">
	import { cn } from '$lib/utils';
	import { isLocalId } from '$lib/localId';
	import ActionPillButton from '$lib/components/ui/ActionPillButton.svelte';
	import ColorSwatch from '$lib/components/ui/ColorSwatch.svelte';
	import DangerTextButton from '$lib/components/ui/DangerTextButton.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import IconDangerButton from '$lib/components/ui/IconDangerButton.svelte';
	import { appStore } from '$lib/stores/app/store.svelte';
	import { dialogStore } from '$lib/stores/dialog.svelte';
	import { popoverStore } from '$lib/stores/popovers.svelte';
	import { getDesktopSaveErrorMessage, isDesktopApp, saveThemeToEditorTarget } from '$lib/platform';
	import { detectThemeAppearance, detectThemeType } from '$lib/colorUtils';
	import toast from 'svelte-french-toast';
	import type { SavedThemeItem, Theme } from '$lib/types/theme';
	import { generateOverridable } from '$lib/api/theme';
	import type { EditorThemeType } from '$lib/types/themeApi';
	import { hydrateThemeExportResponse } from '../theme-export/session';

	async function handleThemeSave(item: SavedThemeItem) {
		if (!isDesktopApp) return;

		try {
			const themeJson = JSON.stringify(item.themeResult.theme, null, 2);
			await saveThemeToEditorTarget({
				editorType: item.editorType,
				themeName: item.name,
				themeJSON: themeJson
			});
			toast.success('Theme saved to editor folder');
			popoverStore.close('themes');
		} catch (error) {
			console.error('Error saving theme to editor folder:', error);
			toast.error(getDesktopSaveErrorMessage(error));
		}
	}

	async function handleThemeLoad(theme: Theme, editorType?: EditorThemeType) {
		const resolvedType = editorType ?? detectThemeType(theme);
		const resolvedAppearance = detectThemeAppearance(theme);
		try {
			const response = await generateOverridable(theme, null, resolvedType, resolvedAppearance);

			appStore.state.colors = [];
			appStore.state.image = null;
			appStore.state.imageLoaded = false;
			hydrateThemeExportResponse(response, resolvedType, resolvedAppearance);

			popoverStore.close('themes');
			popoverStore.state.current = 'themeExport';
		} catch {
			toast.error('Could not load the theme. Please check your JSON.');
		}
	}

	function hasStoredRawOverrides(item: SavedThemeItem): boolean {
		return 'rawThemeOverrides' in item.themeResult && typeof item.themeResult.rawThemeOverrides === 'object';
	}

	function loadSavedThemeResult(item: SavedThemeItem) {
		const resolvedAppearance = detectThemeAppearance(item.themeResult.theme);

		appStore.resetThemeExportSession();
		appStore.state.colors = [];
		appStore.state.image = null;
		appStore.state.imageLoaded = false;
		hydrateThemeExportResponse(item.themeResult, item.editorType, resolvedAppearance);

		popoverStore.close('themes');
		popoverStore.state.current = 'themeExport';
	}

	async function handleSavedThemeLoad(item: SavedThemeItem) {
		if (hasStoredRawOverrides(item)) {
			loadSavedThemeResult(item);
			return;
		}

		await handleThemeLoad(item.themeResult.theme, item.editorType);
	}

	async function handleThemeDelete(themeId: string, themeName: string) {
		const shouldDelete = await dialogStore.confirm({
			title: 'Delete saved theme?',
			message: `Are you sure you want to delete "${themeName}"?`,
			confirmLabel: 'Delete theme',
			variant: 'danger'
		});
		if (!shouldDelete) return;

		appStore.deleteTheme(themeId);
	}

	async function handleDeleteAllThemes() {
		if (appStore.state.savedThemes.length === 0) return;

		const shouldDeleteAll = await dialogStore.confirm({
			title: 'Delete all saved themes?',
			message: `This will permanently delete ${appStore.state.savedThemes.length} saved theme${appStore.state.savedThemes.length === 1 ? '' : 's'}.`,
			confirmLabel: 'Delete all',
			variant: 'danger'
		});
		if (!shouldDeleteAll) return;

		const themeIds = appStore.state.savedThemes.map((item) => item.id);
		await appStore.deleteThemes(themeIds);
	}

	function getPreviewColors(item: SavedThemeItem): string[] {
		return item.themeResult.colors.slice(0, 6).map((color) => color.hex);
	}

	async function handleThemeShareToggle(item: SavedThemeItem) {
		if (isLocalId(item.id)) {
			toast.error('Sign in first to share themes.');
			return;
		}

		await appStore.setThemeShared(item.id, !item.isShared);
	}
</script>

<div
	class={cn(
		'palette-dropdown-base w-80 max-w-[90vw]',
		popoverStore.state.direction === 'right' ? 'left-full ml-2' : 'right-full mr-2'
	)}
	style={`min-width: 260px; ${popoverStore.state.direction === 'right' ? 'left: calc(100% + 0.5rem);' : 'right: calc(100% + 0.5rem);'}`}
	role="dialog"
	aria-labelledby="saved-themes-title"
	tabindex="-1"
>
	<div class="mb-3 flex items-center justify-between">
		<h3 id="saved-themes-title" class="text-brand text-xs font-medium">Saved Themes</h3>

		<DangerTextButton
			onclick={handleDeleteAllThemes}
			title="Delete all saved themes"
			disabled={appStore.state.savedThemes.length === 0}
		>
			Delete all
		</DangerTextButton>
	</div>
	<div class="scrollable-content custom-scrollbar max-h-72 overflow-y-auto">
		{#if appStore.state.savedThemes.length === 0}
			<EmptyState>
				<svg class="mb-3 h-12 w-12 text-zinc-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="1.5"
						d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17h.01"
					/>
				</svg>
				<p class="text-sm text-zinc-400">No saved themes yet</p>
				<p class="mt-1 text-xs text-zinc-500">Copy a theme with save enabled to see it here</p>
			</EmptyState>
		{:else}
			<ul class="flex flex-col gap-3">
				{#each appStore.state.savedThemes as item (item.id)}
					<li
						class="hover:border-brand/50 group relative overflow-hidden rounded-lg border border-zinc-600 bg-zinc-800/50 transition-[background-color,border-color,box-shadow] duration-300 hover:bg-white/5"
					>
						<div class="p-3">
							<h4 class="text-brand mb-2 truncate font-mono text-sm font-semibold" title={item.name}>
								{item.name}
							</h4>
							<div class={cn('mb-3 flex items-center gap-1', isDesktopApp ? 'justify-around' : 'justify-end')}>
								<ActionPillButton
									onclick={() => handleSavedThemeLoad(item)}
									class="gap-1 px-2"
									title="Load into inspector"
								>
									<svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
									</svg>
									Load
								</ActionPillButton>
								<ActionPillButton
									onclick={() => handleThemeShareToggle(item)}
									class="px-2"
									title={isLocalId(item.id)
										? 'Sign in synced theme required to share'
										: item.isShared
											? 'Remove from shared list'
											: 'Share theme publicly'}
								>
									{item.isShared ? 'Unshare' : 'Share'}
								</ActionPillButton>
								{#if isDesktopApp}
									<ActionPillButton
										onclick={() => handleThemeSave(item)}
										class="gap-1 px-2"
										title="Save to editor folder"
									>
										<svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M12 16v-8m0 8l3-3m-3 3l-3-3M4 16v2a2 2 0 002 2h12a2 2 0 002-2v-2"
											></path>
										</svg>
										Save
									</ActionPillButton>
								{/if}
								<IconDangerButton
									onclick={() => handleThemeDelete(item.id, item.name)}
									title="Delete theme"
									class="shrink-0 p-1"
								>
									<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
										/>
									</svg>
								</IconDangerButton>
							</div>
							<div class="flex flex-wrap gap-1.5">
								{#each getPreviewColors(item) as color (color)}
									<div class="group/swatch relative">
										<ColorSwatch {color} title={color} />
									</div>
								{/each}
								{#if getPreviewColors(item).length === 0}
									<span class="text-xs text-zinc-500">No preview colors</span>
								{/if}
							</div>
						</div>

						<div class="group-hover:border-brand/50 group border-t border-zinc-600 bg-zinc-900/50 px-3 py-1.5">
							<div class="flex items-center justify-between gap-2">
								<span class="flex items-center gap-1.5 text-xs text-zinc-500">
									<svg class="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
										/>
									</svg>
									{new Date(item.createdAt).toLocaleDateString('en-US', {
										month: 'short',
										day: 'numeric',
										year: 'numeric'
									})}
								</span>
								{#if item.isShared}
									<span
										class="rounded-full border border-emerald-400/40 bg-emerald-500/15 px-2 py-0.5 text-[10px] font-semibold text-emerald-300"
									>
										Shared
									</span>
								{/if}
							</div>
						</div>
					</li>
				{/each}
			</ul>
		{/if}
	</div>
</div>
