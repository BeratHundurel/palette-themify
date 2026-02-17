import { browser } from '$app/environment';
import { tick } from 'svelte';
import toast from 'svelte-french-toast';

import * as paletteApi from '$lib/api/palette';
import * as themeApi from '$lib/api/theme';
import * as workspaceApi from '$lib/api/workspace';
import { downloadImage } from '$lib/api/wallhaven';
import { CANVAS, SELECTION, IMAGE, UI } from '$lib/constants';
import type { ApplyPaletteSettings } from '$lib/types/applyPaletteSettings';
import type { Color } from '$lib/types/color';
import type { Selector } from '$lib/types/selector';
import type { PaletteData } from '$lib/types/palette';
import type { WorkspaceData } from '$lib/types/workspace';
import type { SavedThemeItem, ThemeExportState } from '$lib/types/theme';
import type { WallhavenResult, WallhavenSettings } from '$lib/types/wallhaven';
import type { EditorThemeType } from '$lib/api/theme';
import type { SortMethod } from '$lib/colorUtils';

import { authStore } from './auth.svelte';
import { tutorialStore } from './tutorial.svelte';

export type SavedPaletteItem = PaletteData;

interface AppState {
	fileInput: HTMLInputElement | null;
	canvas: HTMLCanvasElement | null;
	canvasContext: CanvasRenderingContext2D | null;
	image: HTMLImageElement | null;
	originalImageWidth: number;
	originalImageHeight: number;
	canvasScaleX: number;
	canvasScaleY: number;

	searchQuery: string;

	imageLoaded: boolean;
	isDragging: boolean;
	isExtracting: boolean;
	startX: number;
	startY: number;
	dragRect: DOMRect | null;
	dragScaleX: number;
	dragScaleY: number;

	colors: Color[];
	wallhavenResults: WallhavenResult[];
	wallhavenSettings: WallhavenSettings;
	selectors: Selector[];
	activeSelectorId: string;
	newFilterColor: string;
	savedPalettes: PaletteData[];
	savedWorkspaces: WorkspaceData[];
	sortMethod: SortMethod;
	applyPaletteSettings: ApplyPaletteSettings;
	themeExport: ThemeExportState;
	savedThemes: SavedThemeItem[];
	paletteVersion: number;
}

const THEME_EXPORT_STORAGE_KEY = 'themeExportPreferences';
const SAVED_THEMES_STORAGE_KEY = 'savedThemes';
const WALLHAVEN_SETTINGS_KEY = 'wallhavenSettings';
const APPLY_PALETTE_SETTINGS_KEY = 'applyPaletteSettings';

const DEFAULT_WALLHAVEN_SETTINGS: WallhavenSettings = {
	categories: '111',
	purity: '100',
	sorting: 'relevance',
	order: 'desc',
	topRange: '1M',
	ratios: [],
	apikey: ''
};

const DEFAULT_APPLY_PALETTE_SETTINGS: ApplyPaletteSettings = {
	luminosity: 1,
	nearest: 30,
	power: 4,
	maxDistance: 0
};

function loadThemeExportPreferences(): EditorThemeType {
	if (!browser) return 'vscode';
	const stored = localStorage.getItem(THEME_EXPORT_STORAGE_KEY);
	if (stored) {
		const parsed = JSON.parse(stored);
		return parsed.editorType || 'vscode';
	}
	return 'vscode';
}

function loadThemeExportSaveOnCopy(): boolean {
	if (!browser) return true;
	const stored = localStorage.getItem(THEME_EXPORT_STORAGE_KEY);
	if (stored) {
		const parsed = JSON.parse(stored);
		return parsed.saveOnCopy ?? true;
	}
	return true;
}

function saveThemeExportPreferences(editorType: EditorThemeType, saveOnCopy: boolean) {
	if (!browser) return;
	localStorage.setItem(THEME_EXPORT_STORAGE_KEY, JSON.stringify({ editorType, saveOnCopy }));
}

function loadSavedThemes(): SavedThemeItem[] {
	if (!browser) return [];
	try {
		const stored = localStorage.getItem(SAVED_THEMES_STORAGE_KEY);
		if (!stored) return [];
		const parsed = JSON.parse(stored) as SavedThemeItem[];
		if (!Array.isArray(parsed)) return [];
		return parsed;
	} catch {
		return [];
	}
}

