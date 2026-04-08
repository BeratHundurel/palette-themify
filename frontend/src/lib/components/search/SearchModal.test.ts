import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { appStore } from '$lib/stores/app/store.svelte';
import type { WallhavenSearchResponse } from '$lib/types/wallhaven';

import SearchModal from './SearchModal.svelte';
import { searchWallhaven } from '$lib/api/wallhaven';

vi.mock('$lib/api/wallhaven', () => ({
	searchWallhaven: vi.fn()
}));

vi.mock('svelte-french-toast', () => ({
	default: {
		loading: vi.fn(() => 'toast-id')
	}
}));

const searchWallhavenMock = vi.mocked(searchWallhaven);

function deferred<T>() {
	let resolve!: (value: T) => void;
	const promise = new Promise<T>((res) => {
		resolve = res;
	});
	return { promise, resolve };
}

function makeResponse(ids: string[], page: number, lastPage: number): WallhavenSearchResponse {
	return {
		data: ids.map((id) => ({
			id,
			path: `https://example.com/full/${id}.jpg`,
			thumbs: { original: `https://example.com/thumb/${id}.jpg` }
		})),
		meta: {
			current_page: page,
			last_page: lastPage,
			per_page: ids.length,
			total: ids.length * lastPage
		}
	};
}

