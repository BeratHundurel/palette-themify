export type WallhavenResult = {
	id: string;
	thumbs?: { large?: string; original?: string; small?: string };
	path?: string;
	url?: string;
	resolution?: string;
	dimension_x?: number;
	dimension_y?: number;
	ratio?: string;
	file_size?: number;
	file_type?: string;
	created_at?: string;
	colors?: string[];
	purity?: string;
	category?: string;
	views?: number;
	favorites?: number;
};

export type WallhavenSearchResponse = {
	data: WallhavenResult[];
	meta?: {
		current_page?: number;
		last_page?: number;
		per_page?: number;
		total?: number;
		query?: string | { id: number; tag: string };
		seed?: string;
	};
};

export interface WallhavenSettings {
	categories: string;
	purity: string;
	sorting: string;
	order: string;
	topRange: string;
	ratios: string[];
	apikey?: string;
}

export const AVAILABLE_RATIOS = ['16x9', '16x10', '9x16', '1x1', '3x2', '4x3', '5x4', '21x9', '32x9', '48x9', '9x18'];
