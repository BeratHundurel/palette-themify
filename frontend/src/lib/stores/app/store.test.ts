import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { IMAGE } from '$lib/types/image';
import { DEFAULT_THEME_EXPORT_PREFERENCES } from '$lib/types/theme';
import type { SavedThemeItem } from '$lib/types/theme';

const { authStoreMock, tutorialStoreMock, dialogStoreMock } = vi.hoisted(() => ({
	authStoreMock: {
		state: {
			user: null,
			isAuthenticated: false,
			isLoading: false
		}
	},
	tutorialStoreMock: {
		setCurrentPaletteSaved: vi.fn()
	},
	dialogStoreMock: {
		prompt: vi.fn(),
		confirm: vi.fn(),
		resolveConfirm: vi.fn(),
		resolvePrompt: vi.fn(),
		close: vi.fn(),
		state: {
			confirm: null,
			prompt: null
		}
	}
}));

vi.mock('$app/environment', () => ({
	browser: true
}));

vi.mock('../auth.svelte', () => ({
	authStore: authStoreMock
}));

vi.mock('../tutorial.svelte', () => ({
	tutorialStore: tutorialStoreMock
}));

vi.mock('$lib/stores/dialog.svelte', () => ({
	dialogStore: dialogStoreMock
}));

vi.mock('$lib/api/palette', () => ({
	extractPalette: vi.fn(),
	savePalette: vi.fn(),
	getPalettes: vi.fn(),
	deletePalette: vi.fn()
}));

vi.mock('$lib/api/theme', () => ({
	applyPaletteBlob: vi.fn(),
	generateTheme: vi.fn(),
	generateOverridable: vi.fn()
}));

vi.mock('$lib/api/preferences', () => ({
	getPreferences: vi.fn(),
	savePreferences: vi.fn()
}));

vi.mock('$lib/api/themes', () => ({
	getThemes: vi.fn(),
	saveTheme: vi.fn(),
	updateTheme: vi.fn(),
	deleteTheme: vi.fn()
}));

vi.mock('$lib/api/wallhaven', () => ({
	downloadImage: vi.fn()
}));

vi.mock('svelte-french-toast', () => ({
	default: {
		loading: vi.fn(() => 'toast-id'),
		success: vi.fn(),
		error: vi.fn()
	}
}));

import * as paletteApi from '$lib/api/palette';
import * as preferencesApi from '$lib/api/preferences';
import * as themeApi from '$lib/api/theme';
import * as themesApi from '$lib/api/themes';
import { appStore } from '$lib/stores/app/store.svelte';
import toast from 'svelte-french-toast';

type ThemeResultLike = {
	theme: Record<string, unknown>;
	themeOverrides: Record<string, unknown>;
	rawThemeOverrides: Record<string, unknown>;
	colors: Array<{ hex: string }>;
	boostCoefficient: number;
};

function dirtyThemeExportState() {
	const dirtyResult: ThemeResultLike = {
		theme: { name: 'dirty' },
		themeOverrides: { background: '#000000' },
		rawThemeOverrides: { foreground: '#ffffff' },
		colors: [{ hex: '#111111' }],
		boostCoefficient: 2
	};

	appStore.state.themeExport.themeResult = dirtyResult as never;
	appStore.state.themeExport.themeName = 'Dirty Theme';
	appStore.state.themeExport.lastGeneratedPaletteVersion = 17;
	appStore.state.themeExport.themeVersions = { dirty: dirtyResult as never };
	appStore.state.themeExport.rawThemeOverrides = { background: '#101010' };
	appStore.state.themeExport.hasManualBackgroundOverride = true;
	appStore.state.themeExport.hasManualForegroundOverride = true;
	appStore.state.themeExport.loadedThemeOverridesReference = { background: '#101010' };
	appStore.state.themeExport.backupColors = [{ hex: '#123456' }];
}

