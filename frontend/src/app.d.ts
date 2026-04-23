import type { EditorThemeType } from '$lib/types/themeApi';

// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	interface Window {
		__THEMESMITH_DESKTOP__?: {
			saveThemeToEditorTarget(editorType: EditorThemeType, themeName: string, themeJSON: string): Promise<string>;
		};
		go?: {
			main?: {
				ThemeExportService?: {
					SaveThemeToEditorTarget(editorType: EditorThemeType, themeName: string, themeJSON: string): Promise<string>;
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
