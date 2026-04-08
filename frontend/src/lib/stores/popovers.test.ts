import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { popoverStore } from '$lib/stores/popovers.svelte';

function createAnchor(right: number): HTMLButtonElement {
	const button = document.createElement('button');
	button.className = 'toolbar-button-base';
	button.getBoundingClientRect = () =>
		({
			top: 0,
			left: Math.max(0, right - 100),
			right,
			bottom: 40,
			width: 100,
			height: 40,
			x: Math.max(0, right - 100),
			y: 0,
			toJSON: () => ({})
		}) as DOMRect;
	document.body.append(button);
	return button;
}

function togglePopover(anchor: HTMLElement) {
	popoverStore.toggle('saved', {
		stopPropagation: () => {},
		currentTarget: anchor
	} as unknown as MouseEvent);
	vi.advanceTimersByTime(15);
}

describe('popoverStore', () => {
	beforeEach(() => {
		document.body.innerHTML = '';
		popoverStore.close();
		vi.useFakeTimers();
	});

	afterEach(() => {
		popoverStore.close();
		document.body.innerHTML = '';
		vi.runOnlyPendingTimers();
		vi.useRealTimers();
	});

	it('opens popover and resolves right direction when there is enough space', () => {
		const anchor = createAnchor(250);
		togglePopover(anchor);

		expect(popoverStore.state.current).toBe('saved');
		expect(popoverStore.state.direction).toBe('right');
	});

	it('resolves left direction near viewport edge', () => {
		Object.defineProperty(window, 'innerWidth', { value: 1024, configurable: true, writable: true });
		const anchor = createAnchor(980);
		togglePopover(anchor);

		expect(popoverStore.state.current).toBe('saved');
		expect(popoverStore.state.direction).toBe('left');
	});

	it('closes when clicking outside popover and toolbar button', () => {
		const anchor = createAnchor(250);
		togglePopover(anchor);

		const outside = document.createElement('div');
		document.body.append(outside);
		outside.dispatchEvent(new MouseEvent('click', { bubbles: true }));

		expect(popoverStore.state.current).toBeNull();
	});

	it('stays open when clicking inside popover', () => {
		const anchor = createAnchor(250);
		togglePopover(anchor);

		const popover = document.createElement('div');
		popover.className = 'palette-dropdown-base';
		const child = document.createElement('button');
		popover.append(child);
		document.body.append(popover);

		child.dispatchEvent(new MouseEvent('click', { bubbles: true }));

		expect(popoverStore.state.current).toBe('saved');
	});

	it('stays open when clicking another toolbar button', () => {
		const anchor = createAnchor(250);
		togglePopover(anchor);

		const otherButton = document.createElement('button');
		otherButton.className = 'toolbar-button-base';
		document.body.append(otherButton);

		otherButton.dispatchEvent(new MouseEvent('click', { bubbles: true }));

		expect(popoverStore.state.current).toBe('saved');
	});

	it('toggles off when opening the same popover twice', () => {
		const anchor = createAnchor(250);
		togglePopover(anchor);
		expect(popoverStore.state.current).toBe('saved');

		popoverStore.toggle('saved', {
			stopPropagation: () => {},
			currentTarget: anchor
		} as unknown as MouseEvent);

		expect(popoverStore.state.current).toBeNull();
	});
});