function resetStoreForTest() {
	appStore.state.colors = [];
	appStore.state.canvas = null;
	appStore.state.canvasContext = null;
	appStore.state.image = null;
	appStore.state.imageLoaded = false;
	appStore.state.isExtracting = false;
	appStore.state.savedPalettes = [];
	appStore.state.savedThemes = [];

	appStore.state.themeExport.themeResult = null;
	appStore.state.themeExport.themeName = 'Generated Theme';
	appStore.state.themeExport.lastGeneratedPaletteVersion = 0;
	appStore.state.themeExport.themeVersions = {};
	appStore.state.themeExport.rawThemeOverrides = {};
	appStore.state.themeExport.hasManualBackgroundOverride = false;
	appStore.state.themeExport.hasManualForegroundOverride = false;
	appStore.state.themeExport.loadedThemeOverridesReference = null;
	appStore.state.themeExport.backupColors = null;
	appStore.state.themeExport.editorType = DEFAULT_THEME_EXPORT_PREFERENCES.editorType;
	appStore.state.themeExport.appearance = DEFAULT_THEME_EXPORT_PREFERENCES.appearance;
	appStore.state.themeExport.saveOnCopy = DEFAULT_THEME_EXPORT_PREFERENCES.saveOnCopy;
	appStore.state.themeExport.boostCoefficient = DEFAULT_THEME_EXPORT_PREFERENCES.boostCoefficient;
}

function makeCanvasReturning(blob: Blob): HTMLCanvasElement {
	const canvas = document.createElement('canvas');
	Object.defineProperty(canvas, 'toBlob', {
		configurable: true,
		value: (callback: BlobCallback) => callback(blob)
	});
	return canvas;
}

function markImageLoaded() {
	appStore.state.imageLoaded = true;
	appStore.state.image = {} as HTMLImageElement;
}

function makePalette(id: string, name: string) {
	return {
		id,
		name,
		palette: [{ hex: '#112233' }],
		createdAt: '2026-01-01T00:00:00.000Z'
	};
}

function makeTheme(id: string, name = `Theme ${id}`): SavedThemeItem {
	return {
		id,
		name,
		editorType: 'vscode' as const,
		createdAt: '2026-01-01T00:00:00.000Z',
		themeResult: {
			theme: { name } as unknown as SavedThemeItem['themeResult']['theme'],
			themeOverrides: {},
			rawThemeOverrides: {},
			colors: [{ hex: '#112233' }],
			boostCoefficient: 1
		}
	};
}

