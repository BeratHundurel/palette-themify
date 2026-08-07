import { describe, expect, it } from 'vitest';

import { getInstallTarget, getInstallTargetLabel } from './targets';

describe('theme install targets', () => {
	it('uses the selected VS Code family destination', () => {
		expect(getInstallTarget('vscode', 'vscode')).toBe('vscode');
		expect(getInstallTarget('vscode', 'cursor')).toBe('cursor');
		expect(getInstallTarget('vscode', 'antigravity')).toBe('antigravity');
	});

	it('always installs Zed themes in Zed', () => {
		expect(getInstallTarget('zed', 'cursor')).toBe('zed');
	});

	it('provides user-facing editor labels', () => {
		expect(getInstallTargetLabel('antigravity')).toBe('Antigravity IDE');
		expect(getInstallTargetLabel('cursor')).toBe('Cursor');
	});
});
