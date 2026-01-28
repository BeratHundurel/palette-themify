import type { WallhavenSettings } from '$lib/types/wallhaven';
import type { VSCodeTheme } from './vscode';
import type { ZedTheme } from './zed';

export type Color = {
	hex: string;
};

export type PaletteData = {
	id: string;
	name: string;
	palette: Color[];
	createdAt: string;
	isSystem?: boolean;
};

export type SavePaletteRequest = {
	name: string;
	palette: Color[];
};

export type GetPalettesResponse = {
	palettes: PaletteData[];
};

export type SavePaletteResult = {
	message: string;
	name: string;
};

export type NamedColor = {
	name: string;
	hex: string;
};

export type Selector = {
	id: string;
	color: string;
	selected: boolean;
	selection?: { x: number; y: number; w: number; h: number };
};

export type WorkspaceData = {
	id: string;
	name: string;
	imageData: string;
	colors?: Color[];
	selectors?: Selector[];
	activeSelectorId?: string;
	luminosity?: number;
	nearest?: number;
	power?: number;
	maxDistance?: number;
	wallhavenSettings?: WallhavenSettings;
	shareToken?: string | null;
	createdAt: string;
};

export type SaveWorkspaceRequest = {
	name: string;
	imageData: string;
	colors: Color[];
	selectors: Selector[];
	activeSelectorId: string;
	luminosity: number;
	nearest: number;
	power: number;
	maxDistance: number;
	wallhavenSettings?: WallhavenSettings;
};

export type GetWorkspacesResponse = {
	workspaces: WorkspaceData[];
};

export type SaveWorkspaceResult = {
	message: string;
	name: string;
};

export type ShareWorkspaceResult = {
	shareToken: string;
	shareUrl: string;
};

export type Theme = VSCodeTheme | ZedTheme;

export type ThemeOverrides = {
	background?: string | null;
	foreground?: string | null;
	c1?: string | null;
	c2?: string | null;
	c3?: string | null;
	c4?: string | null;
	c5?: string | null;
	c6?: string | null;
	c7?: string | null;
	c8?: string | null;
};

export type ThemeResponse = {
	baseOverrides: ThemeOverrides;
	theme: Theme;
};

export type ExtractPaletteResponse = {
	palette: Color[];
};
