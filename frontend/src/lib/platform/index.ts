import type { EditorThemeType } from '$lib/api/theme';

export type AppTarget = 'web' | 'desktop';
export type DesktopOS = 'darwin' | 'windows' | 'linux' | 'unknown';
type SaveBridgeFn = (editorType: EditorThemeType, themeName: string, themeJSON: string) => Promise<string>;

let cachedWailsBridge: SaveBridgeFn | null = null;
let cachedWailsRuntime: Promise<typeof import('@wailsio/runtime')> | null = null;

declare global {
	interface Window {
		['image-to-palette']?: {
			desktop?: {
				ThemeExportService?: {
					SaveThemeToEditorTarget?: SaveBridgeFn;
				};
			};
		};
	}
}

function resolveAppTarget(): AppTarget {
	if (import.meta.env.VITE_APP_TARGET === 'desktop') {
		return 'desktop';
	}

	if (import.meta.env.VITE_APP_TARGET === 'web') {
		return 'web';
	}

	if (typeof window !== 'undefined') {
		if (window.__IMAGE_TO_PALETTE_DESKTOP__) {
			return 'desktop';
		}

		if (window._wails?.loadWailsJS) {
			return 'desktop';
		}

		if (resolveWindowWailsBridge()) {
			return 'desktop';
		}
	}

	return 'web';
}

function resolveDesktopOS(): DesktopOS {
	if (typeof window === 'undefined') {
		return 'unknown';
	}

	const runtimeOS = window._wails?.environment?.OS;
	if (runtimeOS === 'darwin' || runtimeOS === 'windows' || runtimeOS === 'linux') {
		return runtimeOS;
	}

	return 'unknown';
}

async function callWailsFunction(editorType: EditorThemeType, themeName: string, themeJSON: string): Promise<string> {
	if (typeof window === 'undefined') {
		throw new Error('Window is not available');
	}

	if (cachedWailsBridge) {
		return await cachedWailsBridge(editorType, themeName, themeJSON);
	}

	const immediateBridge = resolveWindowWailsBridge();
	if (immediateBridge) {
		cachedWailsBridge = immediateBridge;
		return await immediateBridge(editorType, themeName, themeJSON);
	}

	const runtimeBridge = await resolveRuntimeWailsBridge();
	if (runtimeBridge) {
		cachedWailsBridge = runtimeBridge;
		return await runtimeBridge(editorType, themeName, themeJSON);
	}

	const timeoutMs = 400;
	const pollMs = 40;
	const startedAt = Date.now();

	while (Date.now() - startedAt < timeoutMs) {
		await new Promise((resolve) => setTimeout(resolve, pollMs));
		const deferredBridge = resolveWindowWailsBridge();
		if (deferredBridge) {
			cachedWailsBridge = deferredBridge;
			return await deferredBridge(editorType, themeName, themeJSON);
		}
	}

	const runtimeBridgeAfterWait = await resolveRuntimeWailsBridge();
	if (runtimeBridgeAfterWait) {
		cachedWailsBridge = runtimeBridgeAfterWait;
		return await runtimeBridgeAfterWait(editorType, themeName, themeJSON);
	}

	throw new Error('Desktop save bridge is not available. Make sure the Wails runtime is loaded.');
}

function resolveWindowWailsBridge(): SaveBridgeFn | null {
	const wailsBindings = window['image-to-palette'];
	if (wailsBindings?.desktop?.ThemeExportService?.SaveThemeToEditorTarget) {
		return wailsBindings.desktop.ThemeExportService.SaveThemeToEditorTarget;
	}

	const directBinding = window.go?.main?.ThemeExportService?.SaveThemeToEditorTarget;
	if (directBinding && typeof directBinding === 'function') {
		return directBinding;
	}

	const goObject = window.go as unknown;
	if (isRecord(goObject)) {
		const fn = findFunctionInObject(goObject, 'SaveThemeToEditorTarget');
		if (fn && typeof fn === 'function') {
			return fn as SaveBridgeFn;
		}
	}

	return null;
}

async function resolveRuntimeWailsBridge(): Promise<SaveBridgeFn | null> {
	try {
		if (!cachedWailsRuntime) {
			cachedWailsRuntime = import('@wailsio/runtime');
		}

		const runtime = await cachedWailsRuntime;
		const byName = runtime?.Call?.ByName;
		if (typeof byName !== 'function') {
			return null;
		}

		return async (targetEditorType, targetThemeName, targetThemeJSON) => {
			try {
				return await byName(
					'main.ThemeExportService.SaveThemeToEditorTarget',
					targetEditorType,
					targetThemeName,
					targetThemeJSON
				);
			} catch {
				return await byName(
					'ThemeExportService.SaveThemeToEditorTarget',
					targetEditorType,
					targetThemeName,
					targetThemeJSON
				);
			}
		};
	} catch {
		return null;
	}
}

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === 'object' && value !== null;
}

function findFunctionInObject(obj: unknown, fnName: string): SaveBridgeFn | null {
	if (!isRecord(obj)) return null;

	const visited = new Set<object>();
	const queue: Record<string, unknown>[] = [obj];

	while (queue.length > 0) {
		const current = queue.shift();
		if (!current || visited.has(current)) continue;

		visited.add(current);

		const candidate = current[fnName];
		if (typeof candidate === 'function') {
			return candidate as SaveBridgeFn;
		}

		for (const value of Object.values(current)) {
			if (isRecord(value) && !visited.has(value)) {
				queue.push(value);
			}
		}
	}

	return null;
}

export const appTarget = resolveAppTarget();
export const isDesktopApp = appTarget === 'desktop';
export const desktopOS = isDesktopApp ? resolveDesktopOS() : 'unknown';
export const isMacDesktop = desktopOS === 'darwin';

export function getDesktopSaveErrorMessage(error: unknown): string {
	const message = error instanceof Error ? error.message : '';

	if (message.includes('Desktop save bridge is not available')) {
		return 'Could not connect to the desktop bridge. Please restart the app and try again.';
	}

	if (message.includes('Desktop save is only available in desktop builds')) {
		return 'Saving to your editor is only available in the desktop app.';
	}

	if (message.includes('Window is not available')) {
		return 'This action is only available from the desktop app window.';
	}

	return 'Could not save the theme to your editor. Please try again.';
}

export async function saveThemeToEditorTarget(args: {
	editorType: EditorThemeType;
	themeName: string;
	themeJSON: string;
}): Promise<string> {
	if (!isDesktopApp) {
		throw new Error('Desktop save is only available in desktop builds.');
	}

	if (typeof window !== 'undefined' && window.__IMAGE_TO_PALETTE_DESKTOP__?.saveThemeToEditorTarget) {
		return window.__IMAGE_TO_PALETTE_DESKTOP__.saveThemeToEditorTarget(args.editorType, args.themeName, args.themeJSON);
	}

	return await callWailsFunction(args.editorType, args.themeName, args.themeJSON);
}