describe('SearchModal', () => {
	beforeEach(() => {
		vi.useFakeTimers();
		searchWallhavenMock.mockReset();
		appStore.state.searchQuery = '';
	});

	afterEach(() => {
		vi.runOnlyPendingTimers();
		vi.useRealTimers();
	});

	it('debounces search calls and only sends the latest query', async () => {
		searchWallhavenMock.mockResolvedValue(makeResponse(['a-1', 'a-2'], 1, 1));

		render(SearchModal, { isOpen: true });
		const input = screen.getByPlaceholderText('Search wallpapers...');

		await fireEvent.input(input, { target: { value: 'n' } });
		await fireEvent.input(input, { target: { value: 'na' } });
		await fireEvent.input(input, { target: { value: 'nature' } });

		await vi.advanceTimersByTimeAsync(740);
		expect(searchWallhavenMock).not.toHaveBeenCalled();

		await vi.advanceTimersByTimeAsync(15);
		await waitFor(() => expect(searchWallhavenMock).toHaveBeenCalledTimes(1));
		expect(searchWallhavenMock).toHaveBeenCalledWith(appStore.state.wallhavenSettings, 'nature', 1);
	});

	it('loads next page on scroll and deduplicates repeated ids', async () => {
		searchWallhavenMock.mockImplementation(async (_settings, _query, page) => {
			if (page === 1) {
				return makeResponse(['id-1', 'id-2'], 1, 2);
			}

			return makeResponse(['id-2', 'id-3'], 2, 2);
		});

		const { container } = render(SearchModal, { isOpen: true });
		const input = screen.getByPlaceholderText('Search wallpapers...');

		await fireEvent.input(input, { target: { value: 'forest' } });
		await vi.advanceTimersByTimeAsync(760);

		await screen.findByText('2 results for "forest"');

		const scroller = container.querySelector('.custom-scrollbar') as HTMLDivElement;
		expect(scroller).toBeTruthy();

		Object.defineProperty(scroller, 'scrollHeight', { value: 1000, configurable: true });
		Object.defineProperty(scroller, 'clientHeight', { value: 600, configurable: true });
		scroller.scrollTop = 450;

		await fireEvent.scroll(scroller);
		await waitFor(() => expect(searchWallhavenMock).toHaveBeenCalledTimes(2));
		await screen.findByText('3 results for "forest"');
	});

	it('does not load more while a new query debounce is pending', async () => {
		searchWallhavenMock.mockImplementation(async (_settings, query, page) => {
			const currentPage = page ?? 1;

			if (query === 'forest' && currentPage === 1) {
				return makeResponse(['f-1', 'f-2'], 1, 3);
			}

			if (query === 'desert' && currentPage === 1) {
				return makeResponse(['d-1', 'd-2'], 1, 3);
			}

			if (query === 'desert' && currentPage === 2) {
				return makeResponse(['d-3'], 2, 3);
			}

			return makeResponse([], currentPage, currentPage);
		});

		const { container } = render(SearchModal, { isOpen: true });
		const input = screen.getByPlaceholderText('Search wallpapers...');

		await fireEvent.input(input, { target: { value: 'forest' } });
		await vi.advanceTimersByTimeAsync(760);
		await screen.findByText('2 results for "forest"');

		const scroller = container.querySelector('.custom-scrollbar') as HTMLDivElement;
		expect(scroller).toBeTruthy();

		Object.defineProperty(scroller, 'scrollHeight', { value: 1000, configurable: true });
		Object.defineProperty(scroller, 'clientHeight', { value: 600, configurable: true });

		await fireEvent.input(input, { target: { value: 'desert' } });
		scroller.scrollTop = 450;
		await fireEvent.scroll(scroller);

		expect(searchWallhavenMock).toHaveBeenCalledTimes(1);

		await vi.advanceTimersByTimeAsync(760);
		await waitFor(() => expect(searchWallhavenMock).toHaveBeenCalledTimes(2));
		expect(searchWallhavenMock).toHaveBeenNthCalledWith(2, appStore.state.wallhavenSettings, 'desert', 1);
	});

	it('cancels pending debounce when closed before timer fires', async () => {
		searchWallhavenMock.mockResolvedValue(makeResponse(['x-1'], 1, 1));

		render(SearchModal, { isOpen: true });
		const input = screen.getByPlaceholderText('Search wallpapers...');

		await fireEvent.input(input, { target: { value: 'cancel-me' } });
		await fireEvent.click(screen.getByLabelText('Close search'));

		await vi.advanceTimersByTimeAsync(800);
		expect(searchWallhavenMock).not.toHaveBeenCalled();
	});

	it('cancels pending debounce when query is cleared', async () => {
		searchWallhavenMock.mockResolvedValue(makeResponse(['clear-1'], 1, 1));

		render(SearchModal, { isOpen: true });
		const input = screen.getByPlaceholderText('Search wallpapers...');

		await fireEvent.input(input, { target: { value: 'to-clear' } });
		await fireEvent.input(input, { target: { value: '' } });

		await vi.advanceTimersByTimeAsync(800);
		expect(searchWallhavenMock).not.toHaveBeenCalled();
	});

	it('ignores stale responses from older requests', async () => {
		const older = deferred<WallhavenSearchResponse>();
		searchWallhavenMock
			.mockImplementationOnce(() => older.promise)
			.mockResolvedValueOnce(makeResponse(['new-1', 'new-2'], 1, 1));

		render(SearchModal, { isOpen: true });
		const input = screen.getByPlaceholderText('Search wallpapers...');

		await fireEvent.input(input, { target: { value: 'old' } });
		await vi.advanceTimersByTimeAsync(760);
		await waitFor(() => expect(searchWallhavenMock).toHaveBeenCalledTimes(1));

		await fireEvent.input(input, { target: { value: 'new' } });
		await vi.advanceTimersByTimeAsync(760);
		await waitFor(() => expect(searchWallhavenMock).toHaveBeenCalledTimes(2));
		await screen.findByText('2 results for "new"');

		older.resolve(makeResponse(['old-1'], 1, 1));
		await Promise.resolve();
		await Promise.resolve();

		expect(screen.getByText('2 results for "new"')).toBeInTheDocument();
		expect(screen.queryByText('1 result for "new"')).not.toBeInTheDocument();
	});

	it('loads wallpaper when a result is clicked and closes modal', async () => {
		searchWallhavenMock.mockResolvedValue(makeResponse(['pick-1'], 1, 1));
		const loadWallhavenImageSpy = vi.spyOn(appStore, 'loadWallhavenImage').mockResolvedValue();

		render(SearchModal, { isOpen: true });
		const input = screen.getByPlaceholderText('Search wallpapers...');

		await fireEvent.input(input, { target: { value: 'pick' } });
		await vi.advanceTimersByTimeAsync(760);

		const resultThumb = await screen.findByAltText('wallpaper thumb');
		await fireEvent.click(resultThumb.closest('button') as HTMLButtonElement);

		expect(loadWallhavenImageSpy).toHaveBeenCalledWith('https://example.com/full/pick-1.jpg', 'toast-id');
		await waitFor(() => {
			expect(screen.queryByPlaceholderText('Search wallpapers...')).not.toBeInTheDocument();
		});

		loadWallhavenImageSpy.mockRestore();
	});

	it('shows no-results state after successful empty response', async () => {
		searchWallhavenMock.mockResolvedValue(makeResponse([], 1, 1));

		render(SearchModal, { isOpen: true });
		const input = screen.getByPlaceholderText('Search wallpapers...');

		await fireEvent.input(input, { target: { value: 'impossible-query' } });
		await vi.advanceTimersByTimeAsync(760);

		await screen.findByText('No results found');
		expect(screen.getByText('Try adjusting your search terms')).toBeInTheDocument();
		expect(screen.queryByText('Searching wallpapers...')).not.toBeInTheDocument();
	});

	it('shows no-results state when search request fails', async () => {
		searchWallhavenMock.mockRejectedValue(new Error('wallhaven unavailable'));

		render(SearchModal, { isOpen: true });
		const input = screen.getByPlaceholderText('Search wallpapers...');

		await fireEvent.input(input, { target: { value: 'failure-case' } });
		await vi.advanceTimersByTimeAsync(760);

		await screen.findByText('No results found');
		expect(screen.getByText('Try adjusting your search terms')).toBeInTheDocument();
		expect(screen.queryByText('Searching wallpapers...')).not.toBeInTheDocument();
	});

	it('stops pagination once api reports no more pages', async () => {
		searchWallhavenMock.mockResolvedValue(makeResponse(['last-1', 'last-2'], 1, 1));

		const { container } = render(SearchModal, { isOpen: true });
		const input = screen.getByPlaceholderText('Search wallpapers...');

		await fireEvent.input(input, { target: { value: 'final-page' } });
		await vi.advanceTimersByTimeAsync(760);
		await screen.findByText('2 results for "final-page"');

		const scroller = container.querySelector('.custom-scrollbar') as HTMLDivElement;
		expect(scroller).toBeTruthy();

		Object.defineProperty(scroller, 'scrollHeight', { value: 1000, configurable: true });
		Object.defineProperty(scroller, 'clientHeight', { value: 600, configurable: true });
		scroller.scrollTop = 450;

		await fireEvent.scroll(scroller);
		await Promise.resolve();

		expect(searchWallhavenMock).toHaveBeenCalledTimes(1);
		expect(screen.getByText("You've reached the end of the results")).toBeInTheDocument();
	});
});
