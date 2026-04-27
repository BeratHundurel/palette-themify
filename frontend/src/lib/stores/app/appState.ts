import type { SortMethod } from '$lib/colorUtils';
import type { ApplyPaletteSettings } from '$lib/types/applyPaletteSettings';
import type { Color } from '$lib/types/color';
import type { PaletteDTO } from '$lib/types/palette';
import type { Selector } from '$lib/types/selector';
import type { ThemeItem, ThemeExportState } from '$lib/types/theme';
import type { WallhavenResult, WallhavenSettings } from '$lib/types/wallhaven';

export interface AppState {
	fileInput: HTMLInputElement | null;
	canvas: HTMLCanvasElement | null;
	canvasContext: CanvasRenderingContext2D | null;
	image: HTMLImageElement | null;
	originalImageWidth: number;
	originalImageHeight: number;
	canvasScaleX: number;
	canvasScaleY: number;

	searchQuery: string;

	imageLoaded: boolean;
	isDragging: boolean;
	isExtracting: boolean;
	startX: number;
	startY: number;
	dragRect: DOMRect | null;
	dragScaleX: number;
	dragScaleY: number;

	colors: Color[];
	wallhavenResults: WallhavenResult[];
	wallhavenSettings: WallhavenSettings;
	selectors: Selector[];
	activeSelectorId: string;
	newFilterColor: string;
	savedPalettes: PaletteDTO[];
	sortMethod: SortMethod;
	applyPaletteSettings: ApplyPaletteSettings;
	themeExport: ThemeExportState;
	savedThemes: ThemeItem[];
	paletteVersion: number;
}
