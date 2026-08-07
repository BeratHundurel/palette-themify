import type { Color } from './color';
import type { VSCodeTheme } from './vscode';
import type { ZedTheme } from './zed';

export type ThemeAppearance = 'dark' | 'light';
export type EditorThemeType = 'vscode' | 'zed';
export type VSCodeFamilyTarget = 'vscode' | 'cursor' | 'antigravity';
export type EditorInstallTarget = VSCodeFamilyTarget | 'zed';
export type AccentBoostCoefficient = number;

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

export type ThemeGenerationResult = {
	theme: Theme;
	themeOverrides: ThemeOverrides;
	rawThemeOverrides: ThemeOverrides;
	colors: Color[];
	boostCoefficient: number;
};

export type ThemeVersionMap = Record<string, ThemeGenerationResult>;

export type ApplyParams = {
	luminosity: number;
	nearest: number;
	power: number;
	maxDistance: number;
};

export type GenerateThemeRequest = {
	colors: Color[];
	type: EditorThemeType;
	name?: string;
	overrides?: ThemeOverrides | null;
	appearance?: ThemeAppearance | null;
	boostCoefficient?: number | null;
};

export type ThemeResponse<TThemeItem> = {
	theme: TThemeItem;
};

export type ThemesResponse<TThemeItem> = {
	themes: TThemeItem[];
};

export interface ThemeExportState {
	themeName: string;
	themeResult: ThemeGenerationResult | null;
	saveOnCopy: boolean;
	editorType: EditorThemeType;
	vscodeTarget: VSCodeFamilyTarget;
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
	vscodeTarget: VSCodeFamilyTarget;
	appearance: ThemeAppearance;
	saveOnCopy: boolean;
	boostCoefficient: number;
};

export type ThemeItem = {
	id: string;
	name: string;
	editorType: EditorThemeType;
	themeResult: ThemeGenerationResult;
	createdAt: string;
	signature?: string;
	isShared?: boolean;
	sharedAt?: string | null;
};

export const DEFAULT_THEME_EXPORT_PREFERENCES: ThemeExportPreferences = {
	editorType: 'vscode',
	vscodeTarget: 'vscode',
	appearance: 'dark',
	saveOnCopy: true,
	boostCoefficient: 1
};
