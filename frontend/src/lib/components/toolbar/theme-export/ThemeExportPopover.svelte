<script lang="ts">
	import { onDestroy } from 'svelte';
	import toast from 'svelte-french-toast';

	import { generateOverridable, generateTheme, type EditorThemeType, type ThemeAppearance } from '$lib/api/theme';
	import { detectThemeAppearance, detectThemeType } from '$lib/colorUtils';
	import { isDesktopApp } from '$lib/platform';
	import { appStore } from '$lib/stores/app/store.svelte';
	import { popoverStore } from '$lib/stores/popovers.svelte';
	import { tutorialStore } from '$lib/stores/tutorial.svelte';
	import type { Theme, ThemeGenerationResponse, ThemeOverrides } from '$lib/types/theme';
	import { cn } from '$lib/utils';

	import AppearanceSelector from './AppearanceSelector.svelte';
	import EditorSelector from './EditorSelector.svelte';
	import OverrideGrid from './OverrideGrid.svelte';
	import { getRecommendedOverrideColors } from './recommendations';
	import { exportTheme as exportThemeToClipboard, exportThemeToEditorFolder } from './save';
	import SaveBehaviorSection from './SaveBehaviorSection.svelte';
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

	const SHUFFLE_KEYS: Array<keyof ThemeOverrides> = ['c1', 'c2', 'c3', 'c4', 'c5', 'c6', 'c7', 'c8', 'c9'];
	const OVERRIDE_KEYS = overrideFields.map((field) => field.key);

	let themeNameDebounceTimer: ReturnType<typeof setTimeout> | null = null;
	let manualOverrideBadges = $state<Partial<Record<keyof ThemeOverrides, boolean>>>({});

	let paletteVersion = $derived(appStore.state.paletteVersion);
	let themeName = $derived(appStore.state.themeExport.themeName);
	let editorType = $derived(appStore.state.themeExport.editorType);
	let themeAppearance = $derived(appStore.state.themeExport.appearance);
	let themeResult = $derived(appStore.state.themeExport.themeResult);
	let themeOverrides = $derived(appStore.state.themeExport.themeResult?.themeOverrides ?? {});
	let saveOnCopy = $derived(appStore.state.themeExport.saveOnCopy);
	let accentBoostCoefficient = $derived(appStore.state.themeExport.boostCoefficient);
	let backupColors = $derived(appStore.state.themeExport.backupColors ?? []);
	let themeSourceColorCount = $derived(
		appStore.state.colors.length > 0 ? appStore.state.colors.length : backupColors.length
	);
	let paletteColors = $derived(appStore.state.colors.length > 0 ? appStore.state.colors : backupColors);

	const baseOverrides = $derived(themeResult?.themeOverrides ?? {});

	let isOpen = $derived(popoverStore.isOpen('themeExport'));
	let themeNameError = $state<string | null>(null);
	let isGenerating = $state(false);
	let isExportDisabled = $derived(!themeResult || themeNameError !== null);

	onDestroy(() => {
		if (themeNameDebounceTimer) {
			clearTimeout(themeNameDebounceTimer);
		}
	});

	$effect(() => {
		// effect handling scenarios:
		// 1. Popover opened with existing theme but never generated (lastGeneratedPaletteVersion === 0)
		// 2. Popover opened without any theme result (themeResult === null)
		// 3. Popover is open and palette version changed (regeneration needed)

		if (!isOpen || isGenerating) return;

		const lastGenerated = appStore.state.themeExport.lastGeneratedPaletteVersion;
		const shouldGenerate =
			themeResult === null || // Scenario 2: No theme exists
			(themeResult !== null && appStore.state.colors.length > 0 && lastGenerated === 0) || // Scenario 1: Theme exists but never generated
			(lastGenerated !== 0 && lastGenerated !== paletteVersion); // Scenario 3: Palette version changed

		if (shouldGenerate) {
			themeName = 'Generated Theme';
			manualOverrideBadges = {};
			resetThemeExportOverrideState();
			generateThemeFromApi({ overrides: {}, bypassCache: true });
		}
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
		const hasBackupColors = backupColors.length > 0;
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
			} else if (hasBackupColors) {
				response = await generateTheme(
					backupColors,
					nextType,
					themeName.trim(),
					requestOverrides,
					nextAppearance,
					requestAccentBoostCoefficient
				);
			} else {
				if (!themeResult?.theme) {
					return;
				}

				if (
					(nextType !== detectThemeType(themeResult.theme) ||
						nextAppearance !== detectThemeAppearance(themeResult.theme)) &&
					hasBackupColors
				) {
					response = await generateTheme(
						backupColors,
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
		manualOverrideBadges = {};

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
			manualOverrideBadges[key] = true;
		}

		appStore.state.themeExport.rawThemeOverrides = { ...overrides };

		generateThemeFromApi({ overrides, bypassCache: true });
	}

	function getOverrideValue(key: keyof ThemeOverrides): string {
		return themeOverrides[key] ?? baseOverrides[key] ?? '#000000';
	}

	function getOverrideRecommendations(key: keyof ThemeOverrides, currentValue: string): string[] {
		return getRecommendedOverrideColors({
			key,
			currentValue,
			themeOverrides,
			baseOverrides,
			paletteColors,
			overrideKeys: OVERRIDE_KEYS
		});
	}

	function handleClose() {
		popoverStore.close('themeExport');
	}

	function updateThemeOverride(key: keyof ThemeOverrides, value: string) {
		const normalized = normalizeHex(value);
		if (!normalized) return;

		clearThemeVersions();

		const overrides = getCurrentOverrides();
		overrides[key] = normalized;
		manualOverrideBadges[key] = true;
		appStore.state.themeExport.rawThemeOverrides = { ...overrides };
		setManualOverrideFlag(key, true);

		generateThemeFromApi({ overrides, bypassCache: true });
	}

	function switchThemeOverride(key: keyof ThemeOverrides, sourceKey: keyof ThemeOverrides) {
		if (key === sourceKey) return;

		const currentValue = normalizeHex(getOverrideValue(key));
		const sourceValue = normalizeHex(getOverrideValue(sourceKey));
		if (!currentValue || !sourceValue) return;

		clearThemeVersions();

		const overrides = getCurrentOverrides();
		overrides[key] = sourceValue;
		overrides[sourceKey] = currentValue;
		manualOverrideBadges[key] = true;
		manualOverrideBadges[sourceKey] = true;
		appStore.state.themeExport.rawThemeOverrides = { ...overrides };
		setManualOverrideFlag(key, true);
		setManualOverrideFlag(sourceKey, true);

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

	function handleAccentBoostChange(normalized: number) {
		appStore.setThemeExportBoostCoefficient(normalized);
		generateThemeFromApi({ bypassCache: true, accentBoostCoefficient: normalized });
	}

	function handleSaveOnCopyChange(checked: boolean) {
		appStore.setThemeExportSaveOnCopy(checked);
	}

	async function handleExportTheme() {
		await exportThemeToClipboard({
			name: themeName,
			editorType,
			themeResult,
			saveOnCopy,
			onExported: () => tutorialStore.setThemeJsonCopied(true)
		});
	}

	async function handleSaveThemeToEditorFolder() {
		await exportThemeToEditorFolder({
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
			class={cn(
				'share-modal-content border-brand/50 shadow-brand/20 animate-scale-in relative flex w-full max-w-4xl flex-col overflow-hidden rounded-xl border bg-zinc-900 shadow-2xl xl:max-w-6xl',
				isDesktopApp ? 'mt-12 max-h-[90svh]' : 'max-h-[95svh]'
			)}
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
					onclick={handleClose}
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

				<div id="tutorial-theme-appearance">
					<AppearanceSelector selected={themeAppearance} onSelect={handleThemeAppearanceChange} />
				</div>

				<div id="tutorial-theme-editor">
					<EditorSelector selected={editorType} onSelect={handleEditorTypeChange} />
				</div>

				<SaveBehaviorSection
					{saveOnCopy}
					{accentBoostCoefficient}
					onSaveOnCopyChange={handleSaveOnCopyChange}
					onAccentBoostChange={handleAccentBoostChange}
				/>

				<div id="tutorial-theme-overrides">
					<OverrideGrid
						{overrideFields}
						{themeOverrides}
						themeResultExists={Boolean(themeResult)}
						{manualOverrideBadges}
						{getOverrideValue}
						{getOverrideRecommendations}
						onUpdateOverride={updateThemeOverride}
						onSwitchOverride={switchThemeOverride}
						onShuffle={shuffleThemeDistribution}
						onReset={resetOverrides}
					/>
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
						>Theme generated from <span class="text-brand font-semibold">{themeSourceColorCount}</span> palette colors</span
					>
				</div>
				<div class="flex items-center gap-4">
					<button
						type="button"
						onclick={handleClose}
						class="hover:border-brand/50 rounded-lg border border-zinc-600 px-5 py-2.5 text-sm font-medium text-zinc-300 transition-[background-color,border-color] duration-300 hover:bg-zinc-800/50"
					>
						Cancel
					</button>
					<div class="ml-auto flex items-center gap-3">
						{#if isDesktopApp}
							<button
								type="button"
								onclick={handleSaveThemeToEditorFolder}
								disabled={isExportDisabled}
								class="hover:border-brand/50 rounded-lg border border-zinc-600 px-5 py-2.5 text-sm font-medium text-zinc-300 transition-[background-color,border-color] duration-300 hover:bg-zinc-800/50 disabled:cursor-not-allowed disabled:opacity-50"
							>
								Save To Editor Folder
							</button>
						{/if}
						<button
							id="tutorial-theme-copy-json"
							type="button"
							onclick={handleExportTheme}
							disabled={isExportDisabled}
							class="bg-brand shadow-brand/20 hover:shadow-brand/40 rounded-lg px-5 py-2.5 text-sm font-semibold text-zinc-900 transition-[transform,box-shadow] duration-300 hover:scale-105 disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:scale-100"
						>
							Copy Theme JSON
						</button>
					</div>
				</div>
			</div>
		</div>
	</div>
{/if}
