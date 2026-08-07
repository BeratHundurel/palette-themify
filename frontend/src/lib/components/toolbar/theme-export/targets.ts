import type { EditorInstallTarget, EditorThemeType, VSCodeFamilyTarget } from '$lib/types/theme';

export type VSCodeFamilyTargetOption = {
	value: VSCodeFamilyTarget;
	label: string;
	description: string;
};

export const VSCODE_FAMILY_TARGETS: VSCodeFamilyTargetOption[] = [
	{ value: 'vscode', label: 'VS Code', description: 'Standard VS Code editor' },
	{ value: 'cursor', label: 'Cursor', description: 'Classic Editor Window' },
	{ value: 'antigravity', label: 'Antigravity IDE', description: 'VS Code-based IDE' }
];

export function getInstallTarget(editorType: EditorThemeType, vscodeTarget: VSCodeFamilyTarget): EditorInstallTarget {
	return editorType === 'zed' ? 'zed' : vscodeTarget;
}

export function getInstallTargetLabel(target: EditorInstallTarget): string {
	if (target === 'cursor') return 'Cursor';
	if (target === 'antigravity') return 'Antigravity IDE';
	if (target === 'zed') return 'Zed';
	return 'VS Code';
}
