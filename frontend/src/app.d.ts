import type { EditorInstallTarget } from '$lib/types/theme';

// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	interface ImportMetaEnv {
		readonly VITE_API_BASE_URL?: string;
		readonly VITE_ZIG_API_BASE_URL?: string;
		readonly VITE_TELEMETRY_ENABLED?: string;
		readonly VITE_APP_ENV?: string;
		readonly VITE_APP_TARGET?: 'web' | 'desktop';
	}

	interface ImportMeta {
		readonly env: ImportMetaEnv;
	}

	interface Window {
		__THEMESMITH_DESKTOP__?: {
			saveThemeToEditorTarget(target: EditorInstallTarget, themeName: string, themeJSON: string): Promise<string>;
		};
		go?: {
			main?: {
				ThemeExportService?: {
					SaveThemeToEditorTarget(target: EditorInstallTarget, themeName: string, themeJSON: string): Promise<string>;
				};
			};
		};
		_wails?: {
			loadWailsJS?: () => Promise<void>;
		};
	}

	namespace App {
		// interface Error {}
		// interface Locals {}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}
}

export {};
