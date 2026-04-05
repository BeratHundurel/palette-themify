import type { EditorThemeType } from '$lib/api/theme';

export type AppTarget = 'web' | 'desktop';

type ThemeExportService = {
	SaveThemeToEditorTarget(editorType: EditorThemeType, themeName: string, themeJSON: string): Promise<string>;
};

type DesktopSaveFn = (editorType: EditorThemeType, themeName: string, themeJSON: string) => Promise<string>;

function resolveAppTarget(): AppTarget {
	if (import.meta.env.VITE_APP_TARGET === 'desktop') {
		return 'desktop';
	}

	if (typeof window !== 'undefined') {
		if (window.__IMAGE_TO_PALETTE_DESKTOP__) {
			return 'desktop';
		}

		if (window.go?.main?.ThemeExportService) {
			return 'desktop';
		}
	}

	return 'web';
}

function getDesktopBridge(): {
	saveThemeToEditorTarget(editorType: EditorThemeType, themeName: string, themeJSON: string): Promise<string>;
} | null {
	if (typeof window === 'undefined') return null;

	if (window.__IMAGE_TO_PALETTE_DESKTOP__) {
		return window.__IMAGE_TO_PALETTE_DESKTOP__;
	}

	if (window.go?.main?.ThemeExportService) {
		return {
			saveThemeToEditorTarget(editorType: EditorThemeType, themeName: string, themeJSON: string) {
				return window.go!.main!.ThemeExportService!.SaveThemeToEditorTarget(editorType, themeName, themeJSON);
			}
		};
	}

	const saveFn = findSaveThemeMethodInGoBridge();
	if (saveFn) {
		return {
			saveThemeToEditorTarget: saveFn
		};
	}

	return null;
}

function findSaveThemeMethodInGoBridge(): DesktopSaveFn | null {
	if (typeof window === 'undefined' || !window.go) return null;

	const visited = new Set<unknown>();
	const queue: unknown[] = [window.go];

	while (queue.length > 0) {
		const current = queue.shift();
		if (!current || typeof current !== 'object' || visited.has(current)) continue;

		visited.add(current);
		const record = current as Record<string, unknown>;

		const maybeMethod = record.SaveThemeToEditorTarget;
		if (typeof maybeMethod === 'function') {
			return (editorType, themeName, themeJSON) =>
				(maybeMethod as DesktopSaveFn).call(current, editorType, themeName, themeJSON);
		}

		for (const value of Object.values(record)) {
			if (value && typeof value === 'object' && !visited.has(value)) {
				queue.push(value);
			}
		}
	}

	return null;
}

async function ensureDesktopBridgeLoaded() {
	if (typeof window === 'undefined') return;

	if (typeof window._wails?.loadWailsJS === 'function') {
		await window._wails.loadWailsJS();
	}
}

function getThemeExportService(): ThemeExportService | null {
	const bridge = getDesktopBridge();
	if (!bridge) return null;

	return {
		SaveThemeToEditorTarget: bridge.saveThemeToEditorTarget
	};
}

export const appTarget = resolveAppTarget();
export const isDesktopApp = appTarget === 'desktop';

export async function saveThemeToEditorTarget(args: {
	editorType: EditorThemeType;
	themeName: string;
	themeJSON: string;
}): Promise<string> {
	if (!isDesktopApp) {
		throw new Error('Desktop save is only available in desktop builds.');
	}

	const service = getThemeExportService();
	if (service) {
		return service.SaveThemeToEditorTarget(args.editorType, args.themeName, args.themeJSON);
	}

	try {
		await ensureDesktopBridgeLoaded();
	} catch {
		throw new Error('Desktop runtime bridge failed to load.');
	}

	const loadedService = getThemeExportService();
	if (!loadedService) {
		throw new Error('Desktop save bridge is not available.');
	}

	return loadedService.SaveThemeToEditorTarget(args.editorType, args.themeName, args.themeJSON);
}
