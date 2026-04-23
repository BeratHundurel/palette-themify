import type { Color } from './color';
import type { ThemeOverrides } from './theme';

export type ThemeAppearance = 'dark' | 'light';
export type EditorThemeType = 'vscode' | 'zed';
export type AccentBoostCoefficient = number;

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

export type ThemesResponse<TThemeItem> = {
	themes: TThemeItem[];
};

export type SavedThemeResponse<TThemeItem> = {
	message: string;
	theme: TThemeItem;
};

export type SavedThemesBatchResponse<TThemeItem> = {
	message: string;
	themes: TThemeItem[];
};
