import type { EditorThemeType, ThemeOverrides } from '$lib/api/palette';
import type { ThemeResponse } from '$lib/types/palette';

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
	generatedTheme: ThemeResponse | null;
	themeOverrides: ThemeOverrides;
	themeColorsWithUsage: ThemeColorWithUsage[];
	lastGeneratedPaletteVersion: number;
}