function saveSavedThemes(themes: SavedThemeItem[]) {
	if (!browser) return;
	localStorage.setItem(SAVED_THEMES_STORAGE_KEY, JSON.stringify(themes));
}

function loadWallhavenSettings(): WallhavenSettings {
	if (!browser) return DEFAULT_WALLHAVEN_SETTINGS;
	const stored = localStorage.getItem(WALLHAVEN_SETTINGS_KEY);
	if (stored) {
		const parsed = JSON.parse(stored);
		return {
			categories: parsed.categories ?? DEFAULT_WALLHAVEN_SETTINGS.categories,
			purity: parsed.purity ?? DEFAULT_WALLHAVEN_SETTINGS.purity,
			sorting: parsed.sorting ?? DEFAULT_WALLHAVEN_SETTINGS.sorting,
			order: parsed.order ?? DEFAULT_WALLHAVEN_SETTINGS.order,
			topRange: parsed.topRange ?? DEFAULT_WALLHAVEN_SETTINGS.topRange,
			ratios: parsed.ratios ?? DEFAULT_WALLHAVEN_SETTINGS.ratios,
			apikey: parsed.apikey ?? DEFAULT_WALLHAVEN_SETTINGS.apikey
		};
	}
	return DEFAULT_WALLHAVEN_SETTINGS;
}

function saveWallhavenSettings(settings: WallhavenSettings) {
	if (!browser) return;
	localStorage.setItem(WALLHAVEN_SETTINGS_KEY, JSON.stringify(settings));
}

function loadApplyPaletteSettings(): ApplyPaletteSettings {
	if (!browser) return DEFAULT_APPLY_PALETTE_SETTINGS;
	const stored = localStorage.getItem(APPLY_PALETTE_SETTINGS_KEY);
	if (stored) {
		const parsed = JSON.parse(stored);
		return {
			luminosity: parsed.luminosity ?? DEFAULT_APPLY_PALETTE_SETTINGS.luminosity,
			nearest: parsed.nearest ?? DEFAULT_APPLY_PALETTE_SETTINGS.nearest,
			power: parsed.power ?? DEFAULT_APPLY_PALETTE_SETTINGS.power,
			maxDistance: parsed.maxDistance ?? DEFAULT_APPLY_PALETTE_SETTINGS.maxDistance
		};
	}
	return DEFAULT_APPLY_PALETTE_SETTINGS;
}

function saveApplyPaletteSettings(settings: ApplyPaletteSettings) {
	if (!browser) return;
	localStorage.setItem(APPLY_PALETTE_SETTINGS_KEY, JSON.stringify(settings));
}

