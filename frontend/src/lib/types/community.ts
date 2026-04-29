import type { Color } from './color';
import { COLOR_REGEX } from './color';
import type { ThemeGenerationResult } from './theme';

export type CommunityItemKind = 'theme' | 'palette';
export type CommunityItemSort = 'newest' | 'oldest' | 'name';

export type CommunityItem = {
	itemId: number;
	kind: CommunityItemKind;
	name: string;
	jsonData: unknown;
	sharedAt: string;
	createdAt: string;
	editorType?: string;
	signature?: string;
};

export type CommunityItemsResponse = {
	items: CommunityItem[];
};

const isRecord = (value: unknown): value is Record<string, unknown> => typeof value === 'object' && value !== null;

const isColor = (value: unknown): value is Color =>
	isRecord(value) && typeof value.hex === 'string' && COLOR_REGEX.test(value.hex);

const isThemeGenerationResult = (value: unknown): value is ThemeGenerationResult => {
	if (!isRecord(value)) return false;
	if (!isRecord(value.theme)) return false;
	if (!Array.isArray(value.colors) || !value.colors.every(isColor)) return false;
	if (typeof value.boostCoefficient !== 'number') return false;
	return true;
};

export const parseCommunityPalette = (value: unknown): Color[] | null => {
	if (!Array.isArray(value)) return null;
	const colors = value.filter(isColor);
	return colors.length > 0 ? colors : null;
};

export const parseCommunityTheme = (value: unknown): ThemeGenerationResult | null =>
	isThemeGenerationResult(value) ? value : null;
