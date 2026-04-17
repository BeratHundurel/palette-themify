import type { Color } from './color';
import type { EditorThemeType } from '$lib/api/theme';

export type SharedItemKind = 'theme' | 'palette';
export type SharedItemSort = 'newest' | 'oldest' | 'name';

export type SharedItem = {
	id: string;
	kind: SharedItemKind;
	name: string;
	palette: Color[];
	sharedAt: string;
	createdAt: string;
	editorType?: EditorThemeType;
	theme?: unknown;
};

export type SharedItemsResponse = {
	items: SharedItem[];
};
