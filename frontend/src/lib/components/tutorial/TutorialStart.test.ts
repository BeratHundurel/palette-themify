import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { tutorialStore } from '$lib/stores/tutorial.svelte';

import TutorialStart from './TutorialStart.svelte';

describe('TutorialStart', () => {
	beforeEach(() => {
		vi.restoreAllMocks();
		vi.useRealTimers();
		if (!Element.prototype.animate) {
			Element.prototype.animate = () => {
				return {
					cancel: () => {},
					finish: () => {},
					play: () => {},
					pause: () => {},
					reverse: () => {},
					updatePlaybackRate: () => {},
					addEventListener: () => {},
					removeEventListener: () => {},
					onfinish: null,
					oncancel: null,
					currentTime: null,
					playbackRate: 1,
					playState: 'finished',
					startTime: null,
					effect: null,
					timeline: null,
					finished: Promise.resolve(),
					ready: Promise.resolve(),
					commitStyles: () => {},
					persist: () => {},
					replaceState: 'active'
				} as unknown as Animation;
			};
		}
		tutorialStore.reset();
		localStorage.removeItem('tutorial-completed');
		localStorage.removeItem('tutorial-skipped');
	});

	it('renders the prompt for first-time users', async () => {
		render(TutorialStart);

		expect(await screen.findByText('Welcome to ThemeSmith!')).toBeInTheDocument();
		expect(screen.getByText('Take the Tour')).toBeInTheDocument();
	});

	it('starts tutorial after the delayed close animation window', async () => {
		vi.useFakeTimers();
		const startSpy = vi.spyOn(tutorialStore, 'start');

		render(TutorialStart);

		await fireEvent.click(await screen.findByRole('button', { name: /take the tour/i }));

		expect(startSpy).not.toHaveBeenCalled();
		vi.advanceTimersByTime(179);
		expect(startSpy).not.toHaveBeenCalled();
		vi.advanceTimersByTime(1);
		expect(startSpy).toHaveBeenCalledTimes(1);
	});

	it('dismisses prompt and stores tutorial-skipped flag', async () => {
		render(TutorialStart);

		await fireEvent.click(await screen.findByRole('button', { name: /skip for now/i }));

		await waitFor(() => {
			expect(tutorialStore.shouldShowTutorial()).toBe(false);
		});
		expect(localStorage.getItem('tutorial-skipped')).toBe('true');
	});
});
