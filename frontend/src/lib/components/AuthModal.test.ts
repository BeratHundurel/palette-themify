import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { authStore } from '$lib/stores/auth.svelte';
import { appStore } from '$lib/stores/app/store.svelte';

import AuthModal from './AuthModal.svelte';
import toast from 'svelte-french-toast';

vi.mock('svelte-french-toast', () => ({
	default: {
		success: vi.fn(),
		error: vi.fn(),
		loading: vi.fn(() => 'toast-id')
	}
}));

const authResponse = {
	token: 'token-1',
	message: 'ok',
	user: {
		id: 1,
		name: 'Ada',
		email: 'ada@example.com',
		createdAt: '2026-01-01T00:00:00.000Z',
		updatedAt: '2026-01-01T00:00:00.000Z'
	}
};

describe('AuthModal', () => {
	beforeEach(() => {
		vi.clearAllMocks();

		authStore.state.user = null;
		authStore.state.isAuthenticated = false;
		authStore.state.isLoading = false;

		vi.spyOn(appStore, 'syncPalettesOnAuth').mockResolvedValue(undefined);
		vi.spyOn(appStore, 'syncPreferencesOnAuth').mockResolvedValue(undefined);
		vi.spyOn(appStore, 'syncSavedThemesOnAuth').mockResolvedValue(undefined);
	});

	it('prevents native form submission and completes login flow', async () => {
		const loginSpy = vi.spyOn(authStore, 'login').mockResolvedValue(authResponse);

		const { container } = render(AuthModal, { isOpen: true });

		await fireEvent.input(screen.getByLabelText('Username'), { target: { value: 'ada' } });
		await fireEvent.input(screen.getByLabelText('Password'), { target: { value: 'password123' } });

		const form = container.querySelector('form') as HTMLFormElement;
		expect(form).toBeTruthy();

		const submitEvent = new Event('submit', { bubbles: true, cancelable: true });
		form.dispatchEvent(submitEvent);
		expect(submitEvent.defaultPrevented).toBe(true);

		await waitFor(() => expect(loginSpy).toHaveBeenCalledWith('ada', 'password123'));
		await waitFor(() => expect(appStore.syncPalettesOnAuth).toHaveBeenCalledTimes(1));
		await waitFor(() => expect(appStore.syncPreferencesOnAuth).toHaveBeenCalledTimes(1));
		await waitFor(() => expect(appStore.syncSavedThemesOnAuth).toHaveBeenCalledTimes(1));
		expect(toast.success).toHaveBeenCalledWith('Login successful!');

		await waitFor(() => {
			expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
		});
	});

	it('shows validation errors for invalid registration input', async () => {
		const registerSpy = vi.spyOn(authStore, 'register').mockResolvedValue(authResponse);

		render(AuthModal, { isOpen: true });

		await fireEvent.click(screen.getByRole('button', { name: 'Sign up' }));
		await fireEvent.input(screen.getByLabelText('Username'), { target: { value: 'ab' } });
		await fireEvent.input(screen.getByLabelText('Password'), { target: { value: 'short' } });
		await fireEvent.input(screen.getByLabelText('Confirm Password'), { target: { value: 'different' } });

		await fireEvent.click(screen.getByRole('button', { name: 'Create Account' }));

		expect(screen.getByText('Username must be at least 3 characters')).toBeInTheDocument();
		expect(screen.getByText('Password must be at least 8 characters')).toBeInTheDocument();
		expect(screen.getByText('Passwords do not match')).toBeInTheDocument();
		expect(registerSpy).not.toHaveBeenCalled();
	});

	it('submits valid registration, syncs local data and closes modal', async () => {
		const registerSpy = vi.spyOn(authStore, 'register').mockResolvedValue(authResponse);

		render(AuthModal, { isOpen: true });

		await fireEvent.click(screen.getByRole('button', { name: 'Sign up' }));
		await fireEvent.input(screen.getByLabelText('Username'), { target: { value: 'new-user' } });
		await fireEvent.input(screen.getByLabelText('Password'), { target: { value: 'password123' } });
		await fireEvent.input(screen.getByLabelText('Confirm Password'), { target: { value: 'password123' } });

		await fireEvent.click(screen.getByRole('button', { name: 'Create Account' }));

		await waitFor(() => expect(registerSpy).toHaveBeenCalledWith('new-user', 'password123'));
		await waitFor(() => expect(appStore.syncPalettesOnAuth).toHaveBeenCalledTimes(1));
		await waitFor(() => expect(appStore.syncPreferencesOnAuth).toHaveBeenCalledTimes(1));
		await waitFor(() => expect(appStore.syncSavedThemesOnAuth).toHaveBeenCalledTimes(1));
		expect(toast.success).toHaveBeenCalledWith('Account created successfully!');

		await waitFor(() => {
			expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
		});
	});

	it('keeps modal open for web Google auth while redirect is expected', async () => {
		const googleLoginSpy = vi.spyOn(authStore, 'googleLogin').mockResolvedValue(undefined as never);

		render(AuthModal, { isOpen: true });

		await fireEvent.click(screen.getByRole('button', { name: 'Continue with Google' }));

		await waitFor(() => expect(googleLoginSpy).toHaveBeenCalledTimes(1));
		expect(appStore.syncPalettesOnAuth).not.toHaveBeenCalled();
		expect(appStore.syncPreferencesOnAuth).not.toHaveBeenCalled();
		expect(appStore.syncSavedThemesOnAuth).not.toHaveBeenCalled();
		expect(screen.getByRole('dialog')).toBeInTheDocument();
		expect(screen.getByText('Processing...')).toBeInTheDocument();
	});

	it('shows error and stays open when login fails', async () => {
		vi.spyOn(authStore, 'login').mockRejectedValue(new Error('invalid credentials'));

		render(AuthModal, { isOpen: true });

		await fireEvent.input(screen.getByLabelText('Username'), { target: { value: 'ada' } });
		await fireEvent.input(screen.getByLabelText('Password'), { target: { value: 'wrongpass123' } });
		await fireEvent.click(screen.getByRole('button', { name: 'Sign In' }));

		await waitFor(() =>
			expect(toast.error).toHaveBeenCalledWith('Could not sign in. Check your details and try again.')
		);
		expect(appStore.syncPalettesOnAuth).not.toHaveBeenCalled();
		expect(appStore.syncPreferencesOnAuth).not.toHaveBeenCalled();
		expect(appStore.syncSavedThemesOnAuth).not.toHaveBeenCalled();
		expect(screen.getByRole('dialog')).toBeInTheDocument();
	});
});
