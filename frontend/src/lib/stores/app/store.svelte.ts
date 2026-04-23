import { browser } from '$app/environment';
import { tick } from 'svelte';
import toast from 'svelte-french-toast';

import * as paletteApi from '$lib/api/palette';
import * as themeApi from '$lib/api/theme';
import { downloadImage } from '$lib/api/wallhaven';
import { CANVAS } from '$lib/types/canvas';
import { IMAGE } from '$lib/types/image';
import { DEFAULT_SELECTOR_ID, SELECTION } from '$lib/types/selector';
import { DEFAULT_APPLY_PALETTE_SETTINGS, type ApplyPaletteSettings } from '$lib/types/applyPaletteSettings';
import type { AppPreferences, AppPreferencesPayload } from '$lib/types/appPreferences';
import type { AppState } from '$lib/stores/app/appState';
import type { Color } from '$lib/types/color';
import type { PaletteData } from '$lib/types/palette';
import { DEFAULT_THEME_EXPORT_PREFERENCES, type SavedThemeItem, type ThemeExportPreferences } from '$lib/types/theme';
import { DEFAULT_WALLHAVEN_SETTINGS, type WallhavenSettings } from '$lib/types/wallhaven';
import type { EditorThemeType, ThemeAppearance } from '$lib/types/themeApi';

import { loadSavedThemes, saveSavedThemes } from '$lib/stores/app/persistence/savedThemes';
import {
	loadWallhavenSettings,
	parseWallhavenSettings,
	saveWallhavenSettings,
	clearWallhavenSettings
} from '$lib/stores/app/persistence/wallhavenSettings';
import {
	loadApplyPaletteSettings,
	parseApplyPaletteSettings,
	saveApplyPaletteSettings,
	clearApplyPaletteSettings
} from '$lib/stores/app/persistence/applyPaletteSettings';
import {
	clearThemeExportPreferences,
	loadThemeExportPreferences,
	parseThemeExportPreferences,
	saveThemeExportPreferences
} from '$lib/stores/app/persistence/themeExport';
import * as preferencesApi from '$lib/api/preferences';
import * as themesApi from '$lib/api/savedThemes';
import { dialogStore } from '$lib/stores/dialog.svelte';
import { isLocalId } from '$lib/localId';

import { authStore } from '../auth.svelte';

