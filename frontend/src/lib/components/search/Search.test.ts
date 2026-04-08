import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { appStore } from '$lib/stores/app/store.svelte';
import type { WallhavenSearchResponse } from '$lib/types/wallhaven';

import Search from './Search.svelte';
import { searchWallhaven } from '$lib/api/wallhaven';

vi.mock('$lib/api/wallhaven', () => ({
	searchWallhaven: vi.fn()
}));

const searchWallhavenMock = vi.mocked(searchWallhaven);

function makeResponse(ids: string[]): WallhavenSearchResponse {
	return {
		data: ids.map((id) => ({
			id,
			path: `https://example.com/full/${id}.jpg`,
			thumbs: { original: `https://example.com/thumb/${id}.jpg` }
		})),
		meta: {
			current_page: 1,
			last_page: 1,
			per_page: ids.length,
			total: ids.length
		}
	};
}

describe('Search', () => {
	beforeEach(() => {
		vi.useFakeTimers();
		searchWallhavenMock.mockReset();
		searchWallhavenMock.mockResolvedValue(makeResponse([]));
		appStore.state.searchQuery = '';
	});

	afterEach(() => {
		vi.runOnlyPendingTimers();
		vi.useRealTimers();
	});

	it('opens search modal when form is submitted', async () => {
		render(Search);

		await fireEvent.click(screen.getByRole('button', { name: 'Search' }));

		expect(screen.getByPlaceholderText('Search wallpapers...')).toBeInTheDocument();
		await vi.advanceTimersByTimeAsync(800);
		expect(searchWallhavenMock).not.toHaveBeenCalled();
	});

	it('opens search modal when inline input has meaningful text', async () => {
		render(Search);

		const inlineInput = screen.getByPlaceholderText('Search Wallpapers');
		await fireEvent.input(inlineInput, { target: { value: '   aurora' } });

		const modalInput = screen.getByPlaceholderText('Search wallpapers...') as HTMLInputElement;
		expect(modalInput).toBeInTheDocument();
		expect(modalInput.value).toBe('   aurora');

		await vi.advanceTimersByTimeAsync(760);
		await waitFor(() => {
			expect(searchWallhavenMock).toHaveBeenCalledWith(appStore.state.wallhavenSettings, 'aurora', 1);
		});
	});

	it('does not open search modal for whitespace-only inline input', async () => {
		render(Search);

		const inlineInput = screen.getByPlaceholderText('Search Wallpapers');
		await fireEvent.input(inlineInput, { target: { value: '    ' } });
		await vi.advanceTimersByTimeAsync(800);

		expect(screen.queryByPlaceholderText('Search wallpapers...')).not.toBeInTheDocument();
		expect(searchWallhavenMock).not.toHaveBeenCalled();
	});

	it('closes the modal when escape is pressed', async () => {
		render(Search);

		await fireEvent.click(screen.getByRole('button', { name: 'Search' }));
		expect(screen.getByPlaceholderText('Search wallpapers...')).toBeInTheDocument();

		await fireEvent.keyDown(window, { key: 'Escape' });

		await waitFor(() => {
			expect(screen.queryByPlaceholderText('Search wallpapers...')).not.toBeInTheDocument();
		});
	});
});
