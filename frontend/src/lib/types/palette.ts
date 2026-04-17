import type { Color } from './color';

export type PaletteData = {
	id: string;
	name: string;
	palette: Color[];
	createdAt: string;
	isSystem?: boolean;
	isShared?: boolean;
	sharedAt?: string | null;
};

export type GetPalettesResponse = {
	palettes: PaletteData[];
};

export type ExtractPaletteResponse = {
	palette: Color[];
};

export type SavePaletteRequest = {
	name: string;
	palette: Color[];
};

export type SavePaletteResult = {
	message: string;
	name: string;
};
