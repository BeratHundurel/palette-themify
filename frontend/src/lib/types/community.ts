import type { Color } from './color';
import type { EditorThemeType } from '$lib/types/theme';
import type { ThemeItem } from './theme';

export type CommunityItemKind = 'theme' | 'palette';
export type CommunityItemSort = 'newest' | 'oldest' | 'name';

export type CommunityItem = {
	id: string;
	kind: CommunityItemKind;
	name: string;
	palette: Color[];
	sharedAt: string;
	createdAt: string;
	editorType?: EditorThemeType;
	theme?: ThemeItem;
};

export type CommunityItemsResponse = {
	items: CommunityItem[];
};
