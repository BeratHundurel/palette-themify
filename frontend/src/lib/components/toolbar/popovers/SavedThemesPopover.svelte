<script lang="ts">
	import { cn } from '$lib/utils';
	import { appStore } from '$lib/stores/app.svelte';
	import { popoverStore } from '$lib/stores/popovers.svelte';
	import { extractThemeColorsWithUsage } from '$lib/colorUtils';
	import toast from 'svelte-french-toast';
	import type { SavedThemeItem } from '$lib/types/theme';
	import { generateOverridable } from '$lib/api/theme';

	let isImportOpen = $state(false);
	let importJson = $state('');
	let importFileInput: HTMLInputElement | null = $state(null);

	async function handleThemeCopy(item: SavedThemeItem) {
		try {
			const themeJson = JSON.stringify(item.themeResult.theme, null, 2);
			await navigator.clipboard.writeText(themeJson);
			toast.success('Theme JSON copied to clipboard!');
			popoverStore.close('themes');
		} catch {
			toast.error('Could not copy the theme. Please try again.');
		}
	}

	async function handleThemeLoad(theme: unknown) {
		try {
			const response = await generateOverridable(theme);
			appStore.setThemeExportEditorType('zed');
			appStore.state.themeExport.themeName = response.theme.name;
			appStore.state.themeExport.themeResult = response;
			appStore.state.themeExport.themeColorsWithUsage = extractThemeColorsWithUsage(response.theme);
			appStore.state.themeExport.lastGeneratedPaletteVersion = appStore.state.paletteVersion;
			importJson = '';
			popoverStore.close('themes');
			popoverStore.state.current = 'themeExport';
		} catch {
			toast.error('Could not load the theme. Please check your JSON.');
		}
	}

	async function handleImportText(value: string) {
		if (!value.trim()) {
			toast.error('Paste a theme JSON first.');
			return;
		}
		try {
			const parsed = JSON.parse(value);
			await handleThemeLoad(parsed);
		} catch {
			toast.error('Invalid JSON. Please check your input.');
		}
	}

	async function handleImportFile(event: Event) {
		const input = event.target as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;
		try {
			const content = await file.text();
			const parsed = JSON.parse(content);
			await handleThemeLoad(parsed);
		} catch {
			toast.error('Could not read the JSON file.');
		} finally {
			input.value = '';
		}
	}

	function triggerFileImport() {
		if (!importFileInput) return;
		importFileInput.click();
	}

	// function handleThemeLoad(item: SavedThemeItem) {
	// 	appStore.setThemeExportEditorType(item.editorType);
	// 	appStore.state.themeExport.themeName = item.name;
	// 	appStore.state.themeExport.themeResult = item.themeResult;
	// 	appStore.state.themeExport.themeColorsWithUsage = item.themeColorsWithUsage;
	// 	appStore.state.themeExport.lastGeneratedPaletteVersion = appStore.state.paletteVersion;
	// 	popoverStore.close('themes');
	// 	popoverStore.state.current = 'themeExport';
	// }

	async function handleThemeDelete(themeId: string, themeName: string) {
		if (confirm(`Are you sure you want to delete "${themeName}"?`)) {
			appStore.deleteTheme(themeId);
		}
	}

	function getPreviewColors(item: SavedThemeItem): string[] {
		const sorted = item.themeColorsWithUsage
			? [...item.themeColorsWithUsage].sort((a, b) => b.totalUsages - a.totalUsages)
			: [];
		return sorted.slice(0, 6).map((entry) => entry.baseColor);
	}
</script>

<div
	class={cn(
		'palette-dropdown-base w-80',
		popoverStore.state.direction === 'right' ? 'left-full ml-2' : 'right-full mr-2'
	)}
	style={`min-width: 260px; ${popoverStore.state.direction === 'right' ? 'left: calc(100% + 0.5rem);' : 'right: calc(100% + 0.5rem);'}`}
	role="dialog"
	aria-labelledby="saved-themes-title"
	tabindex="-1"
