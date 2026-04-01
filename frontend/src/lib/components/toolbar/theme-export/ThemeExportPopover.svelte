<script lang="ts">
	import { popoverStore } from '$lib/stores/popovers.svelte';
	import { appStore } from '$lib/stores/app.svelte';
	import { generateOverridable, generateTheme, type EditorThemeType, type ThemeAppearance } from '$lib/api/theme';
	import { detectThemeAppearance, detectThemeType } from '$lib/colorUtils';
	import type { Theme, ThemeGenerationResponse, ThemeOverrides } from '$lib/types/theme';
	import toast from 'svelte-french-toast';
	import {
		buildRequestOverrides,
		clearThemeVersions,
		getCachedThemeVersion,
		getCurrentOverrides,
		resetThemeExportOverrideState,
		setActiveThemeResponse,
		setManualOverrideFlag
	} from './session';
	import { normalizeHex, overrideFields, THEME_NAME_DEBOUNCE_MS, validateThemeName } from './utils';
	import { exportTheme as exportThemeToClipboard } from './save';

	const SHUFFLE_KEYS: Array<keyof ThemeOverrides> = ['c1', 'c2', 'c3', 'c4', 'c5', 'c6', 'c7', 'c8', 'c9'];

	let themeNameDebounceTimer: ReturnType<typeof setTimeout> | null = null;

	let paletteVersion = $derived(appStore.state.paletteVersion);
	let themeName = $derived(appStore.state.themeExport.themeName);
	let editorType = $derived(appStore.state.themeExport.editorType);
	let themeAppearance = $derived(appStore.state.themeExport.appearance);
	let themeResult = $derived(appStore.state.themeExport.themeResult);
	let themeOverrides = $derived(appStore.state.themeExport.themeResult?.themeOverrides ?? {});
	let saveOnCopy = $derived(appStore.state.themeExport.saveOnCopy);
	let accentBoostCoefficient = $derived(appStore.state.themeExport.boostCoefficient);
	let accentBoostInput = $state(appStore.state.themeExport.boostCoefficient.toFixed(2));

	const baseOverrides = $derived(themeResult?.themeOverrides ?? {});

	let isOpen = $derived(popoverStore.isOpen('themeExport'));
	let themeNameError = $state<string | null>(null);
	let isGenerating = $state(false);

	$effect(() => {
		if (
			isOpen &&
			!isGenerating &&
			themeResult !== null &&
			appStore.state.colors.length > 0 &&
			appStore.state.themeExport.lastGeneratedPaletteVersion === 0
		) {
			themeName = 'Generated Theme';
			resetThemeExportOverrideState();
			generateThemeFromApi({ overrides: {}, bypassCache: true });
		}
	});

	$effect(() => {
		if (isOpen && themeResult === null && !isGenerating) {
			themeName = 'Generated Theme';
			resetThemeExportOverrideState();
			generateThemeFromApi({ overrides: {}, bypassCache: true });
		}
	});

	$effect(() => {
		if (
			!isOpen ||
			isGenerating ||
			appStore.state.themeExport.lastGeneratedPaletteVersion == 0 ||
			appStore.state.themeExport.lastGeneratedPaletteVersion == paletteVersion
		)
			return;

		themeName = 'Generated Theme';
		resetThemeExportOverrideState();
		generateThemeFromApi({ overrides: {}, bypassCache: true });
	});

	function handleThemeNameChange(e: Event) {
		const target = e.target as HTMLInputElement;
		const newName = target.value;

		appStore.state.themeExport.themeName = newName;
		themeNameError = validateThemeName(newName);

		if (themeNameDebounceTimer) {
			clearTimeout(themeNameDebounceTimer);
		}

		themeNameDebounceTimer = setTimeout(() => {
			updateThemeNameIfValid(newName);
		}, THEME_NAME_DEBOUNCE_MS);
	}

	function updateThemeNameIfValid(name: string) {
		const trimmedName = name.trim();

		if (!themeResult || themeNameError || trimmedName.length === 0) {
			return;
		}

		try {
			updateThemeNameInThemeResult(trimmedName);
		} catch {
			toast.error('Could not update the theme name. Regenerating the theme...');
			generateThemeFromApi();
		}
	}

	function updateThemeNameInTheme(theme: Theme, name: string) {
		try {
			if (editorType === 'zed' && 'themes' in theme) {
				theme.name = name;
				if (theme.themes && Array.isArray(theme.themes) && theme.themes[0]) {
					theme.themes[0].name = name;
				}
			} else if (editorType === 'vscode' && 'colors' in theme) {
				theme.name = name;
			}
		} catch {
			throw new Error('Failed to update theme name');
		}
	}

	function updateThemeNameInThemeResult(name: string) {
		if (!themeResult) return;

		updateThemeNameInTheme(themeResult.theme, name);

		for (const response of Object.values(appStore.state.themeExport.themeVersions) as ThemeGenerationResponse[]) {
			updateThemeNameInTheme(response.theme, name);
		}
	}

	async function generateThemeFromApi(options?: {
		type?: EditorThemeType;
		appearance?: ThemeAppearance;
		overrides?: ThemeOverrides;
		bypassCache?: boolean;
		accentBoostCoefficient?: number;
	}) {
		if (isGenerating) return;

		const validationError = validateThemeName(themeName);
		if (validationError) {
			themeNameError = validationError;
			return;
		}

		const hasPaletteColors = appStore.state.colors && appStore.state.colors.length > 0;
		const nextType = options?.type ?? editorType;
		const nextAppearance = options?.appearance ?? themeAppearance;
		const requestOverrides = options?.overrides ?? buildRequestOverrides(nextAppearance, themeAppearance);
		const requestAccentBoostCoefficient = options?.accentBoostCoefficient ?? accentBoostCoefficient;

		if (!options?.bypassCache) {
			const cached = getCachedThemeVersion(nextType, nextAppearance);
			if (cached) {
				setActiveThemeResponse(cached, nextType, nextAppearance);
				return;
			}
		}

		try {
			isGenerating = true;
			let response: ThemeGenerationResponse;

			if (hasPaletteColors) {
				response = await generateTheme(
					appStore.state.colors,
					nextType,
					themeName.trim(),
					requestOverrides,
					nextAppearance,
					requestAccentBoostCoefficient
				);
				appStore.state.themeExport.loadedThemeOverridesReference = null;
				appStore.state.themeExport.lastGeneratedPaletteVersion = appStore.state.paletteVersion;
			} else {
				if (!themeResult?.theme) {
					return;
				}

				const hasBackupColors =
					appStore.state.themeExport.backupColors && appStore.state.themeExport.backupColors.length > 0;

				if (
					(nextType !== detectThemeType(themeResult.theme) ||
						nextAppearance !== detectThemeAppearance(themeResult.theme)) &&
					hasBackupColors
				) {
					response = await generateTheme(
						appStore.state.themeExport.backupColors!,
						nextType,
						themeName.trim(),
						requestOverrides,
						nextAppearance,
						requestAccentBoostCoefficient
					);
				} else {
					response = await generateOverridable(
						themeResult.theme,
						requestOverrides,
						nextType,
						nextAppearance,
						requestAccentBoostCoefficient
					);
				}
			}

			setActiveThemeResponse(response, nextType, nextAppearance);
		} catch {
			toast.error('Could not generate the theme. Please try again.');
			appStore.state.themeExport.themeResult = null;
			clearThemeVersions();
		} finally {
			isGenerating = false;
		}
	}

	function resetOverrides() {
		const overrides =
			appStore.state.themeExport.loadedThemeOverridesReference && appStore.state.colors.length === 0
				? { ...appStore.state.themeExport.loadedThemeOverridesReference }
				: {};

		resetThemeExportOverrideState(overrides);

		generateThemeFromApi({ overrides, bypassCache: true });
	}

	function shuffleThemeDistribution() {
		const currentOverrides = getCurrentOverrides();
		const source = SHUFFLE_KEYS.map((key) => currentOverrides[key] ?? baseOverrides[key]).filter(
			(value): value is string => Boolean(value)
		);

		if (source.length < SHUFFLE_KEYS.length) {
			toast.error('Not enough generated theme colors to shuffle right now.');
			return;
		}

		const shuffled = [...source];
		for (let i = shuffled.length - 1; i > 0; i--) {
			const j = Math.floor(Math.random() * (i + 1));
			[shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
		}

		const sameOrder = shuffled.every((value, index) => value === source[index]);
		if (sameOrder && shuffled.length > 1) {
			[shuffled[0], shuffled[1]] = [shuffled[1], shuffled[0]];
		}

		clearThemeVersions();

		const overrides = getCurrentOverrides();
		for (const [index, key] of SHUFFLE_KEYS.entries()) {
			overrides[key] = shuffled[index];
		}

		appStore.state.themeExport.rawThemeOverrides = { ...overrides };

		generateThemeFromApi({ overrides, bypassCache: true });
	}

	function getOverrideValue(key: keyof ThemeOverrides): string {
		return themeOverrides[key] ?? baseOverrides[key] ?? '#000000';
	}

	function updateThemeOverride(key: keyof ThemeOverrides, value: string) {
		const normalized = normalizeHex(value);
		if (!normalized) return;

		clearThemeVersions();

		const overrides = getCurrentOverrides();
		overrides[key] = normalized;
		appStore.state.themeExport.rawThemeOverrides = { ...overrides };
		setManualOverrideFlag(key, true);

		generateThemeFromApi({ overrides, bypassCache: true });
	}

	async function handleEditorTypeChange(type: EditorThemeType) {
		if (editorType === type) return;

		if (themeNameDebounceTimer) {
			clearTimeout(themeNameDebounceTimer);
			themeNameDebounceTimer = null;
		}

		appStore.setThemeExportEditorType(type);

		const cached = getCachedThemeVersion(type, themeAppearance);
		if (cached) {
			setActiveThemeResponse(cached, type, themeAppearance);
			return;
		}

		generateThemeFromApi({
			type,
			appearance: themeAppearance,
			overrides: buildRequestOverrides(themeAppearance, themeAppearance)
		});
	}

	async function handleThemeAppearanceChange(appearance: ThemeAppearance) {
		if (themeAppearance === appearance) return;

		const cached = getCachedThemeVersion(editorType, appearance);
		const overrides = buildRequestOverrides(appearance, themeAppearance);
		appStore.setThemeExportAppearance(appearance);
		if (cached) {
			setActiveThemeResponse(cached, editorType, appearance);
			return;
		}

		generateThemeFromApi({
			type: editorType,
			appearance,
			overrides
		});
	}

	function handleAccentBoostInput(event: Event) {
		const target = event.target as HTMLInputElement;
		accentBoostInput = target.value;

		const parsed = Number(target.value);
		if (Number.isNaN(parsed)) return;

		const clamped = Math.min(3, Math.max(0, parsed));
		const normalized = Math.round(clamped * 100) / 100;

		appStore.setThemeExportBoostCoefficient(normalized);
		accentBoostInput = normalized.toFixed(2);

		generateThemeFromApi({ bypassCache: true, accentBoostCoefficient: normalized });
	}

	async function handleExportTheme() {
		await exportThemeToClipboard({
			name: themeName,
			editorType,
			themeResult,
			saveOnCopy
		});
	}
</script>

{#if popoverStore.isOpen('themeExport')}
	<div class="fixed inset-0 z-50 flex items-center justify-center p-4">
		<div class="animate-fade-in absolute inset-0 bg-black/70" aria-hidden="true"></div>

		<div
			role="dialog"
			aria-labelledby="theme-inspector-title"
			aria-modal="true"
			class="share-modal-content border-brand/50 shadow-brand/20 animate-scale-in relative flex max-h-[95vh] w-full max-w-6xl flex-col overflow-hidden rounded-xl border bg-zinc-900 shadow-2xl"
		>
			<div class="flex items-center justify-between border-b border-zinc-700 bg-zinc-800/50 px-6 py-5">
				<div>
					<h2 id="theme-inspector-title" class="text-brand text-2xl font-semibold">
						{editorType === 'vscode' ? 'VS Code' : 'Zed'} Theme Inspector
					</h2>
					<p class="mt-1 text-sm text-zinc-400">Review the generated theme colors before exporting</p>
				</div>
				<button
					type="button"
					onclick={() => popoverStore.close('themeExport')}
					class="hover:text-brand rounded-lg p-2 text-zinc-400 transition-[background-color,color] duration-300 hover:bg-zinc-800/50"
					aria-label="Close"
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						width="24"
						height="24"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
					>
						<line x1="18" y1="6" x2="6" y2="18"></line>
						<line x1="6" y1="6" x2="18" y2="18"></line>
					</svg>
				</button>
			</div>

			<div class="custom-scrollbar flex-1 overflow-y-auto px-6 py-6">
				<div class="mb-8">
					<div class="mb-6 flex items-center gap-2">
						<h3 class="text-brand text-sm font-semibold tracking-wide uppercase">Theme Name</h3>
						<div class="from-brand/50 h-px flex-1 bg-linear-to-r to-transparent"></div>
					</div>
					<div class="space-y-3">
						<div>
							<label for="theme-name" class="mb-2 block text-sm font-medium text-zinc-300">
								Enter a custom name for your theme
							</label>
							<input
								id="theme-name"
								type="text"
								value={themeName}
								oninput={handleThemeNameChange}
								placeholder="Generated Theme"
								class="focus:border-brand/50 w-full rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-300 placeholder-zinc-400 transition-[border-color,box-shadow,background-color] duration-300 focus:outline-none"
								class:border-red-500={themeNameError !== null}
							/>
							{#if themeNameError}
								<p class="mt-1.5 text-xs text-red-400">{themeNameError}</p>
							{/if}
						</div>
					</div>
				</div>

				<div class="mb-8">
					<div class="mb-6 flex items-center gap-2">
						<h3 class="text-brand text-sm font-semibold tracking-wide uppercase">Theme Appearance</h3>
						<div class="from-brand/50 h-px flex-1 bg-linear-to-r to-transparent"></div>
					</div>
					<div class="grid gap-4 md:grid-cols-2">
						<button
							type="button"
							onclick={() => handleThemeAppearanceChange('dark')}
							class="group relative overflow-hidden rounded-xl border px-4 py-4 text-left transition-[background-color,border-color,box-shadow] duration-300 {themeAppearance ===
							'dark'
								? 'border-brand/70 shadow-brand/20 bg-zinc-900/80 shadow-lg'
								: 'hover:border-brand/50 border-zinc-700 bg-zinc-950/60 hover:bg-zinc-900/60'}"
						>
							<div class="flex items-start justify-between gap-3">
								<div>
									<div class="text-sm font-semibold {themeAppearance === 'dark' ? 'text-brand' : 'text-zinc-200'}">
										Dark
									</div>
									<div class="mt-1 text-xs text-zinc-400">Deep surfaces, bold accents</div>
								</div>
								<div class="rounded-lg border border-zinc-700 bg-zinc-950/80 p-2">
									<div class="grid grid-cols-3 gap-1">
										<div class="h-3 w-5 rounded bg-zinc-800"></div>
										<div class="h-3 w-5 rounded bg-zinc-700"></div>
										<div class="h-3 w-5 rounded bg-zinc-600"></div>
										<div class="h-3 w-5 rounded bg-zinc-900"></div>
										<div class="h-3 w-5 rounded bg-zinc-800"></div>
										<div class="h-3 w-5 rounded bg-zinc-700"></div>
										<div class="h-3 w-5 rounded bg-zinc-950"></div>
										<div class="h-3 w-5 rounded bg-zinc-900"></div>
										<div class="h-3 w-5 rounded bg-zinc-800"></div>
									</div>
								</div>
							</div>
							<div class="mt-3 flex items-center gap-2">
								<div
									class="flex h-5 w-5 items-center justify-center rounded-full border transition-[background-color,border-color] {themeAppearance ===
									'dark'
										? 'border-brand bg-brand/20'
										: 'border-zinc-500'}"
								>
									{#if themeAppearance === 'dark'}
										<div class="bg-brand h-2.5 w-2.5 rounded-full"></div>
									{/if}
								</div>
								<span class="text-xs font-medium text-zinc-400">Optimized for low-light setups</span>
							</div>
						</button>

						<button
							type="button"
							onclick={() => handleThemeAppearanceChange('light')}
							class="group relative overflow-hidden rounded-xl border px-4 py-4 text-left transition-[background-color,border-color,box-shadow] duration-300 {themeAppearance ===
							'light'
								? 'border-brand/40 shadow-brand/10 bg-zinc-800/40 shadow-lg'
								: 'hover:border-brand/40 border-zinc-700 bg-zinc-950/60 hover:bg-zinc-900/40'}"
						>
							<div class="flex items-start justify-between gap-3">
								<div>
									<div class="text-sm font-semibold {themeAppearance === 'light' ? 'text-brand' : 'text-zinc-200'}">
										Light
									</div>
									<div class="mt-1 text-xs text-zinc-400">Airy surfaces, crisp contrast</div>
								</div>
								<div class="rounded-lg border border-zinc-700 bg-zinc-900/60 p-2">
									<div class="grid grid-cols-3 gap-1">
										<div class="h-3 w-5 rounded bg-zinc-200"></div>
										<div class="h-3 w-5 rounded bg-zinc-300"></div>
										<div class="h-3 w-5 rounded bg-zinc-400"></div>
										<div class="h-3 w-5 rounded bg-zinc-100"></div>
										<div class="h-3 w-5 rounded bg-zinc-200"></div>
										<div class="h-3 w-5 rounded bg-zinc-300"></div>
										<div class="h-3 w-5 rounded bg-zinc-50"></div>
										<div class="h-3 w-5 rounded bg-zinc-100"></div>
										<div class="h-3 w-5 rounded bg-zinc-200"></div>
									</div>
								</div>
							</div>
							<div class="mt-3 flex items-center gap-2">
								<div
									class="flex h-5 w-5 items-center justify-center rounded-full border transition-[background-color,border-color] {themeAppearance ===
									'light'
										? 'border-brand bg-brand/15'
										: 'border-zinc-500'}"
								>
									{#if themeAppearance === 'light'}
										<div class="bg-brand h-2.5 w-2.5 rounded-full"></div>
									{/if}
								</div>
								<span class="text-xs font-medium text-zinc-400">Balanced for daylight work</span>
							</div>
						</button>
					</div>
				</div>

				<div class="mb-8">
					<div class="mb-6 flex items-center gap-2">
						<h3 class="text-brand text-sm font-semibold tracking-wide uppercase">Editor Type</h3>
						<div class="from-brand/50 h-px flex-1 bg-linear-to-r to-transparent"></div>
					</div>
					<div class="grid grid-cols-2 gap-6">
						<button
							type="button"
							onclick={() => handleEditorTypeChange('vscode')}
							class="group relative overflow-hidden rounded-lg border px-4 py-3 text-left transition-[background-color,border-color,box-shadow] duration-300 {editorType ===
							'vscode'
								? 'border-brand bg-brand/10 shadow-brand/20 shadow-lg'
								: 'hover:border-brand/50 border-zinc-600 hover:bg-zinc-800/50'}"
						>
							<div class="flex items-center gap-2">
								<div
									class="flex h-5 w-5 items-center justify-center rounded-full border transition-[background-color,border-color] {editorType ===
									'vscode'
										? 'border-brand bg-brand/20'
										: 'border-zinc-500'}"
								>
									{#if editorType === 'vscode'}
										<div class="bg-brand h-2.5 w-2.5 rounded-full"></div>
									{/if}
								</div>
								<span class="text-sm font-medium {editorType === 'vscode' ? 'text-brand' : 'text-zinc-200'}"
									>VS Code</span
								>
							</div>
							<p class="mt-1.5 ml-7 text-xs text-zinc-400">Generate theme for Visual Studio Code</p>
						</button>

						<button
							type="button"
							onclick={() => handleEditorTypeChange('zed')}
							class="group relative overflow-hidden rounded-lg border px-4 py-3 text-left transition-[background-color,border-color,box-shadow] duration-300 {editorType ===
							'zed'
								? 'border-brand bg-brand/10 shadow-brand/20 shadow-lg'
								: 'hover:border-brand/50 border-zinc-600 hover:bg-zinc-800/50'}"
						>
							<div class="flex items-center gap-2">
								<div
									class="flex h-5 w-5 items-center justify-center rounded-full border transition-[background-color,border-color] {editorType ===
									'zed'
										? 'border-brand bg-brand/20'
										: 'border-zinc-500'}"
								>
									{#if editorType === 'zed'}
										<div class="bg-brand h-2.5 w-2.5 rounded-full"></div>
									{/if}
								</div>
								<span class="text-sm font-medium {editorType === 'zed' ? 'text-brand' : 'text-zinc-200'}">Zed</span>
							</div>
							<p class="mt-1.5 ml-7 text-xs text-zinc-400">Generate theme for Zed editor</p>
						</button>
					</div>
				</div>

				<div class="mb-8">
					<div class="mb-4 flex items-center gap-2">
						<h3 class="text-brand text-sm font-semibold tracking-wide uppercase">Save Behavior</h3>
						<div class="from-brand/50 h-px flex-1 bg-linear-to-r to-transparent"></div>
					</div>
					<div class="space-y-3">
						<label
							class="flex cursor-pointer items-center justify-between rounded-lg border border-zinc-700/50 bg-zinc-900/60 px-4 py-3"
						>
							<div>
								<div class="text-sm font-medium text-zinc-200">Save on copy</div>
								<p class="mt-1 text-xs text-zinc-500">Store the generated theme locally when copying JSON</p>
							</div>
							<input
								id="save-on-copy"
								type="checkbox"
								checked={saveOnCopy}
								onchange={(e) => appStore.setThemeExportSaveOnCopy((e.target as HTMLInputElement).checked)}
								class="text-brand focus:ring-brand h-4 w-4 rounded border-zinc-600 bg-zinc-800 focus:ring-2"
							/>
						</label>

						<div class="rounded-lg border border-zinc-700/50 bg-zinc-900/60 px-4 py-3">
							<div class="flex items-start justify-between gap-4">
								<div>
									<div class="text-sm font-medium text-zinc-200">Accent boost coefficient</div>
									<p class="mt-1 text-sm text-zinc-500">
										Set between 0.00 and 3.00 (default 1.00). Boosting mainly affects muted mid-tone accents; very
										dark/light or already vivid colors may not change much. Values above 1.00 push harder, but results
										can plateau once saturation/contrast safety limits are reached.
									</p>
								</div>
							</div>

							<div class="mt-3 flex items-center gap-3">
								<input
									type="range"
									min="0"
									max="3"
									step="0.01"
									value={accentBoostCoefficient}
									oninput={handleAccentBoostInput}
									class="accent-brand h-2 w-full cursor-pointer rounded-lg bg-zinc-800"
								/>
								<input
									type="number"
									min="0"
									max="3"
									step="0.01"
									value={accentBoostInput}
									oninput={handleAccentBoostInput}
									onblur={() => (accentBoostInput = accentBoostCoefficient.toFixed(2))}
									class="focus:border-brand/50 w-24 rounded border border-zinc-700 bg-zinc-900 p-2 text-xs text-zinc-300 transition-[border-color,box-shadow,background-color] duration-300 focus:outline-none"
								/>
							</div>
						</div>
					</div>
				</div>

				<div class="mb-8">
					<div class="mb-3 flex items-center justify-between gap-4">
						<div class="flex items-center gap-2">
							<h3 class="text-brand text-sm font-semibold tracking-wide uppercase">Base Color Overrides</h3>
							<span class="text-xs text-zinc-500">Auto-regenerates variants</span>
						</div>
						<div class="flex items-center gap-2">
							<button
								type="button"
								onclick={shuffleThemeDistribution}
								disabled={!themeResult}
								class="hover:border-brand/50 rounded border border-zinc-700/50 px-3 py-1.5 text-xs text-zinc-400 transition-colors hover:bg-zinc-800/50 hover:text-zinc-300 disabled:cursor-not-allowed disabled:opacity-40"
								title="Shuffle C1–C9"
							>
								Shuffle
							</button>
							<button
								type="button"
								onclick={resetOverrides}
								disabled={Object.values(themeOverrides).every((value) => value == null)}
								class="hover:border-brand/50 rounded border border-zinc-700/50 px-3 py-1.5 text-xs text-zinc-400 transition-colors hover:bg-zinc-800/50 hover:text-zinc-300 disabled:cursor-not-allowed disabled:opacity-40"
								title="Return to defaults"
							>
								Reset
							</button>
						</div>
					</div>

					<div class="grid gap-2 md:grid-cols-2">
						{#each overrideFields as field (field.key)}
							{@const isModified = themeOverrides[field.key] != null}
							{@const currentValue = getOverrideValue(field.key)}
							<div
								class="group flex items-center gap-3 rounded-lg border border-zinc-800/80 bg-zinc-900/40 px-3 py-2 transition-colors hover:border-zinc-700/80 hover:bg-zinc-900/60"
							>
								<div class="relative">
									<input
										type="color"
										value={currentValue}
										class="h-8 w-8 cursor-pointer rounded border-0 outline-none"
										oninput={(e) => updateThemeOverride(field.key, (e.target as HTMLInputElement).value)}
										title="Click to pick color"
									/>
									{#if isModified}
										<div class="bg-brand absolute -top-0.5 -right-0.5 h-2 w-2 rounded-full ring-2 ring-zinc-900"></div>
									{/if}
								</div>

								<div class="flex min-w-0 flex-1 items-center gap-2">
									<div class="flex min-w-0 flex-1 flex-col">
										<span class="text-sm font-medium text-zinc-200 {isModified ? 'text-brand' : ''}">{field.label}</span
										>
										<span class="truncate text-xs text-zinc-500">{field.hint}</span>
									</div>

									<input
										type="text"
										value={currentValue}
										placeholder="#000000"
										class="focus:border-brand/60 w-24 rounded border border-zinc-700/50 bg-zinc-800/60 px-2 py-1.5 font-mono text-xs text-zinc-300 placeholder-zinc-600 transition-colors focus:bg-zinc-800 focus:outline-none {isModified
											? 'border-brand/40 bg-zinc-800'
											: ''}"
										oninput={(e) => updateThemeOverride(field.key, (e.target as HTMLInputElement).value)}
									/>
								</div>
							</div>
						{/each}
					</div>
				</div>
			</div>

			<div class="flex items-center justify-between border-t border-zinc-700 bg-zinc-800/50 px-6 py-5">
				<div class="flex items-center gap-2 text-sm text-zinc-400">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						width="16"
						height="16"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
						class="text-brand"
					>
						<path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
						<polyline points="22 4 12 14.01 9 11.01" />
					</svg>
					<span
						>Theme generated from <span class="text-brand font-semibold">{appStore.state.colors.length}</span> palette colors</span
					>
				</div>
				<div class="flex gap-4">
					<button
						type="button"
						onclick={() => popoverStore.close('themeExport')}
						class="hover:border-brand/50 rounded-lg border border-zinc-600 px-5 py-2.5 text-sm font-medium text-zinc-300 transition-[background-color,border-color] duration-300 hover:bg-zinc-800/50"
					>
						Cancel
					</button>
					<button
						type="button"
						onclick={handleExportTheme}
						disabled={!themeResult || themeNameError !== null}
						class="bg-brand shadow-brand/20 hover:shadow-brand/40 rounded-lg px-5 py-2.5 text-sm font-semibold text-zinc-900 transition-[transform,box-shadow] duration-300 hover:scale-105 disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:scale-100"
					>
						Copy Theme JSON
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}
