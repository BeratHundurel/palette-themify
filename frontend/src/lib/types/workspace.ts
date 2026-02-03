import type { WallhavenSettings } from '$lib/types/wallhaven';

import type { Selector } from './selector';
import type { Color } from './color';

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
