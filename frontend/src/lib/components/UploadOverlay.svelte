<script lang="ts">
	import { generateOverridable } from '$lib/api/theme';
	import { detectThemeAppearance, detectThemeType } from '$lib/colorUtils';
	import { hydrateThemeExportResponse } from '$lib/components/toolbar/theme-export/session';
	import { popoverStore } from '$lib/stores/popovers.svelte';
	import { appStore } from '$lib/stores/app/store.svelte';
	import type { EditorThemeType } from '$lib/types/themeApi';
	import type { Theme } from '$lib/types/theme';
	import { cn } from '$lib/utils';
	import toast from 'svelte-french-toast';

	let dragDepth = $state(0);
	let isDragOver = $state(false);
	let importMode = $state<'image' | 'theme'>('image');
	let importJson = $state('');

	let importFileInput: HTMLInputElement | null = $state(null);
	const activeModeTransform = $derived(importMode === 'theme' ? 'calc(100% + 0.5rem)' : '0px');

	function setImportMode(mode: 'image' | 'theme') {
		importMode = mode;
		dragDepth = 0;
		isDragOver = false;
	}

	function handleDragEnter() {
		if (importMode !== 'image') return;
		dragDepth += 1;
		isDragOver = true;
	}

	function handleDragLeave() {
		if (importMode !== 'image') return;
		dragDepth = Math.max(0, dragDepth - 1);
		if (dragDepth === 0) {
			isDragOver = false;
		}
	}

	function handleDrop(event: DragEvent) {
		event.preventDefault();
		event.stopPropagation();
		if (importMode !== 'image') return;
		dragDepth = 0;
		isDragOver = false;
		void appStore.handleDrop(event);
	}

	function handleDragOver(event: DragEvent) {
		event.preventDefault();
		event.stopPropagation();
	}

	function handleDragEnterEvent(event: DragEvent) {
		event.preventDefault();
		event.stopPropagation();
		handleDragEnter();
	}

	function handleDragLeaveEvent(event: DragEvent) {
		event.preventDefault();
		event.stopPropagation();
		handleDragLeave();
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

			importJson = '';
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

	function triggerThemeFileImport() {
		if (!importFileInput) return;
		importFileInput.click();
	}
</script>

<section
	class={cn(
		'absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2',
		appStore.state.imageLoaded ? 'pointer-events-none hidden' : 'pointer-events-auto'
	)}
	aria-hidden={appStore.state.imageLoaded ? 'true' : 'false'}
>
	<div class="mb-6 flex items-center justify-center">
		<div
			role="group"
			aria-label="Import mode"
			class="relative inline-flex w-60 gap-2 rounded-xl border border-white/40 bg-white/10 p-1 shadow-md backdrop-blur-sm"
		>
			<div
				class="bg-brand pointer-events-none absolute inset-y-1 left-1 rounded-lg shadow-sm transition-[transform,width] duration-300 ease-out"
				style="width: calc((100% - 1rem) / 2); transform: translateX({activeModeTransform});"
			></div>
			<button
				type="button"
				onclick={() => setImportMode('image')}
				class={cn(
					'relative z-10 h-9 flex-1 rounded-lg px-4 text-sm font-medium whitespace-nowrap transition-colors duration-300',
					importMode === 'image' ? 'text-zinc-900' : 'text-zinc-100 hover:text-white'
				)}
			>
				Image
			</button>
			<button
				type="button"
				onclick={() => setImportMode('theme')}
				class={cn(
					'relative z-10 h-9 flex-1 rounded-lg px-4 text-sm font-medium whitespace-nowrap transition-colors duration-300',
					importMode === 'theme' ? 'text-zinc-900' : 'text-zinc-100 hover:text-white'
				)}
			>
				Theme JSON
			</button>
		</div>
	</div>

	<div class="flex min-h-68 w-md items-start justify-center">
		{#if importMode === 'image'}
			<button
				type="button"
				ondrop={handleDrop}
				ondragover={handleDragOver}
				ondragenter={handleDragEnterEvent}
				ondragleave={handleDragLeaveEvent}
				onclick={appStore.triggerFileSelect}
				class={cn(
					'group flex h-56 w-full cursor-pointer flex-col items-center justify-center rounded-2xl border-2 border-dashed transition-[background-color,border-color] duration-300',
					isDragOver ? 'border-white bg-white/28' : 'border-white/60 bg-white/12 hover:border-white hover:bg-white/24'
				)}
				aria-label="Upload an image or drag and drop it here"
			>
				<svg
					viewBox="0 0 24 24"
					width="3.25rem"
					height="3.25rem"
					fill="#fff"
					role="img"
					aria-label="Upload icon"
					class="mb-3 opacity-80 transition-opacity group-hover:opacity-100"
				>
					<path
						d="M19 7v3h-2V7h-3V5h3V2h2v3h3v2h-3zm-3 4V8h-3V5H5a2 2 0 00-2 2v12c0 1.1.9 2 2 2h12a2 2 0 002-2v-8h-3zM5 19l3-4 2 3 3-4 4 5H5z"
					/>
				</svg>
				<span class="text-sm text-zinc-300 group-hover:text-white">Upload an image or drag and drop it here</span>
				<input
					type="file"
					name="file"
					bind:this={appStore.state.fileInput}
					class="hidden"
					accept="image/*"
					oninput={appStore.onFileChange}
					aria-label="Choose image to extract a palette from"
				/>
			</button>
		{:else}
			<div class="w-full rounded-2xl border border-white/40 bg-white/10 p-4 shadow-lg backdrop-blur-sm">
				<div class="mb-3 flex items-start justify-between gap-4">
					<div>
						<p class="text-sm font-semibold text-white">Import Theme JSON</p>
						<p class="mt-0.5 text-xs text-zinc-200/85">Paste JSON or upload a `.json` file</p>
					</div>
					<span
						class="rounded-full border border-white/35 bg-white/12 px-2.5 py-1 text-[10px] font-medium tracking-wide text-zinc-200"
					>
						Theme
					</span>
				</div>

				<textarea
					rows="6"
					bind:value={importJson}
					placeholder={`VSCode:\n{\n  "$schema": "vscode://schemas/color-theme",\n  "name": "My Theme",\n  "type": "dark",\n  "colors": {\n    "editor.background": "#0f1115",\n    "editor.foreground": "#d6deeb"\n  },\n  "tokenColors": []\n}\n\nZed:\n{\n  "$schema": "https://zed.dev/schema/themes/v0.2.0.json",\n  "name": "My Theme",\n  "author": "You",\n  "themes": [{\n    "name": "My Theme",\n    "appearance": "dark",\n    "style": {}\n  }]\n}`}
					class="custom-scrollbar h-34 w-full resize-none rounded-lg bg-zinc-950/65 p-3 font-mono text-xs leading-5 text-zinc-100 placeholder-zinc-400/80 transition-[background-color] duration-200 focus:bg-zinc-950/80 focus:outline-none"
				></textarea>

				<div class="mt-3.5 flex items-center justify-between gap-2">
					<input
						type="file"
						accept="application/json,.json"
						bind:this={importFileInput}
						onchange={handleImportFile}
						class="hidden"
					/>
					<button
						type="button"
						onclick={triggerThemeFileImport}
						class="rounded-lg border border-white/45 bg-white/10 px-3.5 py-2 text-xs font-medium text-zinc-100 transition-[background-color,border-color,color] hover:border-white/70 hover:bg-white/18"
					>
						Choose file
					</button>
					<button
						type="button"
						onclick={() => handleImportText(importJson)}
						class="bg-brand hover:bg-brand/90 rounded-lg px-4 py-2 text-xs font-semibold text-zinc-900 transition-[background-color,transform]"
					>
						Generate Editable Theme
					</button>
				</div>
			</div>
		{/if}
	</div>
</section>
