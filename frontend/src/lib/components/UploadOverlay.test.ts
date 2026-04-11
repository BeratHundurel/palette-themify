import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { appStore } from '$lib/stores/app/store.svelte';

import UploadOverlay from './UploadOverlay.svelte';

async function dispatchCancelableEvent(target: EventTarget, type: string): Promise<Event> {
	const event = new Event(type, { bubbles: true, cancelable: true });
	target.dispatchEvent(event);
	await Promise.resolve();
	return event;
}

describe('UploadOverlay', () => {
	beforeEach(() => {
		appStore.state.imageLoaded = false;
		vi.restoreAllMocks();
	});

	it('shows drag-over styles on dragenter and clears them after matching dragleave events', async () => {
		render(UploadOverlay);
		const dropTarget = screen.getByLabelText('Upload an image or drag and drop it here');

		expect(dropTarget).toHaveClass('border-white/60');
		expect(dropTarget).not.toHaveClass('bg-white/28');

		await fireEvent.dragEnter(dropTarget);
		await fireEvent.dragEnter(dropTarget);
		await waitFor(() => expect(dropTarget).toHaveClass('bg-white/28'));

		await fireEvent.dragLeave(dropTarget);
		expect(dropTarget).toHaveClass('bg-white/28');

		await fireEvent.dragLeave(dropTarget);
		await waitFor(() => {
			expect(dropTarget).toHaveClass('border-white/60');
			expect(dropTarget).not.toHaveClass('bg-white/28');
		});
	});

	it('prevents default browser behavior for drag events', async () => {
		render(UploadOverlay);
		const dropTarget = screen.getByLabelText('Upload an image or drag and drop it here');

		const dragEnterEvent = await dispatchCancelableEvent(dropTarget, 'dragenter');
		const dragOverEvent = await dispatchCancelableEvent(dropTarget, 'dragover');
		const dragLeaveEvent = await dispatchCancelableEvent(dropTarget, 'dragleave');

		expect(dragEnterEvent.defaultPrevented).toBe(true);
		expect(dragOverEvent.defaultPrevented).toBe(true);
		expect(dragLeaveEvent.defaultPrevented).toBe(true);
	});

	it('calls appStore.handleDrop and clears drag-over styles on drop', async () => {
		const handleDropSpy = vi.spyOn(appStore, 'handleDrop').mockResolvedValue(undefined);

		render(UploadOverlay);
		const dropTarget = screen.getByLabelText('Upload an image or drag and drop it here');

		await fireEvent.dragEnter(dropTarget);
		await waitFor(() => expect(dropTarget).toHaveClass('bg-white/28'));

		const dropEvent = await dispatchCancelableEvent(dropTarget, 'drop');

		expect(dropEvent.defaultPrevented).toBe(true);
		expect(handleDropSpy).toHaveBeenCalledTimes(1);
		expect(handleDropSpy).toHaveBeenCalledWith(dropEvent);
		expect(dropTarget).toHaveClass('border-white/60');
		expect(dropTarget).not.toHaveClass('bg-white/28');
	});
});
