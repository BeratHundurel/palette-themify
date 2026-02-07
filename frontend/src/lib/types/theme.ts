import type { EditorThemeType } from '$lib/api/theme';

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
};

export type ThemeResponse = {
	themeOverrides: ThemeOverrides;
	theme: Theme;
};

export interface ThemeColorWithUsage {
	baseColor: string;
	label: string;
	variants: Array<{
		color: string;
		usages: string[];
	}>;
	totalUsages: number;
}

export interface ThemeExportState {
	editorType: EditorThemeType;
	themeName: string;
	themeResult: ThemeResponse | null;
	themeColorsWithUsage: ThemeColorWithUsage[];
	lastGeneratedPaletteVersion: number;
	saveOnCopy: boolean;
}

export type SavedThemeItem = {
	id: string;
	name: string;
	editorType: EditorThemeType;
	themeResult: ThemeResponse;
	themeColorsWithUsage: ThemeColorWithUsage[];
	createdAt: string;
};