function createAppStore() {
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
			{ id: UI.DEFAULT_SELECTOR_ID, color: 'oklch(79.2% 0.209 151.711)', selected: true },
			{ id: 'red', color: 'oklch(64.5% 0.246 16.439)', selected: false },
			{ id: 'blue', color: 'oklch(71.5% 0.143 215.221)', selected: false }
		],
		activeSelectorId: UI.DEFAULT_SELECTOR_ID,
		newFilterColor: '',
		savedPalettes: [],
		savedWorkspaces: [],
		sortMethod: 'none',
		applyPaletteSettings: loadApplyPaletteSettings(),
		themeExport: {
			themeResult: null,
			themeColorsWithUsage: [],
			themeName: 'Generated Theme',
			lastGeneratedPaletteVersion: 0,
			editorType: loadThemeExportPreferences(),
			saveOnCopy: loadThemeExportSaveOnCopy(),
			loadedThemeOverridesReference: null
		},
		savedThemes: loadSavedThemes(),
		paletteVersion: 0
	});

	let lastPaletteSignature = '';

	$effect.root(() => {
		$effect(() => {
			const signature = state.colors
				.map((color) => color.hex.toUpperCase())
				.sort()
				.join('|');
			if (signature === lastPaletteSignature) return;
			lastPaletteSignature = signature;
			state.paletteVersion += 1;
		});

		$effect(() => {
			const settings = state.wallhavenSettings;
			if (settings) {
				saveWallhavenSettings(settings);
			}
		});

		$effect(() => {
			const settings = state.applyPaletteSettings;
			if (settings) {
				saveApplyPaletteSettings(settings);
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

	function getCleanImageDataUrl(): string | null {
		if (!state.canvas || !state.image) return null;

		const tempCanvas = document.createElement('canvas');
		tempCanvas.width = state.canvas.width;
		tempCanvas.height = state.canvas.height;
		const tempCtx = tempCanvas.getContext('2d');
		if (!tempCtx) return null;

		tempCtx.drawImage(state.image, 0, 0, state.canvas.width, state.canvas.height);
		return tempCanvas.toDataURL(IMAGE.OUTPUT_FORMAT);
	}

	return {
		get state() {
			return state;
		},

		setThemeExportEditorType(editorType: EditorThemeType) {
			state.themeExport.editorType = editorType;
			saveThemeExportPreferences(editorType, state.themeExport.saveOnCopy);
		},

		setThemeExportSaveOnCopy(saveOnCopy: boolean) {
			state.themeExport.saveOnCopy = saveOnCopy;
			saveThemeExportPreferences(state.themeExport.editorType, saveOnCopy);
		},

		async drawToCanvas(file: File) {
			const reader = new FileReader();
			reader.onload = () => {
				state.image = new Image();
				state.image.onload = () => {
					if (!state.canvas || !state.image) return;

					state.canvasContext = state.canvas.getContext('2d')!;
					state.originalImageWidth = state.image.width;
					state.originalImageHeight = state.image.height;

					const dimensions = calculateImageDimensions(state.originalImageWidth, state.originalImageHeight);

					state.canvasScaleX = dimensions.scaleX;
					state.canvasScaleY = dimensions.scaleY;
					state.canvas.width = dimensions.width;
					state.canvas.height = dimensions.height;
					state.canvas.style.width = dimensions.width + 'px';
					state.canvas.style.height = dimensions.height + 'px';

					state.canvasContext.drawImage(state.image, 0, 0, dimensions.width, dimensions.height);
					state.imageLoaded = true;
					drawImageAndBoxes();
				};
				state.image.src = reader.result as string;
			};
			reader.readAsDataURL(file);
		},

		async drawBlobToCanvas(blob: Blob) {
			return new Promise<void>((resolve, reject) => {
				const url = URL.createObjectURL(blob);
				const newImage = new Image();

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
						drawImageAndBoxes();
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
			this.state.activeSelectorId = UI.DEFAULT_SELECTOR_ID;

			this.state.selectors.forEach((selector) => {
				selector.selection = undefined;
				selector.selected = selector.id === UI.DEFAULT_SELECTOR_ID;
			});

			this.redrawCanvas();
		},

		redrawCanvas() {
			drawImageAndBoxes();
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
			drawImageAndBoxes();
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

			await appStore.drawToCanvas(file);
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
				await appStore.drawToCanvas(files[0]);
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

		async savePalette() {
			if (!state.colors || state.colors.length === 0) {
				toast.error('Extract a palette before saving.');
				return;
			}
			const paletteName = prompt('Enter a name for your palette:');
			if (!paletteName) return;

			try {
				if (authStore.state.isAuthenticated && !authStore.isDemoUser()) {
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

				tutorialStore.setCurrentPaletteSaved(true);
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

			const toastId = toast.loading('Applying palette...');

			try {
				if (!state.canvas) return;

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
				if (authStore.state.isAuthenticated && !authStore.isDemoUser()) {
					const response = await paletteApi.getPalettes();
					state.savedPalettes = response.palettes;
				} else if (authStore.state.isAuthenticated && authStore.isDemoUser()) {
					const response = await paletteApi.getPalettes();
					const serverPalettes = response.palettes;

					const stored = localStorage.getItem('savedPalettes');
					const localPalettes = stored ? (JSON.parse(stored) as PaletteData[]) : [];

					state.savedPalettes = [...localPalettes, ...serverPalettes];
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
				const isLocalPalette = paletteId.startsWith('local_');

				if (authStore.state.isAuthenticated && !authStore.isDemoUser() && !isLocalPalette) {
					await paletteApi.deletePalette(paletteId);
					toast.success('Palette deleted');
					await appStore.loadSavedPalettes();
				} else if (authStore.isDemoUser() && !isLocalPalette) {
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

		async syncPalettesOnAuth() {
			if (!browser) return;

			if (authStore.state.isAuthenticated && !authStore.isDemoUser()) {
				const stored = localStorage.getItem('savedPalettes');
				if (stored) {
					try {
						const localPalettes = JSON.parse(stored) as PaletteData[];
						const toastId = toast.loading('Syncing your palettes...');
						try {
							await Promise.all(
								localPalettes.map(async (palette) => {
									try {
										await paletteApi.savePalette(palette.name, palette.palette);
									} catch (err) {
										console.error('Failed to sync palette:', palette.name, err);
									}
								})
							);
							localStorage.removeItem('savedPalettes');
							toast.success('Palettes synced successfully', { id: toastId });
						} catch {
							toast.error('Some palettes could not be synced. Try again later.', { id: toastId });
						}
					} catch {
						console.error('Failed to parse local palettes');
					}
				}
				await appStore.loadSavedPalettes();
			} else if (authStore.state.isAuthenticated && authStore.isDemoUser()) {
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
				const imageDataUrl = getCleanImageDataUrl();
				if (!imageDataUrl) {
					toast.error('Could not prepare the download. Please try again.', { id: toastId });
					return;
				}

				const blob = await fetch(imageDataUrl).then((res) => res.blob());

				const url = URL.createObjectURL(blob);
				const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, -5);
				const filename = `palette-applied-${timestamp}.png`;

				const link = document.createElement('a');
				link.href = url;
				link.download = filename;
				link.click();

				URL.revokeObjectURL(url);
				toast.success('Image downloaded', { id: toastId });
			} catch {
				toast.error('Download failed. Please try again.', { id: toastId });
			}
		},

		async saveWorkspace() {
			if (!state.canvas || !state.imageLoaded || !state.image) {
				toast.error('Load an image before saving a workspace.');
				return;
			}

			const workspaceName = prompt('Enter a name for your workspace:');
			if (!workspaceName) return;

			const toastId = toast.loading('Saving workspace...');

			try {
				const imageDataUrl = getCleanImageDataUrl();
				if (!imageDataUrl) {
					toast.error('Could not prepare the workspace image. Please try again.', { id: toastId });
					return;
				}

				if (authStore.state.isAuthenticated && !authStore.isDemoUser()) {
					await workspaceApi.saveWorkspace(workspaceName, imageDataUrl, {
						colors: state.colors,
						selectors: state.selectors,
						activeSelectorId: state.activeSelectorId,
						luminosity: state.applyPaletteSettings.luminosity,
						nearest: state.applyPaletteSettings.nearest,
						power: state.applyPaletteSettings.power,
						maxDistance: state.applyPaletteSettings.maxDistance,
						wallhavenSettings: state.wallhavenSettings
					});
					toast.success('Workspace saved: ' + workspaceName, { id: toastId });
					await appStore.loadSavedWorkspaces();
				} else {
					if (browser) {
						const newWorkspace: WorkspaceData = {
							id: `local_${Date.now()}`,
							name: workspaceName,
							imageData: imageDataUrl,
							colors: state.colors,
							selectors: state.selectors,
							activeSelectorId: state.activeSelectorId,
							luminosity: state.applyPaletteSettings.luminosity,
							nearest: state.applyPaletteSettings.nearest,
							power: state.applyPaletteSettings.power,
							maxDistance: state.applyPaletteSettings.maxDistance,
							wallhavenSettings: state.wallhavenSettings,
							createdAt: new Date().toISOString()
						};

						const stored = localStorage.getItem('savedWorkspaces');
						const workspaces = stored ? JSON.parse(stored) : [];
						workspaces.push(newWorkspace);
						localStorage.setItem('savedWorkspaces', JSON.stringify(workspaces));

						await appStore.loadSavedWorkspaces();
						toast.success('Workspace saved: ' + workspaceName, { id: toastId });
					}
				}
			} catch {
				toast.error('Could not save the workspace. Please try again.', { id: toastId });
			}
		},

		async loadWorkspace(workspace: WorkspaceData) {
			const toastId = toast.loading('Loading workspace...');

			try {
				const img = new Image();
				img.onload = async () => {
					state.image = img;
					if (!state.canvas) {
						toast.error('Canvas is not available yet. Try again in a moment.', { id: toastId });
						return;
					}

					state.canvasContext = state.canvas.getContext('2d')!;
					state.originalImageWidth = img.width;
					state.originalImageHeight = img.height;

					const dimensions = calculateImageDimensions(state.originalImageWidth, state.originalImageHeight);

					state.canvasScaleX = dimensions.scaleX;
					state.canvasScaleY = dimensions.scaleY;
					state.canvas.width = dimensions.width;
					state.canvas.height = dimensions.height;
					state.canvas.style.width = dimensions.width + 'px';
					state.canvas.style.height = dimensions.height + 'px';

					state.canvasContext.drawImage(img, 0, 0, dimensions.width, dimensions.height);
					state.imageLoaded = true;

					state.colors = workspace.colors || [];
					state.selectors = workspace.selectors || [];
					state.activeSelectorId = workspace.activeSelectorId || UI.DEFAULT_SELECTOR_ID;
					state.applyPaletteSettings = {
						luminosity: workspace.luminosity || 1,
						nearest: workspace.nearest || 30,
						power: workspace.power || 4,
						maxDistance: workspace.maxDistance || 0
					};
					state.wallhavenSettings = workspace.wallhavenSettings || {
						categories: '111',
						purity: '100',
						sorting: 'relevance',
						order: 'desc',
						topRange: '1M',
						ratios: [],
						apikey: ''
					};

					drawImageAndBoxes();
					toast.success('Workspace loaded: ' + workspace.name, { id: toastId });
				};

				img.onerror = () => {
					toast.error('Could not load the workspace image.', { id: toastId });
				};

				img.src = workspace.imageData;
			} catch {
				toast.error('Could not load the workspace. Please try again.', { id: toastId });
			}
		},

		async loadSavedWorkspaces() {
			if (!browser) return;

			try {
				if (authStore.state.isAuthenticated && !authStore.isDemoUser()) {
					const response = await workspaceApi.getWorkspaces();
					state.savedWorkspaces = response.workspaces.map((w) => ({
						...w,
						colors: w.colors || [],
						selectors: w.selectors || [],
						activeSelectorId: w.activeSelectorId || UI.DEFAULT_SELECTOR_ID,
						luminosity: w.luminosity || 1,
						nearest: w.nearest || 30,
						power: w.power || 4,
						maxDistance: w.maxDistance || 0,
						wallhavenSettings: w.wallhavenSettings || {
							categories: '111',
							purity: '100',
							sorting: 'relevance',
							order: 'desc',
							topRange: '1M',
							ratios: [],
							apikey: ''
						}
					}));
				} else if (authStore.state.isAuthenticated && authStore.isDemoUser()) {
					const response = await workspaceApi.getWorkspaces();
					const serverWorkspaces = response.workspaces.map((w) => ({
						...w,
						colors: w.colors || [],
						selectors: w.selectors || [],
						activeSelectorId: w.activeSelectorId || UI.DEFAULT_SELECTOR_ID,
						luminosity: w.luminosity || 1,
						nearest: w.nearest || 30,
						power: w.power || 4,
						maxDistance: w.maxDistance || 0,
						wallhavenSettings: w.wallhavenSettings || {
							categories: '111',
							purity: '100',
							sorting: 'relevance',
							order: 'desc',
							topRange: '1M',
							ratios: [],
							apikey: ''
						}
					}));

					const stored = localStorage.getItem('savedWorkspaces');
					const localWorkspaces = stored ? (JSON.parse(stored) as WorkspaceData[]) : [];

					state.savedWorkspaces = [...localWorkspaces, ...serverWorkspaces];
				} else {
					const stored = localStorage.getItem('savedWorkspaces');
					if (stored) {
						try {
							const localWorkspaces = JSON.parse(stored) as WorkspaceData[];
							state.savedWorkspaces = localWorkspaces;
						} catch {
							state.savedWorkspaces = [];
						}
					} else {
						state.savedWorkspaces = [];
					}
				}
			} catch (error) {
				console.error('Failed to load saved workspaces:', error);
				state.savedWorkspaces = [];
			}
		},

		async deleteWorkspace(workspaceId: string) {
			try {
				const isLocalWorkspace = workspaceId.startsWith('local_');

				if (authStore.state.isAuthenticated && !authStore.isDemoUser() && !isLocalWorkspace) {
					await workspaceApi.deleteWorkspace(workspaceId);
					toast.success('Workspace deleted');
					await appStore.loadSavedWorkspaces();
				} else if (authStore.isDemoUser() && !isLocalWorkspace) {
					await workspaceApi.deleteWorkspace(workspaceId);
					toast.success('Workspace deleted');
					await appStore.loadSavedWorkspaces();
				} else {
					if (browser) {
						const stored = localStorage.getItem('savedWorkspaces');
						const workspaces = stored ? JSON.parse(stored) : [];
						const filtered = workspaces.filter((w: WorkspaceData) => w.id !== workspaceId);
						localStorage.setItem('savedWorkspaces', JSON.stringify(filtered));
					}
					await appStore.loadSavedWorkspaces();
					toast.success('Workspace deleted');
				}
			} catch {
				toast.error('Could not delete the workspace. Please try again.');
			}
		},

		async syncWorkspacesOnAuth() {
			if (!browser) return;

			if (authStore.state.isAuthenticated && !authStore.isDemoUser()) {
				const stored = localStorage.getItem('savedWorkspaces');
				if (stored) {
					try {
						const localWorkspaces = JSON.parse(stored) as WorkspaceData[];
						const toastId = toast.loading('Syncing your workspaces...');
						try {
							await Promise.all(
								localWorkspaces.map(async (workspace) => {
									try {
										await workspaceApi.saveWorkspace(workspace.name, workspace.imageData, {
											colors: workspace.colors || [],
											selectors: workspace.selectors || [],
											activeSelectorId: workspace.activeSelectorId || UI.DEFAULT_SELECTOR_ID,
											luminosity: workspace.luminosity || 1,
											nearest: workspace.nearest || 30,
											power: workspace.power || 4,
											maxDistance: workspace.maxDistance || 0,
											wallhavenSettings: workspace.wallhavenSettings || {
												categories: '111',
												purity: '100',
												sorting: 'relevance',
												order: 'desc',
												topRange: '1M',
												ratios: [],
												apikey: ''
											}
										});
									} catch (err) {
										console.error('Failed to sync workspace:', workspace.name, err);
									}
								})
							);
							localStorage.removeItem('savedWorkspaces');
							toast.success('Workspaces synced successfully', { id: toastId });
						} catch {
							toast.error('Some workspaces could not be synced. Try again later.', { id: toastId });
						}
					} catch {
						console.error('Failed to parse local workspaces');
					}
				}
				await appStore.loadSavedWorkspaces();
			} else if (authStore.state.isAuthenticated && authStore.isDemoUser()) {
				await appStore.loadSavedWorkspaces();
			} else {
				if (state.savedWorkspaces.length > 0) {
					localStorage.setItem('savedWorkspaces', JSON.stringify(state.savedWorkspaces));
				}
			}
		},

		saveThemeToLocal(theme: SavedThemeItem) {
			state.savedThemes = [theme, ...state.savedThemes];
			saveSavedThemes(state.savedThemes);
		},

		replaceSavedTheme(themeId: string, theme: SavedThemeItem) {
			const index = state.savedThemes.findIndex((item) => item.id === themeId);
			if (index === -1) {
				state.savedThemes = [theme, ...state.savedThemes];
			} else {
				const next = [...state.savedThemes];
				next[index] = theme;
				state.savedThemes = next;
			}
			saveSavedThemes(state.savedThemes);
		},

		deleteTheme(themeId: string) {
			state.savedThemes = state.savedThemes.filter((item) => item.id !== themeId);
			saveSavedThemes(state.savedThemes);
		}
	};
}

export const appStore = createAppStore();