>
	<div class="mb-3 flex items-center justify-between p-2">
		<h3 id="saved-themes-title" class="text-brand text-xs font-medium">Saved Themes</h3>
		<button
			type="button"
			title="Import theme"
			onclick={() => (isImportOpen = !isImportOpen)}
			class="text-brand hover:bg-brand/10 flex cursor-pointer items-center gap-1 rounded-md px-2.5 text-xs font-medium transition-[transform,background-color] hover:scale-105"
		>
			Import a theme</button
		>
	</div>
	{#if isImportOpen}
		<div class="mb-3 rounded-lg border border-zinc-700/60 bg-zinc-900/60 p-3">
			<div class="mb-2 flex items-center justify-between">
				<div class="text-xs font-semibold text-zinc-300">Import Zed theme JSON</div>
				<button
					type="button"
					onclick={() => (isImportOpen = !isImportOpen)}
					class="rounded-md px-2 py-1 text-xs text-zinc-500 transition-[background-color,color] hover:bg-zinc-800/60 hover:text-zinc-200"
				>
					Close
				</button>
			</div>
			<textarea
				rows="6"
				bind:value={importJson}
				placeholder="Paste your Zed theme JSON"
				class="focus:border-brand/50 w-full rounded-md border border-zinc-700 bg-zinc-950/60 p-2 font-mono text-xs text-zinc-200 placeholder-zinc-600 transition-[border-color,box-shadow,background-color] duration-200 focus:outline-none"
			></textarea>
			<div class="mt-2 flex items-center justify-between gap-2">
				<div class="flex items-center gap-2">
					<input
						type="file"
						accept="application/json,.json"
						bind:this={importFileInput}
						onchange={handleImportFile}
						class="hidden"
					/>
					<button
						type="button"
						onclick={triggerFileImport}
						class="hover:border-brand/50 rounded-md border border-zinc-700 px-3 py-1.5 text-xs font-medium text-zinc-300 transition-[background-color,border-color] duration-200 hover:bg-zinc-800/60"
					>
						Choose JSON file
					</button>
				</div>
				<button
					type="button"
					onclick={() => handleImportText(importJson)}
					class="bg-brand/90 hover:bg-brand rounded-md px-3 py-1.5 text-xs font-semibold text-zinc-900 transition-[transform,background-color] hover:scale-105"
				>
					Import theme
				</button>
			</div>
		</div>
	{:else}
		<div class="scrollable-content custom-scrollbar max-h-72 overflow-y-auto">
			{#if appStore.state.savedThemes.length === 0}
				<div class="flex flex-col items-center justify-center py-12 text-center">
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
				</div>
			{:else}
				<ul class="flex flex-col gap-3">
					{#each appStore.state.savedThemes as item (item.id)}
						<li
							class="hover:border-brand/50 group relative overflow-hidden rounded-lg border border-zinc-600 bg-zinc-800/50 transition-[background-color,border-color,box-shadow] duration-300 hover:bg-white/5"
						>
							<div class="p-3">
								<h4 class="text-brand truncate font-mono text-sm font-semibold" title={item.name}>
									{item.name}
								</h4>
								<div class="mb-2 flex items-center justify-end gap-1">
									<button
										class="text-brand hover:bg-brand/10 flex cursor-pointer items-center gap-1 rounded-md px-2.5 text-xs font-medium transition-[transform,background-color] hover:scale-105"
										onclick={() => handleThemeLoad(item.themeResult.theme)}
										type="button"
										title="Load into inspector"
									>
										<svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
										</svg>
										Load
									</button>
									<button
										class="text-brand hover:bg-brand/10 flex cursor-pointer items-center gap-1 rounded-md px-2.5 text-xs font-medium transition-[transform,background-color] hover:scale-105"
										onclick={() => handleThemeCopy(item)}
										type="button"
										title="Copy theme"
									>
										<svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
											<path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
										</svg>
										Copy
									</button>
									<button
										class="flex cursor-pointer items-center gap-1 rounded-md p-1.5 text-zinc-500 transition-[transform,background-color,color] hover:scale-110 hover:bg-red-500/10 hover:text-red-400"
										onclick={() => handleThemeDelete(item.id, item.name)}
										type="button"
										title="Delete theme"
									>
										<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
											/>
										</svg>
									</button>
								</div>
								<div class="flex flex-wrap gap-1.5">
									{#each getPreviewColors(item) as color (color)}
										<div class="group/swatch relative">
											<span
												class="inline-block h-8 w-8 cursor-pointer rounded-md border border-zinc-700/50 shadow-md transition-[transform,box-shadow,border-color] duration-300 hover:scale-105 hover:ring-2 hover:ring-white/50"
												style="background-color: {color}"
												title={color}
											></span>
										</div>
									{/each}
									{#if getPreviewColors(item).length === 0}
										<span class="text-xs text-zinc-500">No preview colors</span>
									{/if}
								</div>
							</div>

							<div class="group-hover:border-brand/50 group border-t border-zinc-600 bg-zinc-900/50 px-3 py-1.5">
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
							</div>
						</li>
					{/each}
				</ul>
			{/if}
		</div>
	{/if}
</div>
