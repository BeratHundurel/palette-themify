import { browser } from '$app/environment';
import type { User } from '$lib/api/auth';
import * as authApi from '$lib/api/auth';

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

		async login(email: string, password: string) {
			state.isLoading = true;

			try {
				const response = await authApi.login({ email, password });
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

		async register(name: string, email: string, password: string) {
			state.isLoading = true;

			try {
				const response = await authApi.register({ name, email, password });
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

		async demoLogin() {
			state.isLoading = true;

			try {
				const response = await authApi.demoLogin();
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
			const { url } = await authApi.getGoogleAuthUrl();
			window.location.href = url;
		},

		async handleGoogleCallback(code: string) {
			state.isLoading = true;

			try {
				const response = await authApi.handleGoogleCallback(code);
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

		async logout() {
			await authApi.logout();
			state = {
				user: null,
				isAuthenticated: false,
				isLoading: false
			};
		},

		setUser(user: User) {
			state.user = user;
		},

		isDemoUser(): boolean {
			const email = state.user?.email;
			return email?.startsWith('demo-') === true && email?.endsWith('@imagepalette.com') === true;
		}
	};
}

export const authStore = createAuthStore();
