import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { authStore } from '$lib/stores/auth.svelte';
import { appStore } from '$lib/stores/app/store.svelte';

import UserProfile from './UserProfile.svelte';
import toast from 'svelte-french-toast';

vi.mock('svelte-french-toast', () => ({
	default: {
		success: vi.fn(),
		error: vi.fn()
	}
}));

const demoUser = {
	id: 1,
	name: 'Ada',
	email: 'ada@example.com',
	createdAt: '2026-01-01T00:00:00.000Z',
	updatedAt: '2026-01-01T00:00:00.000Z'
};

describe('UserProfile', () => {
	beforeEach(() => {
		vi.clearAllMocks();

		authStore.state.user = demoUser;
		authStore.state.isAuthenticated = true;
		authStore.state.isLoading = false;

		vi.spyOn(appStore, 'syncPalettesOnAuth').mockResolvedValue(undefined);
		vi.spyOn(appStore, 'syncPreferencesOnAuth').mockResolvedValue(undefined);
		vi.spyOn(appStore, 'syncSavedThemesOnAuth').mockResolvedValue(undefined);
		vi.spyOn(appStore, 'loadSavedPalettes').mockResolvedValue(undefined);
	});

	it('renders nothing when there is no authenticated user', () => {
		authStore.state.user = null;
		render(UserProfile);

		expect(screen.queryByText('Sign Out')).not.toBeInTheDocument();
		expect(screen.queryByText('Ada')).not.toBeInTheDocument();
	});

	it('opens dropdown on profile click and closes on outside click', async () => {
		render(UserProfile);

		await fireEvent.click(screen.getByRole('button', { name: /Ada/ }));
		expect(screen.getByText('ada@example.com')).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Sign Out' })).toBeInTheDocument();

		await fireEvent.click(document.body);

		await waitFor(() => {
			expect(screen.queryByText('ada@example.com')).not.toBeInTheDocument();
		});
	});

	it('logs out successfully and syncs local auth-bound data', async () => {
		const logoutSpy = vi.spyOn(authStore, 'logout').mockResolvedValue(undefined);

		render(UserProfile);

		await fireEvent.click(screen.getByRole('button', { name: /Ada/ }));
		await fireEvent.click(screen.getByRole('button', { name: 'Sign Out' }));

		await waitFor(() => expect(logoutSpy).toHaveBeenCalledTimes(1));
		await waitFor(() => expect(appStore.syncPalettesOnAuth).toHaveBeenCalledTimes(1));
		await waitFor(() => expect(appStore.syncPreferencesOnAuth).toHaveBeenCalledTimes(1));
		await waitFor(() => expect(appStore.syncSavedThemesOnAuth).toHaveBeenCalledTimes(1));
		await waitFor(() => expect(appStore.loadSavedPalettes).toHaveBeenCalledTimes(1));
		expect(toast.success).toHaveBeenCalledWith('Logged out successfully');
	});

	it('shows an error toast when logout fails', async () => {
		vi.spyOn(authStore, 'logout').mockRejectedValue(new Error('network error'));

		render(UserProfile);

		await fireEvent.click(screen.getByRole('button', { name: /Ada/ }));
		await fireEvent.click(screen.getByRole('button', { name: 'Sign Out' }));

		await waitFor(() => expect(toast.error).toHaveBeenCalledWith('Could not sign out. Please try again.'));
		expect(appStore.syncPalettesOnAuth).not.toHaveBeenCalled();
		expect(appStore.syncPreferencesOnAuth).not.toHaveBeenCalled();
		expect(appStore.syncSavedThemesOnAuth).not.toHaveBeenCalled();
		expect(appStore.loadSavedPalettes).not.toHaveBeenCalled();
	});
});
