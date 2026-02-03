import { getAuthHeaders } from './auth';
import type {
	SaveWorkspaceRequest,
	GetWorkspacesResponse,
	SaveWorkspaceResult,
	ShareWorkspaceResult,
	WorkspaceData
} from '$lib/types/workspace';

import { buildURL, ensureOk } from './base';

export async function saveWorkspace(
	name: string,
	imageData: string,
	workspaceState: Omit<SaveWorkspaceRequest, 'name' | 'imageData'>
): Promise<SaveWorkspaceResult> {
	const response = await fetch(buildURL('/workspaces'), {
		method: 'POST',
		headers: getAuthHeaders(),
		body: JSON.stringify({
			name,
			imageData,
			...workspaceState
		})
	});

	await ensureOk(response);
	return response.json();
}

export async function getWorkspaces(): Promise<GetWorkspacesResponse> {
	const response = await fetch(buildURL('/workspaces'), {
		method: 'GET',
		headers: getAuthHeaders()
	});

	await ensureOk(response);
	return response.json();
}

export async function deleteWorkspace(workspaceId: string): Promise<void> {
	const response = await fetch(buildURL(`/workspaces/${workspaceId}`), {
		method: 'DELETE',
		headers: getAuthHeaders()
	});

	await ensureOk(response);
}

export async function shareWorkspace(workspaceId: string): Promise<ShareWorkspaceResult> {
	const response = await fetch(buildURL(`/workspaces/${workspaceId}/share`), {
		method: 'POST',
		headers: getAuthHeaders()
	});

	await ensureOk(response);
	return response.json();
}

export async function removeWorkspaceShare(workspaceId: string): Promise<void> {
	const response = await fetch(buildURL(`/workspaces/${workspaceId}/share`), {
		method: 'DELETE',
		headers: getAuthHeaders()
	});

	await ensureOk(response);
}

export async function getSharedWorkspace(shareToken: string): Promise<WorkspaceData> {
	const response = await fetch(buildURL('/shared', { token: shareToken }), {
		method: 'GET'
	});

	await ensureOk(response);
	return response.json();
}
