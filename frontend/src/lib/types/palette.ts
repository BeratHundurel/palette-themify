import type { WallhavenSettings } from '$lib/types/wallhaven';

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