describe('appStore', () => {
	beforeEach(() => {
		vi.restoreAllMocks();
		vi.clearAllMocks();
		dialogStoreMock.prompt.mockResolvedValue('Saved Palette');
		authStoreMock.state.isAuthenticated = false;
		resetStoreForTest();
		localStorage.clear();
	});

	afterEach(() => {
		vi.useRealTimers();
	});

	describe('extractPalette', () => {
		it('shows a validation error when file is missing', async () => {
			await appStore.extractPalette(null as unknown as Blob);

			expect(toast.error).toHaveBeenCalledWith('Choose an image to extract colors.');
			expect(paletteApi.extractPalette).not.toHaveBeenCalled();
		});

		it('blocks oversized files before API call', async () => {
			const oversizedFile = { size: IMAGE.MAX_FILE_SIZE + 1 } as Blob;

			await appStore.extractPalette(oversizedFile);

			expect(toast.error).toHaveBeenCalledWith('File is too large. Choose a file under 50MB.');
			expect(paletteApi.extractPalette).not.toHaveBeenCalled();
		});

		it('sets colors, resets theme-export session and resolves toast on success', async () => {
			dirtyThemeExportState();
			vi.mocked(paletteApi.extractPalette).mockResolvedValue({
				palette: [{ hex: '#AABBCC' }, { hex: '#112233' }]
			});

			await appStore.extractPalette(new Blob(['img'], { type: 'image/png' }));

			expect(appStore.state.colors).toEqual([{ hex: '#AABBCC' }, { hex: '#112233' }]);
			expect(appStore.state.themeExport.themeResult).toBeNull();
			expect(appStore.state.themeExport.themeName).toBe('Generated Theme');
			expect(appStore.state.themeExport.lastGeneratedPaletteVersion).toBe(0);
			expect(appStore.state.themeExport.themeVersions).toEqual({});
			expect(appStore.state.themeExport.rawThemeOverrides).toEqual({});
			expect(appStore.state.themeExport.hasManualBackgroundOverride).toBe(false);
			expect(appStore.state.themeExport.hasManualForegroundOverride).toBe(false);
			expect(appStore.state.themeExport.loadedThemeOverridesReference).toBeNull();
			expect(appStore.state.themeExport.backupColors).toBeNull();
			expect(toast.success).toHaveBeenCalledWith('Palette extracted', { id: 'toast-id' });
			expect(appStore.state.isExtracting).toBe(false);
		});

		it('reuses existing toast id when provided', async () => {
			vi.mocked(paletteApi.extractPalette).mockResolvedValue({
				palette: [{ hex: '#AABBCC' }]
			});

			await appStore.extractPalette(new Blob(['img'], { type: 'image/png' }), 'existing-toast');

			expect(toast.loading).not.toHaveBeenCalled();
			expect(toast.success).toHaveBeenCalledWith('Palette extracted', { id: 'existing-toast' });
		});

		it('shows a no-colors error when API returns an empty palette', async () => {
			vi.mocked(paletteApi.extractPalette).mockResolvedValue({ palette: [] });

			await appStore.extractPalette(new Blob(['img'], { type: 'image/png' }));

			expect(toast.error).toHaveBeenCalledWith('No colors found in the selected area. Try a larger selection.', {
				id: 'toast-id'
			});
			expect(toast.success).not.toHaveBeenCalled();
			expect(appStore.state.isExtracting).toBe(false);
		});

		it('shows a generic extraction error when API throws', async () => {
			vi.mocked(paletteApi.extractPalette).mockRejectedValue(new Error('network down'));

			await appStore.extractPalette(new Blob(['img'], { type: 'image/png' }));

			expect(toast.error).toHaveBeenCalledWith('Could not extract a palette. Please try again.', {
				id: 'toast-id'
			});
			expect(appStore.state.isExtracting).toBe(false);
		});
	});

	describe('applyPalette', () => {
		it('rejects empty palette input', async () => {
			await appStore.applyPalette([]);

			expect(toast.error).toHaveBeenCalledWith('Select a palette to apply.');
			expect(themeApi.applyPaletteBlob).not.toHaveBeenCalled();
		});

		it('rejects palettes containing invalid hex values', async () => {
			await appStore.applyPalette([{ hex: '#112233' }, { hex: '#GGGGGG' }]);

			expect(toast.error).toHaveBeenCalledWith('This palette has invalid colors. Please reselect or recreate it.');
			expect(themeApi.applyPaletteBlob).not.toHaveBeenCalled();
		});

		it('handles missing canvas before attempting API calls', async () => {
			markImageLoaded();
			appStore.state.canvas = null;

			await appStore.applyPalette([{ hex: '#112233' }]);

			expect(toast.error).toHaveBeenCalledWith('Load an image before applying a palette.');
			expect(toast.loading).not.toHaveBeenCalled();
			expect(themeApi.applyPaletteBlob).not.toHaveBeenCalled();
		});

		it('applies palette and refreshes extracted colors on success', async () => {
			const srcBlob = new Blob(['src'], { type: 'image/png' });
			const outBlob = new Blob(['out'], { type: 'image/png' });
			markImageLoaded();
			appStore.state.canvas = makeCanvasReturning(srcBlob);
			appStore.state.applyPaletteSettings = {
				luminosity: 1.7,
				nearest: 9,
				power: 3,
				maxDistance: 45
			};
			dirtyThemeExportState();

			vi.spyOn(appStore, 'drawBlobToCanvas').mockResolvedValue();
			vi.mocked(themeApi.applyPaletteBlob).mockResolvedValue(outBlob);
			vi.mocked(paletteApi.extractPalette).mockResolvedValue({ palette: [{ hex: '#00AA00' }] });

			await appStore.applyPalette([{ hex: '#112233' }, { hex: '#445566' }]);

			expect(themeApi.applyPaletteBlob).toHaveBeenCalledWith(srcBlob, [{ hex: '#112233' }, { hex: '#445566' }], {
				luminosity: 1.7,
				nearest: 9,
				power: 3,
				maxDistance: 45
			});
			expect(appStore.drawBlobToCanvas).toHaveBeenCalledWith(outBlob);
			expect(appStore.state.colors).toEqual([{ hex: '#00AA00' }]);
			expect(appStore.state.themeExport.themeResult).toBeNull();
			expect(toast.success).toHaveBeenCalledWith('Applied palette', { id: 'toast-id' });
		});

		it('shows specific error when applied image yields no colors', async () => {
			const srcBlob = new Blob(['src'], { type: 'image/png' });
			const outBlob = new Blob(['out'], { type: 'image/png' });
			markImageLoaded();
			appStore.state.canvas = makeCanvasReturning(srcBlob);
			vi.spyOn(appStore, 'drawBlobToCanvas').mockResolvedValue();
			vi.mocked(themeApi.applyPaletteBlob).mockResolvedValue(outBlob);
			vi.mocked(paletteApi.extractPalette).mockResolvedValue({ palette: [] });

			await appStore.applyPalette([{ hex: '#112233' }]);

			expect(toast.error).toHaveBeenCalledWith('Palette applied, but no colors were detected. Try another image.', {
				id: 'toast-id'
			});
			expect(toast.success).not.toHaveBeenCalled();
		});

		it('shows generic error when apply API fails', async () => {
			const srcBlob = new Blob(['src'], { type: 'image/png' });
			markImageLoaded();
			appStore.state.canvas = makeCanvasReturning(srcBlob);
			vi.mocked(themeApi.applyPaletteBlob).mockRejectedValue(new Error('apply failed'));

			await appStore.applyPalette([{ hex: '#112233' }]);

			expect(toast.error).toHaveBeenCalledWith('Could not apply the palette. Please try again.', {
				id: 'toast-id'
			});
		});
	});

	describe('preferences sync', () => {
		it('persists preferences locally for unauthenticated users only', () => {
			authStoreMock.state.isAuthenticated = false;
			appStore.state.themeExport.appearance = 'light';

			appStore.persistPreferencesLocal();

			const raw = localStorage.getItem('appPreferences');
			expect(raw).toBeTruthy();
			expect(JSON.parse(raw as string).themeExport.appearance).toBe('light');
			expect(preferencesApi.savePreferences).not.toHaveBeenCalled();
		});

		it('debounces authenticated preference sync and keeps latest values', async () => {
			vi.useFakeTimers();
			authStoreMock.state.isAuthenticated = true;
			vi.mocked(preferencesApi.savePreferences).mockResolvedValue({ preferences: {} });

			appStore.state.themeExport.appearance = 'dark';
			appStore.persistPreferencesLocal();

			appStore.state.themeExport.appearance = 'light';
			appStore.persistPreferencesLocal();

			await vi.advanceTimersByTimeAsync(299);
			expect(preferencesApi.savePreferences).not.toHaveBeenCalled();

			await vi.advanceTimersByTimeAsync(2);
			await Promise.resolve();

			expect(preferencesApi.savePreferences).toHaveBeenCalledTimes(1);
			expect(preferencesApi.savePreferences).toHaveBeenCalledWith(
				expect.objectContaining({
					themeExport: expect.objectContaining({ appearance: 'light' })
				})
			);
			expect(localStorage.getItem('appPreferences')).toBeNull();
		});

		it('prefers remote non-default preferences over local cache', async () => {
			authStoreMock.state.isAuthenticated = true;
			vi.mocked(preferencesApi.getPreferences).mockResolvedValue({
				preferences: {
					wallhavenSettings: {
						categories: '010',
						purity: '111',
						sorting: 'favorites',
						order: 'asc',
						topRange: '6M',
						ratios: ['16x9'],
						apikey: 'remote-key'
					}
				}
			});
			localStorage.setItem(
				'appPreferences',
				JSON.stringify({
					wallhavenSettings: {
						sorting: 'date_added',
						order: 'desc'
					}
				})
			);

			await appStore.syncPreferencesOnAuth();

			expect(appStore.state.wallhavenSettings.sorting).toBe('favorites');
			expect(appStore.state.wallhavenSettings.order).toBe('asc');
			expect(preferencesApi.savePreferences).not.toHaveBeenCalled();

			const persisted = JSON.parse(localStorage.getItem('appPreferences') as string);
			expect(persisted.wallhavenSettings.sorting).toBe('favorites');
			expect(persisted.wallhavenSettings.order).toBe('asc');
			expect(persisted.wallhavenSettings.sorting).not.toBe('date_added');
		});

		it('merges and uploads local preferences when server has defaults', async () => {
			authStoreMock.state.isAuthenticated = true;
			vi.mocked(preferencesApi.getPreferences).mockResolvedValue({ preferences: {} });
			vi.mocked(preferencesApi.savePreferences).mockResolvedValue({ preferences: {} });
			localStorage.setItem(
				'appPreferences',
				JSON.stringify({
					applyPaletteSettings: { nearest: 14 },
					themeExport: { appearance: 'light', saveOnCopy: false },
					wallhavenSettings: { sorting: 'toplist', topRange: '1y' }
				})
			);

			await appStore.syncPreferencesOnAuth();

			expect(preferencesApi.savePreferences).toHaveBeenCalledTimes(1);
			expect(preferencesApi.savePreferences).toHaveBeenCalledWith(
				expect.objectContaining({
					applyPaletteSettings: expect.objectContaining({ nearest: 14 }),
					themeExport: expect.objectContaining({ appearance: 'light', saveOnCopy: false }),
					wallhavenSettings: expect.objectContaining({ sorting: 'toplist', topRange: '1y' })
				})
			);
			expect(appStore.state.themeExport.appearance).toBe('light');
			expect(localStorage.getItem('appPreferences')).toBeNull();
		});

		it('stores preferences locally when user is not authenticated', async () => {
			authStoreMock.state.isAuthenticated = false;
			appStore.state.themeExport.saveOnCopy = false;

			await appStore.syncPreferencesOnAuth();

			const raw = localStorage.getItem('appPreferences');
			expect(raw).toBeTruthy();
			expect(JSON.parse(raw as string).themeExport.saveOnCopy).toBe(false);
		});
	});

	describe('palette persistence', () => {
		it('prevents saving palette when no colors exist', async () => {
			appStore.state.colors = [];

			await appStore.savePalette();

			expect(toast.error).toHaveBeenCalledWith('Extract a palette before saving.');
			expect(paletteApi.savePalette).not.toHaveBeenCalled();
		});

		it('saves palette to local storage for unauthenticated users', async () => {
			authStoreMock.state.isAuthenticated = false;
			appStore.state.colors = [{ hex: '#112233' }];
			dialogStoreMock.prompt.mockResolvedValue('Local Sunset');

			await appStore.savePalette();

			const stored = JSON.parse(localStorage.getItem('savedPalettes') as string);
			expect(stored).toHaveLength(1);
			expect(stored[0].id).toMatch(/^local_/);
			expect(stored[0].name).toBe('Local Sunset');
			expect(stored[0].palette).toEqual([{ hex: '#112233' }]);
			expect(toast.success).toHaveBeenCalledWith('Palette saved: Local Sunset');
			expect(tutorialStoreMock.setCurrentPaletteSaved).toHaveBeenCalledWith(true);
		});

		it('saves palette through API for authenticated users', async () => {
			authStoreMock.state.isAuthenticated = true;
			appStore.state.colors = [{ hex: '#112233' }];
			dialogStoreMock.prompt.mockResolvedValue('Remote Sunset');
			vi.mocked(paletteApi.savePalette).mockResolvedValue({ message: 'ok', name: 'Remote Sunset' });
			const loadSavedPalettesSpy = vi.spyOn(appStore, 'loadSavedPalettes').mockResolvedValue(undefined);

			await appStore.savePalette();

			expect(paletteApi.savePalette).toHaveBeenCalledWith('Remote Sunset', [{ hex: '#112233' }]);
			expect(loadSavedPalettesSpy).toHaveBeenCalledTimes(1);
			expect(toast.success).toHaveBeenCalledWith('Palette saved: Remote Sunset');
			expect(tutorialStoreMock.setCurrentPaletteSaved).toHaveBeenCalledWith(true);
		});

		it('loads local palettes and recovers from malformed cache', async () => {
			authStoreMock.state.isAuthenticated = false;
			localStorage.setItem('savedPalettes', JSON.stringify([makePalette('local_1', 'A')]));

			await appStore.loadSavedPalettes();
			expect(appStore.state.savedPalettes).toHaveLength(1);
			expect(appStore.state.savedPalettes[0].id).toBe('local_1');

			localStorage.setItem('savedPalettes', '{broken json');
			await appStore.loadSavedPalettes();
			expect(appStore.state.savedPalettes).toEqual([]);
		});

		it('deletes local palettes from storage for unauthenticated users', async () => {
			authStoreMock.state.isAuthenticated = false;
			localStorage.setItem('savedPalettes', JSON.stringify([makePalette('local_1', 'A'), makePalette('local_2', 'B')]));
			const loadSavedPalettesSpy = vi.spyOn(appStore, 'loadSavedPalettes').mockResolvedValue(undefined);

			await appStore.deletePalette('local_1');

			const persisted = JSON.parse(localStorage.getItem('savedPalettes') as string);
			expect(persisted).toHaveLength(1);
			expect(persisted[0].id).toBe('local_2');
			expect(loadSavedPalettesSpy).toHaveBeenCalledTimes(1);
			expect(toast.success).toHaveBeenCalledWith('Palette deleted');
		});

		it('syncs only local palettes to API on login', async () => {
			authStoreMock.state.isAuthenticated = true;
			localStorage.setItem(
				'savedPalettes',
				JSON.stringify([
					makePalette('local_1', 'Local One'),
					makePalette('remote_1', 'Remote Existing'),
					{ ...makePalette('', 'No Id Yet'), id: '' }
				])
			);
			vi.mocked(paletteApi.savePalette).mockResolvedValue({ message: 'ok', name: 'saved' });
			const loadSavedPalettesSpy = vi.spyOn(appStore, 'loadSavedPalettes').mockResolvedValue(undefined);

			await appStore.syncPalettesOnAuth();

			expect(paletteApi.savePalette).toHaveBeenCalledTimes(2);
			expect(paletteApi.savePalette).toHaveBeenNthCalledWith(1, 'Local One', [{ hex: '#112233' }]);
			expect(paletteApi.savePalette).toHaveBeenNthCalledWith(2, 'No Id Yet', [{ hex: '#112233' }]);
			expect(toast.loading).toHaveBeenCalledWith('Syncing your palettes...');
			expect(toast.success).toHaveBeenCalledWith('Palettes synced successfully', { id: 'toast-id' });
			expect(localStorage.getItem('savedPalettes')).toBeNull();
			expect(loadSavedPalettesSpy).toHaveBeenCalledTimes(1);
		});

		it('stores in-memory palettes locally when user logs out', async () => {
			authStoreMock.state.isAuthenticated = false;
			appStore.state.savedPalettes = [makePalette('remote_1', 'Remote One')];

			await appStore.syncPalettesOnAuth();

			const stored = JSON.parse(localStorage.getItem('savedPalettes') as string);
			expect(stored).toHaveLength(1);
			expect(stored[0].id).toBe('remote_1');
			expect(paletteApi.savePalette).not.toHaveBeenCalled();
		});

		it('loads palettes from API for authenticated users', async () => {
			authStoreMock.state.isAuthenticated = true;
			vi.mocked(paletteApi.getPalettes).mockResolvedValue({
				palettes: [makePalette('remote_1', 'Remote One')]
			});

			await appStore.loadSavedPalettes();

			expect(appStore.state.savedPalettes).toEqual([makePalette('remote_1', 'Remote One')]);
		});

		it('falls back to empty list when palette API fails', async () => {
			authStoreMock.state.isAuthenticated = true;
			vi.spyOn(console, 'error').mockImplementation(() => {});
			vi.mocked(paletteApi.getPalettes).mockRejectedValue(new Error('palette api down'));

			await appStore.loadSavedPalettes();

			expect(appStore.state.savedPalettes).toEqual([]);
			expect(console.error).toHaveBeenCalled();
		});

		it('deletes remote palette through API when authenticated', async () => {
			authStoreMock.state.isAuthenticated = true;
			vi.mocked(paletteApi.deletePalette).mockResolvedValue({ message: 'ok' });
			const loadSavedPalettesSpy = vi.spyOn(appStore, 'loadSavedPalettes').mockResolvedValue(undefined);

			await appStore.deletePalette('remote_99');

			expect(paletteApi.deletePalette).toHaveBeenCalledWith('remote_99');
			expect(loadSavedPalettesSpy).toHaveBeenCalledTimes(1);
			expect(toast.success).toHaveBeenCalledWith('Palette deleted');
		});

		it('shows user-facing error when palette deletion fails', async () => {
			authStoreMock.state.isAuthenticated = true;
			vi.mocked(paletteApi.deletePalette).mockRejectedValue(new Error('delete failed'));

			await appStore.deletePalette('remote_99');

			expect(toast.error).toHaveBeenCalledWith('Could not delete the palette. Please try again.');
		});
	});

	describe('theme persistence', () => {
		it('stores saved theme locally with computed signature', () => {
			const persistThemeChangeSpy = vi.spyOn(appStore, 'persistThemeChange').mockResolvedValue(undefined);
			const theme = makeTheme('theme-1', 'Local Theme');

			appStore.saveThemeToLocal(theme);

			expect(appStore.state.savedThemes).toHaveLength(1);
			expect(appStore.state.savedThemes[0].id).toBe('theme-1');
			expect(appStore.state.savedThemes[0].signature).toBe(JSON.stringify(theme.themeResult));
			const persisted = JSON.parse(localStorage.getItem('savedThemes') as string);
			expect(persisted[0].id).toBe('theme-1');
			expect(persistThemeChangeSpy).toHaveBeenCalledWith(
				expect.objectContaining({ id: 'theme-1', signature: JSON.stringify(theme.themeResult) }),
				'create'
			);
		});

		it('deduplicates applyThemeResponse using signature match', () => {
			const existing = {
				...makeTheme('theme-1', 'Existing'),
				signature: 'same-signature'
			};
			const incoming = {
				...makeTheme('theme-2', 'Incoming'),
				signature: 'same-signature'
			};
			appStore.state.savedThemes = [existing];

			appStore.applyThemeResponse(incoming);

			expect(appStore.state.savedThemes).toHaveLength(1);
			expect(appStore.state.savedThemes[0].id).toBe('theme-2');
			expect(appStore.state.savedThemes[0].name).toBe('Incoming');
		});

		it('skips remote theme persistence when user is unauthenticated', async () => {
			authStoreMock.state.isAuthenticated = false;

			await appStore.persistThemeChange(makeTheme('theme-1'), 'create');

			expect(themesApi.saveTheme).not.toHaveBeenCalled();
			expect(themesApi.updateTheme).not.toHaveBeenCalled();
			expect(themesApi.deleteTheme).not.toHaveBeenCalled();
		});

		it('routes create/update/delete theme persistence to API when authenticated', async () => {
			authStoreMock.state.isAuthenticated = true;
			const applyThemeResponseSpy = vi.spyOn(appStore, 'applyThemeResponse').mockImplementation(() => {});
			vi.mocked(themesApi.saveTheme).mockResolvedValue({ message: 'ok', theme: makeTheme('remote-create', 'Created') });
			vi.mocked(themesApi.updateTheme).mockResolvedValue({
				message: 'ok',
				theme: makeTheme('remote-update', 'Updated')
			});
			vi.mocked(themesApi.deleteTheme).mockResolvedValue({ message: 'ok' });

			await appStore.persistThemeChange(makeTheme('theme-local-create', 'Local Create'), 'create');
			await appStore.persistThemeChange(makeTheme('theme-local-update', 'Local Update'), 'update', 'theme-update-id');
			await appStore.persistThemeChange(null, 'delete', 'theme-delete-id');

			expect(themesApi.saveTheme).toHaveBeenCalledTimes(1);
			expect(themesApi.updateTheme).toHaveBeenCalledWith(
				'theme-update-id',
				expect.objectContaining({ id: 'theme-local-update', signature: expect.any(String) })
			);
			expect(themesApi.deleteTheme).toHaveBeenCalledWith('theme-delete-id');
			expect(applyThemeResponseSpy).toHaveBeenNthCalledWith(
				1,
				expect.objectContaining({ id: 'remote-create' }),
				'theme-local-create'
			);
			expect(applyThemeResponseSpy).toHaveBeenNthCalledWith(2, expect.objectContaining({ id: 'remote-update' }));
		});

		it('syncs saved themes from local storage to API and refreshes from server', async () => {
			authStoreMock.state.isAuthenticated = true;
			localStorage.setItem('savedThemes', JSON.stringify([makeTheme('local-theme-1', 'Local Theme')]));
			vi.mocked(themesApi.saveTheme).mockResolvedValue({
				message: 'ok',
				theme: makeTheme('remote-theme-1', 'Remote Theme')
			});
			vi.mocked(themesApi.getThemes).mockResolvedValue({ themes: [makeTheme('server-theme-1', 'Server Theme')] });

			await appStore.syncSavedThemesOnAuth();

			expect(themesApi.saveTheme).toHaveBeenCalledTimes(1);
			expect(themesApi.saveTheme).toHaveBeenCalledWith(
				expect.objectContaining({ id: 'local-theme-1', signature: expect.any(String) })
			);
			expect(themesApi.getThemes).toHaveBeenCalledTimes(1);
			expect(appStore.state.savedThemes).toEqual([makeTheme('server-theme-1', 'Server Theme')]);
			const persisted = JSON.parse(localStorage.getItem('savedThemes') as string);
			expect(persisted[0].id).toBe('server-theme-1');
		});

		it('moves in-memory themes to local storage when logging out', async () => {
			authStoreMock.state.isAuthenticated = false;
			appStore.state.savedThemes = [makeTheme('theme-memory-1', 'Memory Theme')];

			await appStore.syncSavedThemesOnAuth();

			const persisted = JSON.parse(localStorage.getItem('savedThemes') as string);
			expect(persisted).toHaveLength(1);
			expect(persisted[0].id).toBe('theme-memory-1');
			expect(themesApi.getThemes).not.toHaveBeenCalled();
		});

		it('falls back to empty saved themes when API fetch fails', async () => {
			appStore.state.savedThemes = [makeTheme('existing-theme', 'Existing')];
			vi.spyOn(console, 'error').mockImplementation(() => {});
			vi.mocked(themesApi.getThemes).mockRejectedValue(new Error('api down'));

			await appStore.loadSavedThemesFromApi();

			expect(appStore.state.savedThemes).toEqual([]);
		});

		it('keeps execution safe when remote theme sync throws', async () => {
			authStoreMock.state.isAuthenticated = true;
			vi.spyOn(console, 'error').mockImplementation(() => {});
			vi.mocked(themesApi.saveTheme).mockRejectedValue(new Error('network down'));

			await expect(appStore.persistThemeChange(makeTheme('theme-crash-1'), 'create')).resolves.toBeUndefined();

			expect(console.error).toHaveBeenCalled();
		});
	});

	describe('downloadImage', () => {
		it('requires image and canvas before downloading', async () => {
			appStore.state.canvas = null;
			appStore.state.imageLoaded = false;
			appStore.state.image = null;

			await appStore.downloadImage();

			expect(toast.error).toHaveBeenCalledWith('Load an image before downloading.');
		});

		it('downloads a clean image blob and revokes object URL', async () => {
			vi.useFakeTimers();
			markImageLoaded();
			const canvas = document.createElement('canvas');
			canvas.width = 640;
			canvas.height = 320;
			appStore.state.canvas = canvas;

			const exportedBlob = new Blob(['img'], { type: 'image/png' });
			vi.spyOn(HTMLCanvasElement.prototype, 'getContext').mockReturnValue({
				drawImage: vi.fn()
			} as never);
			vi.spyOn(HTMLCanvasElement.prototype, 'toBlob').mockImplementation(function (callback: BlobCallback) {
				callback(exportedBlob);
			});
			const createObjectURLSpy = vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob:test-download');
			const revokeObjectURLSpy = vi.spyOn(URL, 'revokeObjectURL').mockImplementation(() => {});
			const anchorClickSpy = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {});

			await appStore.downloadImage();

			expect(createObjectURLSpy).toHaveBeenCalledWith(exportedBlob);
			expect(anchorClickSpy).toHaveBeenCalledTimes(1);
			await vi.advanceTimersByTimeAsync(1);
			expect(revokeObjectURLSpy).toHaveBeenCalledWith('blob:test-download');
			expect(toast.success).toHaveBeenCalledWith('Image downloaded', { id: 'toast-id' });
		});
	});
});
