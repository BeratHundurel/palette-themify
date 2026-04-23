import type { EditorThemeType, ThemeAppearance } from '$lib/types/themeApi';
import type { Color } from './color';
import type { VSCodeTheme } from './vscode';
import type { ZedTheme } from './zed';

export type Theme = VSCodeTheme | ZedTheme;

export type ThemeOverrides = {
	background?: string;
	foreground?: string;
	c1?: string;
	c2?: string;
	c3?: string;
	c4?: string;
	c5?: string;
	c6?: string;
	c7?: string;
	c8?: string;
	c9?: string;
	constants?: string;
};

export type ThemeGenerationResponse = {
	theme: Theme;
	themeOverrides: ThemeOverrides;
	rawThemeOverrides: ThemeOverrides;
	colors: Color[];
	boostCoefficient: number;
};

export type ThemeVersionMap = Record<string, ThemeGenerationResponse>;

export interface ThemeExportState {
	themeName: string;
	themeResult: ThemeGenerationResponse | null;
	saveOnCopy: boolean;
	editorType: EditorThemeType;
	appearance: ThemeAppearance;
	boostCoefficient: number;
	lastGeneratedPaletteVersion: number;
	themeVersions: ThemeVersionMap;
	rawThemeOverrides: ThemeOverrides;
	hasManualBackgroundOverride: boolean;
	hasManualForegroundOverride: boolean;
	loadedThemeOverridesReference: ThemeOverrides | null;
	backupColors: Color[] | null;
}

export type ThemeExportPreferences = {
	editorType: EditorThemeType;
	appearance: ThemeAppearance;
	saveOnCopy: boolean;
	boostCoefficient: number;
};

export type SavedThemeItem = {
	id: string;
	name: string;
	editorType: EditorThemeType;
	themeResult: ThemeGenerationResponse;
	createdAt: string;
	signature?: string;
	isShared?: boolean;
	sharedAt?: string | null;
};

export const DEFAULT_THEME_EXPORT_PREFERENCES: ThemeExportPreferences = {
	editorType: 'vscode',
	appearance: 'dark',
	saveOnCopy: true,
	boostCoefficient: 1
};
