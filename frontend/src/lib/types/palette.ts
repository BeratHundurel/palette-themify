import type { Color } from './color';

export type PaletteDTO = {
	id: string;
	name: string;
	palette: Color[];
	createdAt: string;
	isSystem?: boolean;
	isShared?: boolean;
	sharedAt?: string | null;
};

export type GetPalettesResponse = {
	palettes: PaletteDTO[];
};

export type ExtractPaletteResponse = {
	palette: Color[];
};

export type SavePaletteRequest = {
	name: string;
	palette: Color[];
};

export type SavePalettesBatchRequest = {
	palettes: Array<{
		name: string;
		palette: Color[];
	}>;
};
