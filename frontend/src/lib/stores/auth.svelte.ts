import { browser } from '$app/environment';
import type { User } from '$lib/api/auth';
import { isDesktopApp } from '$lib/platform';
import * as authApi from '$lib/api/auth';

const GOOGLE_DESKTOP_POLL_INTERVAL_MS = 1200;
const GOOGLE_DESKTOP_TIMEOUT_MS = 2 * 60 * 1000;

async function openGoogleLoginURL(url: string): Promise<void> {
	if (typeof window === 'undefined') {
		throw new Error('Google login is only available in browser contexts.');
	}

	window.open(url, '_blank', 'noopener,noreferrer');
}

interface AuthState {
	user: User | null;
	isAuthenticated: boolean;
	isLoading: boolean;
}

function createAuthStore() {
	let state = $state<AuthState>({
		user: null,
		isAuthenticated: false,
		isLoading: true
	});

	return {
		get state() {
			return state;
		},

		async init() {
			if (!browser) return;

			state.isLoading = true;

			try {
				if (authApi.isAuthenticated()) {
					const { user } = await authApi.getCurrentUser();
					state = {
						user,
						isAuthenticated: true,
						isLoading: false
					};
				} else {
					state = {
						user: null,
						isAuthenticated: false,
						isLoading: false
					};
				}
			} catch {
				authApi.removeAuthToken();
				state = {
					user: null,
					isAuthenticated: false,
					isLoading: false
				};
			}
		},

		async login(username: string, password: string) {
			state.isLoading = true;

			try {
				const response = await authApi.login({ username, password });
				state = {
					user: response.user,
					isAuthenticated: true,
					isLoading: false
				};
				return response;
			} catch (error) {
				state.isLoading = false;
				throw error;
			}
		},

		async register(username: string, password: string) {
			state.isLoading = true;

			try {
				const response = await authApi.register({ username, password });
				state = {
					user: response.user,
					isAuthenticated: true,
					isLoading: false
				};
				return response;
			} catch (error) {
				state.isLoading = false;
				throw error;
			}
		},

		async googleLogin() {
			state.isLoading = true;

			if (!browser) {
				state.isLoading = false;
				throw new Error('Google login is only available in browser contexts.');
			}

			if (isDesktopApp) {
				try {
					const { url, session_id: sessionId } = await authApi.getGoogleAuthUrl('desktop');
					if (!sessionId) {
						throw new Error('Missing desktop OAuth session identifier.');
					}

					await openGoogleLoginURL(url);

					const startedAt = Date.now();
					while (Date.now() - startedAt < GOOGLE_DESKTOP_TIMEOUT_MS) {
						await new Promise((resolve) => setTimeout(resolve, GOOGLE_DESKTOP_POLL_INTERVAL_MS));

						const status = await authApi.getGoogleDesktopAuthStatus(sessionId);
						if (status.status === 'pending') {
							continue;
						}

						if (status.status === 'error') {
							throw new Error(status.error || 'Google sign in failed.');
						}

						const response = status.auth;
						authApi.setAuthToken(response.token);
						state = {
							user: response.user,
							isAuthenticated: true,
							isLoading: false
						};
						return response;
					}

					throw new Error('Google sign in timed out. Please try again.');
				} catch (error) {
					state.isLoading = false;
					throw error;
				}
			}

			const { url } = await authApi.getGoogleAuthUrl('web', window.location.origin);
			window.location.href = url;
		},

		async logout() {
			await authApi.logout();
			state = {
				user: null,
				isAuthenticated: false,
				isLoading: false
			};
		},

		setUser(user: User) {
			state = {
				user,
				isAuthenticated: true,
				isLoading: false
			};
		}
	};
}

export const authStore = createAuthStore();