function createAppStore() {
	const themeExportPreferences = loadThemeExportPreferences();
	const state = $state<AppState>({
		fileInput: null,
		canvas: null,
		canvasContext: null,
		image: null,
		originalImageWidth: 0,
		originalImageHeight: 0,
		canvasScaleX: 1,
		canvasScaleY: 1,

		searchQuery: '',

		imageLoaded: false,
		isDragging: false,
		isExtracting: false,
		startX: 0,
		startY: 0,
		dragRect: null,
		dragScaleX: 1,
		dragScaleY: 1,

		colors: [],
		wallhavenResults: [],
		wallhavenSettings: loadWallhavenSettings(),
		selectors: [
			{ id: DEFAULT_SELECTOR_ID, color: 'oklch(79.2% 0.209 151.711)', selected: true },
			{ id: 'red', color: 'oklch(64.5% 0.246 16.439)', selected: false },
			{ id: 'blue', color: 'oklch(71.5% 0.143 215.221)', selected: false }
		],
		activeSelectorId: DEFAULT_SELECTOR_ID,
		newFilterColor: '',
		savedPalettes: [],
		sortMethod: 'none',
		applyPaletteSettings: loadApplyPaletteSettings(),
		themeExport: {
			themeResult: null,
			themeName: 'Generated Theme',
			lastGeneratedPaletteVersion: 0,
			editorType: themeExportPreferences.editorType,
			appearance: themeExportPreferences.appearance,
			saveOnCopy: themeExportPreferences.saveOnCopy,
			boostCoefficient: themeExportPreferences.boostCoefficient,
			themeVersions: {},
			rawThemeOverrides: {},
			hasManualBackgroundOverride: false,
			hasManualForegroundOverride: false,
			loadedThemeOverridesReference: null,
			backupColors: null
		},
		savedThemes: loadSavedThemes(),
		paletteVersion: 0
	});

	let lastPaletteSignature = '';
	let preferencesSyncTimer: ReturnType<typeof setTimeout> | null = null;
	let redrawFrameId: number | null = null;

	function isObjectRecord(value: unknown): value is Record<string, unknown> {
		return typeof value === 'object' && value !== null && !Array.isArray(value);
	}

	function toPartialObject<T extends object>(value: unknown): Partial<T> | undefined {
		if (!isObjectRecord(value)) return undefined;
		return value as Partial<T>;
	}

	function toPreferencesPayload(value: unknown): AppPreferencesPayload {
		if (!isObjectRecord(value)) return {};
		return {
			applyPaletteSettings: toPartialObject<ApplyPaletteSettings>(value.applyPaletteSettings),
			wallhavenSettings: toPartialObject<WallhavenSettings>(value.wallhavenSettings),
			themeExport: toPartialObject<ThemeExportPreferences>(value.themeExport)
		};
	}

	function hasNonDefaultPreferences(preferences: AppPreferencesPayload): boolean {
		const apply = parseApplyPaletteSettings(preferences.applyPaletteSettings);
		const wallhaven = parseWallhavenSettings(preferences.wallhavenSettings);
		const themeExport = parseThemeExportPreferences(preferences.themeExport);

		const isApplyDefault =
			apply.luminosity === DEFAULT_APPLY_PALETTE_SETTINGS.luminosity &&
			apply.nearest === DEFAULT_APPLY_PALETTE_SETTINGS.nearest &&
			apply.power === DEFAULT_APPLY_PALETTE_SETTINGS.power &&
			apply.maxDistance === DEFAULT_APPLY_PALETTE_SETTINGS.maxDistance;

		const isWallhavenDefault =
			wallhaven.categories === DEFAULT_WALLHAVEN_SETTINGS.categories &&
			wallhaven.purity === DEFAULT_WALLHAVEN_SETTINGS.purity &&
			wallhaven.sorting === DEFAULT_WALLHAVEN_SETTINGS.sorting &&
			wallhaven.order === DEFAULT_WALLHAVEN_SETTINGS.order &&
			wallhaven.topRange === DEFAULT_WALLHAVEN_SETTINGS.topRange &&
			wallhaven.apikey === DEFAULT_WALLHAVEN_SETTINGS.apikey &&
			wallhaven.ratios.length === DEFAULT_WALLHAVEN_SETTINGS.ratios.length &&
			wallhaven.ratios.every((ratio, index) => ratio === DEFAULT_WALLHAVEN_SETTINGS.ratios[index]);

		const isThemeExportDefault =
			themeExport.editorType === DEFAULT_THEME_EXPORT_PREFERENCES.editorType &&
			themeExport.appearance === DEFAULT_THEME_EXPORT_PREFERENCES.appearance &&
			themeExport.saveOnCopy === DEFAULT_THEME_EXPORT_PREFERENCES.saveOnCopy &&
			themeExport.boostCoefficient === DEFAULT_THEME_EXPORT_PREFERENCES.boostCoefficient;

		return !(isApplyDefault && isWallhavenDefault && isThemeExportDefault);
	}

	$effect.root(() => {
		$effect(() => {
			if (state.colors.length !== 0) {
				const signature = state.colors
					.map((color) => color.hex.toUpperCase())
					.sort()
					.join('|');
				if (signature === lastPaletteSignature) return;
				lastPaletteSignature = signature;
				state.paletteVersion += 1;
			}
		});

		$effect(() => {
			const settings = state.wallhavenSettings;
			if (settings) {
				saveWallhavenSettings(settings);
				appStore.persistPreferencesLocal();
			}
		});

		$effect(() => {
			const settings = state.applyPaletteSettings;
			if (settings) {
				saveApplyPaletteSettings(settings);
				appStore.persistPreferencesLocal();
			}
		});
	});

	function calculateImageDimensions(
		originalWidth: number,
		originalHeight: number,
		maxWidth = CANVAS.MAX_WIDTH,
		maxHeight = CANVAS.MAX_HEIGHT
	) {
		const scale = Math.min(maxWidth / originalWidth, maxHeight / originalHeight, 1);

		const width = originalWidth * scale;
		const height = originalHeight * scale;

		return {
			width,
			height,
			scaleX: originalWidth / width,
			scaleY: originalHeight / height
		};
	}

	function getMousePos(event: MouseEvent) {
		if (!state.canvas) return { x: 0, y: 0 };

		const rect = state.dragRect ?? state.canvas.getBoundingClientRect();
		const scaleX = state.dragRect ? state.dragScaleX : state.canvas.width / rect.width;
		const scaleY = state.dragRect ? state.dragScaleY : state.canvas.height / rect.height;

		return {
			x: (event.clientX - rect.left) * scaleX,
			y: (event.clientY - rect.top) * scaleY
		};
	}

	function isValidSelection(selection: { x: number; y: number; w: number; h: number }): boolean {
		return (
			selection.w > SELECTION.MIN_WIDTH &&
			selection.h > SELECTION.MIN_HEIGHT &&
			!isNaN(selection.x) &&
			!isNaN(selection.y) &&
			isFinite(selection.x) &&
			isFinite(selection.y)
		);
	}

	function isValidHexColor(hex: string): boolean {
		return /^#[0-9A-F]{6}$/i.test(hex);
	}

	function drawImageAndBoxes() {
		if (!state.canvasContext || !state.canvas || !state.image) return;

		state.canvasContext.clearRect(0, 0, state.canvas.width, state.canvas.height);
		state.canvasContext.drawImage(state.image, 0, 0, state.canvas.width, state.canvas.height);

		const animationOffset = -(Date.now() / SELECTION.ANIMATION_SPEED) % 10;

		state.selectors.forEach((selector) => {
			if (!selector.selection || !state.canvasContext || !state.canvas) return;

			const { x, y, w, h } = selector.selection;
			const clampedX = Math.max(0, Math.min(x, state.canvas.width - 1));
			const clampedY = Math.max(0, Math.min(y, state.canvas.height - 1));
			const clampedW = Math.min(w, state.canvas.width - clampedX);
			const clampedH = Math.min(h, state.canvas.height - clampedY);

			if (clampedW <= 0 || clampedH <= 0) return;

			state.canvasContext.save();
			state.canvasContext.strokeStyle = 'rgba(255, 255, 255, 0.95)';
			state.canvasContext.lineWidth = SELECTION.STROKE_WIDTH.OUTER;
			state.canvasContext.strokeRect(clampedX - 2, clampedY - 2, clampedW + 4, clampedH + 4);

			state.canvasContext.strokeStyle = 'rgba(0, 0, 0, 0.8)';
			state.canvasContext.lineWidth = SELECTION.STROKE_WIDTH.MIDDLE;
			state.canvasContext.strokeRect(clampedX - 1, clampedY - 1, clampedW + 2, clampedH + 2);

			state.canvasContext.strokeStyle = selector.color;
			state.canvasContext.lineWidth = SELECTION.STROKE_WIDTH.INNER;
			state.canvasContext.strokeRect(clampedX, clampedY, clampedW, clampedH);

			if (selector.selected) {
				state.canvasContext.strokeStyle = 'rgba(255, 255, 255, 0.9)';
				state.canvasContext.lineWidth = SELECTION.STROKE_WIDTH.ANIMATED;
				state.canvasContext.setLineDash(SELECTION.DASH_PATTERN);
				state.canvasContext.lineDashOffset = animationOffset;
				state.canvasContext.strokeRect(clampedX + 2, clampedY + 2, clampedW - 4, clampedH - 4);
			}
			state.canvasContext.restore();
		});
	}

	function scheduleCanvasRedraw() {
		if (!browser) {
			drawImageAndBoxes();
			return;
		}

		if (redrawFrameId !== null) return;

		redrawFrameId = window.requestAnimationFrame(() => {
			redrawFrameId = null;
			drawImageAndBoxes();
		});
	}

	async function getCleanImageBlob(): Promise<Blob | null> {
		if (!state.canvas || !state.image) return null;

		const tempCanvas = document.createElement('canvas');
		tempCanvas.width = state.canvas.width;
		tempCanvas.height = state.canvas.height;
		const tempCtx = tempCanvas.getContext('2d');
		if (!tempCtx) return null;

		tempCtx.drawImage(state.image, 0, 0, state.canvas.width, state.canvas.height);

		return await new Promise<Blob | null>((resolve) => {
			tempCanvas.toBlob((blob) => resolve(blob), IMAGE.OUTPUT_FORMAT);
		});
	}

	return {
		get state() {
			return state;
		},

		setThemeExportEditorType(editorType: EditorThemeType) {
			state.themeExport.editorType = editorType;
			saveThemeExportPreferences({
				editorType,
				appearance: state.themeExport.appearance,
				saveOnCopy: state.themeExport.saveOnCopy,
				boostCoefficient: state.themeExport.boostCoefficient
			});
			this.persistPreferencesLocal();
		},

		setThemeExportAppearance(appearance: ThemeAppearance) {
			state.themeExport.appearance = appearance;
			saveThemeExportPreferences({
				editorType: state.themeExport.editorType,
				appearance,
				saveOnCopy: state.themeExport.saveOnCopy,
				boostCoefficient: state.themeExport.boostCoefficient
			});
			this.persistPreferencesLocal();
		},

		setThemeExportSaveOnCopy(saveOnCopy: boolean) {
			state.themeExport.saveOnCopy = saveOnCopy;
			saveThemeExportPreferences({
				editorType: state.themeExport.editorType,
				appearance: state.themeExport.appearance,
				saveOnCopy,
				boostCoefficient: state.themeExport.boostCoefficient
			});
			this.persistPreferencesLocal();
		},

		setThemeExportBoostCoefficient(boostCoefficient: number) {
			state.themeExport.boostCoefficient = boostCoefficient;
			saveThemeExportPreferences({
				editorType: state.themeExport.editorType,
				appearance: state.themeExport.appearance,
				saveOnCopy: state.themeExport.saveOnCopy,
				boostCoefficient
			});
			this.persistPreferencesLocal();
		},

		resetThemeExportSession() {
			state.themeExport.themeResult = null;
			state.themeExport.themeName = 'Generated Theme';
			state.themeExport.lastGeneratedPaletteVersion = 0;
			state.themeExport.themeVersions = {};
			state.themeExport.rawThemeOverrides = {};
			state.themeExport.hasManualBackgroundOverride = false;
			state.themeExport.hasManualForegroundOverride = false;
			state.themeExport.loadedThemeOverridesReference = null;
			state.themeExport.backupColors = null;
		},

		saveThemeToLocal(theme: SavedThemeItem) {
			const preparedTheme = this.ensureThemeSignature(theme);
			state.savedThemes = [preparedTheme, ...state.savedThemes];
			saveSavedThemes(state.savedThemes);
			this.persistThemeChange(preparedTheme, 'create');
		},

		replaceSavedTheme(themeId: string, theme: SavedThemeItem) {
			const index = state.savedThemes.findIndex((item) => item.id === themeId);
			const preparedTheme = this.ensureThemeSignature(theme);
			if (index === -1) {
				state.savedThemes = [preparedTheme, ...state.savedThemes];
			} else {
				const next = [...state.savedThemes];
				next[index] = preparedTheme;
				state.savedThemes = next;
			}
			saveSavedThemes(state.savedThemes);
			this.persistThemeChange(preparedTheme, 'update', themeId);
		},

		deleteTheme(themeId: string) {
			state.savedThemes = state.savedThemes.filter((item) => item.id !== themeId);
			saveSavedThemes(state.savedThemes);
			this.persistThemeChange(null, 'delete', themeId);
		},

		async deleteThemes(themeIds: string[]) {
			if (themeIds.length === 0) return;

			const uniqueThemeIds = Array.from(new Set(themeIds));
			const previousThemes = [...state.savedThemes];
			state.savedThemes = state.savedThemes.filter((item) => !uniqueThemeIds.includes(item.id));
			saveSavedThemes(state.savedThemes);

			if (!browser || !authStore.state.isAuthenticated) return;

			const remoteThemeIds = uniqueThemeIds.filter((id) => !isLocalId(id));
			if (remoteThemeIds.length === 0) return;

			try {
				await themesApi.deleteThemes(remoteThemeIds);
				await this.loadSavedThemesFromApi();
			} catch (error) {
				console.error('Failed to sync theme batch delete:', error);
				state.savedThemes = previousThemes;
				saveSavedThemes(state.savedThemes);
				toast.error('Could not delete the themes. Please try again.');
			}
		},

		async setThemeShared(themeId: string, shared: boolean) {
			if (!browser || !authStore.state.isAuthenticated) {
				toast.error('Sign in to share themes.');
				return;
			}

			try {
				const response = shared ? await themesApi.shareTheme(themeId) : await themesApi.unshareTheme(themeId);
				this.applyThemeResponse(response.theme);
				toast.success(shared ? 'Theme shared' : 'Theme removed from shared list');
			} catch {
				toast.error(
					shared ? 'Could not share the theme. Please try again.' : 'Could not unshare the theme. Please try again.'
				);
			}
		},

		async persistThemeChange(theme: SavedThemeItem | null, action: 'create' | 'update' | 'delete', themeId?: string) {
			if (!browser) return;
			if (!authStore.state.isAuthenticated) return;
			const preparedTheme = theme ? this.ensureThemeSignature(theme) : null;

			try {
				if (action === 'create' && preparedTheme) {
					const response = await themesApi.saveTheme(preparedTheme);
					this.applyThemeResponse(response.theme, preparedTheme.id);
				}
				if (action === 'update' && preparedTheme && themeId) {
					const response = await themesApi.updateTheme(themeId, preparedTheme);
					this.applyThemeResponse(response.theme);
				}
				if (action === 'delete' && themeId) {
					await themesApi.deleteTheme(themeId);
				}
			} catch (error) {
				console.error('Failed to sync theme change:', error);
			}
		},

		async syncSavedThemesOnAuth() {
			if (!browser) return;

			if (authStore.state.isAuthenticated) {
				const stored = localStorage.getItem('savedThemes');
				if (stored) {
					try {
						const localThemes = loadSavedThemes().map((theme) => this.ensureThemeSignature(theme));
						if (localThemes.length > 0) {
							const response = await themesApi.saveThemes(localThemes);
							response.themes.forEach((theme, index) => {
								const sourceTheme = localThemes[index];
								this.applyThemeResponse(theme, sourceTheme?.id);
							});
						}
						localStorage.removeItem('savedThemes');
					} catch (error) {
						console.error('Failed to sync local themes:', error);
					}
				}
				await this.loadSavedThemesFromApi();
			} else {
				if (state.savedThemes.length > 0) {
					localStorage.setItem('savedThemes', JSON.stringify(state.savedThemes));
				}
			}
		},

		async loadSavedThemesFromApi() {
			try {
				const response = await themesApi.getThemes();
				state.savedThemes = response.themes;
				saveSavedThemes(state.savedThemes);
			} catch (error) {
				console.error('Failed to load saved themes:', error);
				state.savedThemes = [];
			}
		},

		applyThemeResponse(theme: SavedThemeItem, sourceThemeId?: string) {
			const preparedTheme = this.ensureThemeSignature(theme);
			const preparedResultSignature = this.getThemeSignature(preparedTheme.themeResult);
			const existingIndex = state.savedThemes.findIndex(
				(item) =>
					(sourceThemeId ? item.id === sourceThemeId : false) ||
					item.id === preparedTheme.id ||
					(item.signature && preparedTheme.signature && item.signature === preparedTheme.signature) ||
					this.getThemeSignature(item.themeResult) === preparedResultSignature
			);
			if (existingIndex === -1) {
				state.savedThemes = [preparedTheme, ...state.savedThemes];
			} else {
				const next = [...state.savedThemes];
				next[existingIndex] = preparedTheme;
				state.savedThemes = next;
			}
			saveSavedThemes(state.savedThemes);
		},

		ensureThemeSignature(theme: SavedThemeItem): SavedThemeItem {
			if (theme.signature) return theme;
			const signature = this.getThemeSignature(theme.themeResult);
			return {
				...theme,
				signature
			};
		},

		getThemeSignature(themeResult: SavedThemeItem['themeResult'] | null): string {
			if (!themeResult) return '';
			try {
				return JSON.stringify(themeResult);
			} catch {
				return '';
			}
		},

		async syncPreferencesOnAuth() {
			if (!browser) return;

			if (authStore.state.isAuthenticated) {
				try {
					const response = await preferencesApi.getPreferences<AppPreferencesPayload>();
					const preferences = toPreferencesPayload(response.preferences);
					const stored = localStorage.getItem('appPreferences');
					let localPreferences: AppPreferencesPayload = {};
					if (stored) {
						try {
							localPreferences = toPreferencesPayload(JSON.parse(stored));
						} catch {
							localPreferences = {};
						}
					}

					const hasLocalPreferences = Object.keys(localPreferences).length > 0;
					if (hasNonDefaultPreferences(preferences) || !hasLocalPreferences) {
						this.applyPreferences(preferences);
						localStorage.removeItem('appPreferences');
						return;
					}

					const merged: AppPreferencesPayload = { ...preferences, ...localPreferences };
					const updated = this.applyPreferences(merged);
					await preferencesApi.savePreferences(updated);
					localStorage.removeItem('appPreferences');
				} catch (error) {
					console.error('Failed to sync preferences:', error);
				}
			} else {
				const preferences = this.buildPreferences();
				localStorage.setItem('appPreferences', JSON.stringify(preferences));
			}
		},

		buildPreferences(): AppPreferences {
			return {
				applyPaletteSettings: state.applyPaletteSettings,
				wallhavenSettings: state.wallhavenSettings,
				themeExport: {
					editorType: state.themeExport.editorType,
					appearance: state.themeExport.appearance,
					saveOnCopy: state.themeExport.saveOnCopy,
					boostCoefficient: state.themeExport.boostCoefficient
				}
			};
		},

		persistPreferencesLocal() {
			if (!browser) return;
			const preferences = this.buildPreferences();
			localStorage.setItem('appPreferences', JSON.stringify(preferences));

			if (!authStore.state.isAuthenticated) return;

			if (preferencesSyncTimer) {
				clearTimeout(preferencesSyncTimer);
			}

			preferencesSyncTimer = setTimeout(async () => {
				if (!authStore.state.isAuthenticated) return;
				try {
					await preferencesApi.savePreferences(this.buildPreferences());
					localStorage.removeItem('appPreferences');
				} catch (error) {
					console.error('Failed to persist preferences:', error);
				}
			}, 300);
		},

		applyPreferences(preferences: AppPreferencesPayload): AppPreferences {
			const applyPaletteSettings = this.parseApplyPaletteSettings(preferences.applyPaletteSettings);
			const wallhavenSettings = this.parseWallhavenSettings(preferences.wallhavenSettings);
			const themeExport = this.parseThemeExportPreferences(preferences.themeExport);

			state.applyPaletteSettings = applyPaletteSettings;
			state.wallhavenSettings = wallhavenSettings;
			state.themeExport.editorType = themeExport.editorType;
			state.themeExport.appearance = themeExport.appearance;
			state.themeExport.saveOnCopy = themeExport.saveOnCopy;
			state.themeExport.boostCoefficient = themeExport.boostCoefficient;

			clearApplyPaletteSettings();
			clearWallhavenSettings();
			clearThemeExportPreferences();

			saveApplyPaletteSettings(applyPaletteSettings);
			saveWallhavenSettings(wallhavenSettings);
			saveThemeExportPreferences(themeExport);

			return this.buildPreferences();
		},

		parseApplyPaletteSettings,

		parseWallhavenSettings,

		parseThemeExportPreferences,

		async drawBlobToCanvas(blob: Blob) {
			return new Promise<void>((resolve, reject) => {
				const url = URL.createObjectURL(blob);
				const newImage = new Image();
				newImage.decoding = 'async';

				newImage.onload = () => {
					try {
						const context = state.canvas?.getContext('2d');
						if (!context) throw new Error('Could not get canvas context');
						state.canvasContext = context;
						state.originalImageWidth = newImage.width;
						state.originalImageHeight = newImage.height;

						const dimensions = calculateImageDimensions(state.originalImageWidth, state.originalImageHeight);
						if (!state.canvas || !state.canvasContext) {
							URL.revokeObjectURL(url);
							reject(new Error('Canvas context not available'));
							return;
						}

						state.canvasScaleX = dimensions.scaleX;
						state.canvasScaleY = dimensions.scaleY;
						state.canvas.width = dimensions.width;
						state.canvas.height = dimensions.height;
						state.canvas.style.width = dimensions.width + 'px';
						state.canvas.style.height = dimensions.height + 'px';

						state.canvasContext.drawImage(newImage, 0, 0, dimensions.width, dimensions.height);
						state.image = newImage;
						state.imageLoaded = true;
						scheduleCanvasRedraw();
						URL.revokeObjectURL(url);
						resolve();
					} catch (error) {
						URL.revokeObjectURL(url);
						reject(error);
					}
				};

				newImage.onerror = () => {
					URL.revokeObjectURL(url);
					reject(new Error('Failed to load image'));
				};

				newImage.src = url;
			});
		},

		clearAllSelections() {
			this.state.activeSelectorId = DEFAULT_SELECTOR_ID;

			this.state.selectors.forEach((selector) => {
				selector.selection = undefined;
				selector.selected = selector.id === DEFAULT_SELECTOR_ID;
			});

			this.redrawCanvas();
		},

		redrawCanvas() {
			if (redrawFrameId !== null) {
				cancelAnimationFrame(redrawFrameId);
				redrawFrameId = null;
			}
			drawImageAndBoxes();
		},

		restoreCanvasImage() {
			if (!state.canvas || !state.image || !state.imageLoaded) return;

			const context = state.canvas.getContext('2d');
			if (!context) return;
			state.canvasContext = context;

			const scaledWidth =
				state.canvasScaleX > 0 ? Math.round(state.originalImageWidth / state.canvasScaleX) : state.image.width;
			const scaledHeight =
				state.canvasScaleY > 0 ? Math.round(state.originalImageHeight / state.canvasScaleY) : state.image.height;

			const width = Number.isFinite(scaledWidth) && scaledWidth > 0 ? scaledWidth : state.image.width;
			const height = Number.isFinite(scaledHeight) && scaledHeight > 0 ? scaledHeight : state.image.height;

			state.canvas.width = width;
			state.canvas.height = height;
			state.canvas.style.width = width + 'px';
			state.canvas.style.height = height + 'px';

			this.redrawCanvas();
		},

		handleMouseDown(e: MouseEvent) {
			if (!state.canvas || state.isExtracting || !state.activeSelectorId) return;
			state.isDragging = true;
			state.dragRect = state.canvas.getBoundingClientRect();
			state.dragScaleX = state.canvas.width / state.dragRect.width;
			state.dragScaleY = state.canvas.height / state.dragRect.height;
			const pos = getMousePos(e);
			state.startX = pos.x;
			state.startY = pos.y;
		},

		handleMouseMove(e: MouseEvent) {
			if (state.isExtracting) {
				state.isDragging = false;
				state.dragRect = null;
				return;
			}
			if (!state.isDragging || !state.activeSelectorId) return;
			const pos = getMousePos(e);
			const activeSelector = state.selectors.find((s) => s.id === state.activeSelectorId);
			if (activeSelector) {
				activeSelector.selection = {
					x: Math.min(state.startX, pos.x),
					y: Math.min(state.startY, pos.y),
					w: Math.abs(pos.x - state.startX),
					h: Math.abs(pos.y - state.startY)
				};
			}
			scheduleCanvasRedraw();
		},

		async handleMouseUp() {
			if (state.isExtracting) {
				state.isDragging = false;
				state.dragRect = null;
				return;
			}

			state.isDragging = false;
			state.dragRect = null;
			await appStore.extractPaletteFromSelection();
		},

		async onFileChange(event: Event) {
			const input = event.target as HTMLInputElement;
			const file = input?.files?.[0];
			if (!file) return;

			await appStore.drawBlobToCanvas(file);
			await appStore.extractPalette(file);
		},

		triggerFileSelect() {
			if (state.fileInput) {
				state.fileInput.value = '';
				state.fileInput.click();
			}
		},

		async handleDrop(event: DragEvent) {
			event.preventDefault();
			const files = event.dataTransfer?.files;
			if (files?.length) {
				await appStore.drawBlobToCanvas(files[0]);
				await appStore.extractPalette(files[0]);
			}
		},

		async loadWallhavenImage(imageUrl: string, existingToastId?: string) {
			try {
				this.clearAllSelections();
				const blob = await downloadImage(imageUrl);
				await this.drawBlobToCanvas(blob);
				await this.extractPalette(blob, existingToastId);
			} catch (err) {
				console.error('Failed to load wallhaven image', err);
			}
		},

		async extractPaletteFromSelection() {
			if (state.isExtracting) {
				return;
			}

			state.isExtracting = true;
			const toastId = toast.loading('Extracting palette...');
			await tick();

			if (!state.canvasContext || !state.canvas || !state.image) {
				toast.error('Canvas is not ready yet. Try again in a moment.', { id: toastId });
				state.isExtracting = false;
				return;
			}

			const validSelections = state.selectors.filter((s) => s.selection && isValidSelection(s.selection));

			if (validSelections.length === 0) {
				toast.error('Select an area on the image to extract colors.', { id: toastId });
				state.isExtracting = false;
				return;
			}

			let minX = Infinity,
				minY = Infinity,
				maxX = -Infinity,
				maxY = -Infinity;

			for (const s of validSelections) {
				if (!s.selection) continue;
				const scaledX = Math.round(s.selection.x * state.canvasScaleX);
				const scaledY = Math.round(s.selection.y * state.canvasScaleY);
				const scaledW = Math.round(s.selection.w * state.canvasScaleX);
				const scaledH = Math.round(s.selection.h * state.canvasScaleY);

				minX = Math.min(minX, scaledX);
				minY = Math.min(minY, scaledY);
				maxX = Math.max(maxX, scaledX + scaledW);
				maxY = Math.max(maxY, scaledY + scaledH);
			}

			if (minX < Infinity && minY < Infinity && maxX > -Infinity && maxY > -Infinity) {
				const mergedWidth = maxX - minX;
				const mergedHeight = maxY - minY;

				const mergedCanvas = document.createElement('canvas');
				mergedCanvas.width = mergedWidth;
				mergedCanvas.height = mergedHeight;
				const mergedCtx = mergedCanvas.getContext('2d');

				if (mergedCtx) {
					for (const s of validSelections) {
						if (!s.selection) continue;
						const scaledX = Math.round(s.selection.x * state.canvasScaleX);
						const scaledY = Math.round(s.selection.y * state.canvasScaleY);
						const scaledW = Math.round(s.selection.w * state.canvasScaleX);
						const scaledH = Math.round(s.selection.h * state.canvasScaleY);

						if (scaledX >= 0 && scaledY >= 0 && scaledW > 0 && scaledH > 0) {
							mergedCtx.drawImage(
								state.image,
								scaledX,
								scaledY,
								scaledW,
								scaledH,
								scaledX - minX,
								scaledY - minY,
								scaledW,
								scaledH
							);
						}
					}

					const blob = await new Promise<Blob>((resolve) =>
						mergedCanvas.toBlob((b) => resolve(b!), IMAGE.OUTPUT_FORMAT)
					);

					await appStore.extractPalette(blob, toastId);
				}
			}
		},

		async extractPalette(file: Blob | File, existingToastId?: string) {
			if (!file) {
				toast.error('Choose an image to extract colors.');
				return;
			}

			if (file.size > IMAGE.MAX_FILE_SIZE) {
				toast.error(`File is too large. Choose a file under ${IMAGE.MAX_FILE_SIZE / 1024 / 1024}MB.`);
				return;
			}

			state.isExtracting = true;
			const toastId = existingToastId ?? toast.loading('Extracting palette...');

			try {
				const result = await paletteApi.extractPalette(file);
				if (result.palette.length > 0) {
					this.resetThemeExportSession();
					state.colors = result.palette;
					toast.success('Palette extracted', { id: toastId });
				} else {
					toast.error('No colors found in the selected area. Try a larger selection.', { id: toastId });
				}
			} catch {
				toast.error('Could not extract a palette. Please try again.', {
					id: toastId
				});
			} finally {
				state.isExtracting = false;
			}
		},

		async syncPaletteWithSelections() {
			if (state.isExtracting || !state.canvas || !state.image) return;

			const hasValidSelection = state.selectors.some((selector) =>
				selector.selection ? isValidSelection(selector.selection) : false
			);

			if (hasValidSelection) {
				await this.extractPaletteFromSelection();
				return;
			}

			const cleanCanvas = document.createElement('canvas');
			cleanCanvas.width = state.canvas.width;
			cleanCanvas.height = state.canvas.height;
			const cleanContext = cleanCanvas.getContext('2d');
			if (!cleanContext) return;

			cleanContext.drawImage(state.image, 0, 0, state.canvas.width, state.canvas.height);
			const blob = await new Promise<Blob>((resolve) => cleanCanvas.toBlob((b) => resolve(b!), IMAGE.OUTPUT_FORMAT));
			await this.extractPalette(blob);
		},

		async savePalette() {
			if (!state.colors || state.colors.length === 0) {
				toast.error('Extract a palette before saving.');
				return;
			}
			const paletteName = await dialogStore.prompt({
				title: 'Save current palette',
				message: 'Enter a name for your palette.',
				placeholder: 'My Palette',
				confirmLabel: 'Save palette'
			});
			if (!paletteName) return;

			try {
				if (authStore.state.isAuthenticated) {
					const data = await paletteApi.savePalette(paletteName, state.colors);
					toast.success('Palette saved: ' + data.name);
					await appStore.loadSavedPalettes();
				} else {
					if (browser) {
						const newPalette: PaletteData = {
							id: `local_${Date.now()}`,
							name: paletteName,
							palette: state.colors,
							createdAt: new Date().toISOString()
						};

						const stored = localStorage.getItem('savedPalettes');
						const palettes = stored ? JSON.parse(stored) : [];
						palettes.push(newPalette);
						localStorage.setItem('savedPalettes', JSON.stringify(palettes));

						await appStore.loadSavedPalettes();
						toast.success('Palette saved: ' + paletteName);
					}
				}
			} catch {
				toast.error('Could not save the palette. Please try again.');
			}
		},

		async applyPalette(palette: Color[]) {
			// Before calling this method we are checking if there is an image loaded in the UI

			if (!palette || palette.length === 0) {
				toast.error('Select a palette to apply.');
				return;
			}

			const paletteToApply = [...palette];

			const invalidColors = paletteToApply.filter((color) => !isValidHexColor(color.hex));
			if (invalidColors.length > 0) {
				toast.error('This palette has invalid colors. Please reselect or recreate it.');
				return;
			}

			if (!state.canvas || !state.imageLoaded || !state.image) {
				toast.error('Load an image before applying a palette.');
				return;
			}

			const toastId = toast.loading('Applying palette...');

			try {
				const srcBlob = await new Promise<Blob>((resolve) =>
					state.canvas!.toBlob((b) => resolve(b!), IMAGE.OUTPUT_FORMAT)
				);

				const outBlob = await themeApi.applyPaletteBlob(srcBlob, paletteToApply, {
					luminosity: state.applyPaletteSettings.luminosity,
					nearest: state.applyPaletteSettings.nearest,
					power: state.applyPaletteSettings.power,
					maxDistance: state.applyPaletteSettings.maxDistance
				});
				await appStore.drawBlobToCanvas(outBlob);

				const extracted = await paletteApi.extractPalette(outBlob);
				if (extracted.palette.length > 0) {
					this.resetThemeExportSession();
					state.colors = extracted.palette;
				} else {
					toast.error('Palette applied, but no colors were detected. Try another image.', { id: toastId });
					return;
				}

				toast.success('Applied palette', { id: toastId });
			} catch {
				toast.error('Could not apply the palette. Please try again.', { id: toastId });
			}
		},

		async loadSavedPalettes() {
			if (!browser) return;

			try {
				if (authStore.state.isAuthenticated) {
					const response = await paletteApi.getPalettes();
					state.savedPalettes = response.palettes;
				} else {
					const stored = localStorage.getItem('savedPalettes');
					if (stored) {
						try {
							const localPalettes = JSON.parse(stored) as PaletteData[];
							state.savedPalettes = localPalettes;
						} catch {
							state.savedPalettes = [];
						}
					} else {
						state.savedPalettes = [];
					}
				}
			} catch (error) {
				console.error('Failed to load saved palettes:', error);
				state.savedPalettes = [];
			}
		},

		async deletePalette(paletteId: string) {
			try {
				const isLocalPalette = isLocalId(paletteId);

				if (authStore.state.isAuthenticated && !isLocalPalette) {
					await paletteApi.deletePalette(paletteId);
					toast.success('Palette deleted');
					await appStore.loadSavedPalettes();
				} else {
					if (browser) {
						const stored = localStorage.getItem('savedPalettes');
						const palettes = stored ? JSON.parse(stored) : [];
						const filtered = palettes.filter((p: PaletteData) => p.id !== paletteId);
						localStorage.setItem('savedPalettes', JSON.stringify(filtered));
					}
					await appStore.loadSavedPalettes();
					toast.success('Palette deleted');
				}
			} catch {
				toast.error('Could not delete the palette. Please try again.');
			}
		},

		async deletePalettes(paletteIds: string[]) {
			if (paletteIds.length === 0) return;

			const uniquePaletteIds = Array.from(new Set(paletteIds));

			try {
				if (authStore.state.isAuthenticated) {
					const remotePaletteIds = uniquePaletteIds.filter((id) => !isLocalId(id));
					if (remotePaletteIds.length > 0) {
						await paletteApi.deletePalettes(remotePaletteIds);
					}

					if (browser) {
						const stored = localStorage.getItem('savedPalettes');
						if (stored) {
							const palettes = JSON.parse(stored) as PaletteData[];
							const filtered = palettes.filter((p) => !uniquePaletteIds.includes(p.id));
							localStorage.setItem('savedPalettes', JSON.stringify(filtered));
						}
					}

					await appStore.loadSavedPalettes();
					toast.success('Palettes deleted');
					return;
				}

				if (browser) {
					const stored = localStorage.getItem('savedPalettes');
					const palettes = stored ? (JSON.parse(stored) as PaletteData[]) : [];
					const filtered = palettes.filter((p) => !uniquePaletteIds.includes(p.id));
					localStorage.setItem('savedPalettes', JSON.stringify(filtered));
				}
				await appStore.loadSavedPalettes();
				toast.success('Palettes deleted');
			} catch {
				toast.error('Could not delete the palettes. Please try again.');
			}
		},

		async setPaletteShared(paletteId: string, shared: boolean) {
			if (!browser || !authStore.state.isAuthenticated) {
				toast.error('Sign in to share palettes.');
				return;
			}

			try {
				const response = shared ? await paletteApi.sharePalette(paletteId) : await paletteApi.unsharePalette(paletteId);

				state.savedPalettes = state.savedPalettes.map((item) =>
					item.id === paletteId
						? {
								...item,
								isShared: response.palette.isShared,
								sharedAt: response.palette.sharedAt
							}
						: item
				);
				toast.success(shared ? 'Palette shared' : 'Palette removed from shared list');
			} catch {
				toast.error(
					shared ? 'Could not share the palette. Please try again.' : 'Could not unshare the palette. Please try again.'
				);
			}
		},

		async syncPalettesOnAuth() {
			if (!browser) return;

			if (authStore.state.isAuthenticated) {
				const stored = localStorage.getItem('savedPalettes');
				if (stored) {
					try {
						const localPalettes = JSON.parse(stored) as PaletteData[];
						const palettesToSync = localPalettes.filter((palette) => !palette.id || isLocalId(palette.id));

						if (palettesToSync.length > 0) {
							const toastId = toast.loading('Syncing your palettes...');
							try {
								await paletteApi.savePalettes(
									palettesToSync.map((palette) => ({
										name: palette.name,
										palette: palette.palette
									}))
								);
								toast.success('Palettes synced successfully', { id: toastId });
							} catch {
								toast.error('Could not sync palettes. Try again later.', { id: toastId });
							}
						}

						localStorage.removeItem('savedPalettes');
					} catch {
						console.error('Failed to parse local palettes');
					}
				}
				await appStore.loadSavedPalettes();
			} else {
				if (state.savedPalettes.length > 0) {
					localStorage.setItem('savedPalettes', JSON.stringify(state.savedPalettes));
				}
			}
		},

		async downloadImage() {
			if (!state.canvas || !state.imageLoaded || !state.image) {
				toast.error('Load an image before downloading.');
				return;
			}

			const toastId = toast.loading('Preparing download...');

			try {
				const blob = await getCleanImageBlob();
				if (!blob) {
					toast.error('Could not prepare the download. Please try again.', { id: toastId });
					return;
				}

				const url = URL.createObjectURL(blob);
				const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, -5);
				const filename = `palette-applied-${timestamp}.png`;

				const link = document.createElement('a');
				link.href = url;
				link.download = filename;
				document.body.append(link);
				link.click();
				link.remove();

				setTimeout(() => URL.revokeObjectURL(url), 0);
				toast.success('Image downloaded', { id: toastId });
			} catch {
				toast.error('Download failed. Please try again.', { id: toastId });
			}
		}
	};
}

export const appStore = createAppStore();
