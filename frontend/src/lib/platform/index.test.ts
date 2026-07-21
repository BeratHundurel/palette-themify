import { describe, expect, it } from 'vitest';

import { getDesktopSaveErrorMessage } from './index';

describe('getDesktopSaveErrorMessage', () => {
	it.each([
		new Error('permission denied: cannot write theme'),
		'Access is denied.',
		{ message: 'operation not permitted' },
		new Error('read-only file system')
	])('explains filesystem permission failures', (error) => {
		expect(getDesktopSaveErrorMessage(error)).toContain('does not have permission');
	});

	it('explains a full destination drive', () => {
		expect(getDesktopSaveErrorMessage(new Error('no space left on device'))).toContain('drive is full');
	});

	it('explains an invalid existing VS Code manifest', () => {
		expect(getDesktopSaveErrorMessage(new Error('failed to parse existing VS Code package.json'))).toContain(
			'themesmith-local'
		);
	});
});
