<script lang="ts">
	import { authStore } from '$lib/stores/auth.svelte';
	import { appStore } from '$lib/stores/app.svelte';
	import toast from 'svelte-french-toast';

	let { isOpen = $bindable(false) } = $props();

	let mode = $state<'login' | 'register'>('login');
	let loading = $state(false);

	let formData = $state({
		name: '',
		email: '',
		password: '',
		confirmPassword: ''
	});

	let errors = $state<Record<string, string>>({});
	let touched = $state<Record<string, boolean>>({});

	function validate() {
		const newErrors: Record<string, string> = {};

		if (mode === 'register' && !formData.name.trim()) {
			newErrors.name = 'Name is required';
		}

		if (!formData.email) {
			newErrors.email = 'Email is required';
		} else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
			newErrors.email = 'Invalid email format';
		}

		if (!formData.password) {
			newErrors.password = 'Password is required';
		} else if (formData.password.length < 8) {
			newErrors.password = 'Password must be at least 8 characters';
		}

		if (mode === 'register' && formData.password !== formData.confirmPassword) {
			newErrors.confirmPassword = 'Passwords do not match';
		}

		errors = newErrors;
		return Object.keys(newErrors).length === 0;
	}

	function handleBlur(field: string) {
		touched[field] = true;
		validate();
	}

	function handleInput(field: string) {
		if (touched[field] && errors[field]) {
			delete errors[field];
		}
	}

	async function handleSubmit() {
		touched = { email: true, password: true, name: true, confirmPassword: true };

		if (!validate()) return;

		loading = true;

		try {
			if (mode === 'login') {
				await authStore.login(formData.email, formData.password);
				toast.success('Login successful!');
			} else {
				await authStore.register(formData.name, formData.email, formData.password);
				toast.success('Account created successfully!');
			}

			await appStore.syncPalettesOnAuth();
			await appStore.syncWorkspacesOnAuth();

			resetForm();
			isOpen = false;
		} catch (error) {
			const message = error instanceof Error ? error.message : 'An error occurred';
			toast.error(message);
		} finally {
			loading = false;
		}
	}

	async function handleDemoLogin() {
		loading = true;

		try {
			await authStore.demoLogin();
			toast.success('Welcome to the demo!');

			await appStore.syncPalettesOnAuth();
			await appStore.syncWorkspacesOnAuth();

			resetForm();
			isOpen = false;
		} catch (error) {
			const message = error instanceof Error ? error.message : 'An error occurred';
			toast.error(message);
		} finally {
			loading = false;
		}
	}

	function resetForm() {
		formData = {
			name: '',
			email: '',
			password: '',
			confirmPassword: ''
		};
		errors = {};
		touched = {};
	}

	function switchMode() {
		mode = mode === 'login' ? 'register' : 'login';
		resetForm();
	}

	function handleClose() {
		resetForm();
		isOpen = false;
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape' && isOpen) {
			handleClose();
		}
	}

	$effect(() => {
		if (isOpen) {
			document.addEventListener('keydown', handleKeydown);
			return () => document.removeEventListener('keydown', handleKeydown);
		}
	});
</script>

{#if isOpen}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/70"
		role="dialog"
		aria-modal="true"
		aria-labelledby="auth-modal-title"
		tabindex="-1"
		onclick={(e) => e.target === e.currentTarget && handleClose()}
		onkeydown={(e) => e.key === 'Enter' && e.target === e.currentTarget && handleClose()}
	>
		<div class="relative mx-4 w-full max-w-md">
			<div class="border-brand/50 rounded-xl border bg-zinc-900 p-6 text-zinc-300">
				<button
					type="button"
					class="absolute top-4 right-4 cursor-pointer text-zinc-400 transition-colors hover:text-zinc-300"
					onclick={handleClose}
					aria-label="Close modal"
				>
					<svg class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>

				<div class="mb-6">
					<h2 id="auth-modal-title" class="text-2xl font-semibold text-zinc-300">
						{mode === 'login' ? 'Sign In' : 'Create Account'}
					</h2>
					<p class="mt-2 text-sm text-zinc-400">
						{mode === 'login' ? 'Welcome back!' : 'Join us to save your palettes'}
					</p>
				</div>

				<form onsubmit={handleSubmit} class="space-y-4">
					{#if mode === 'register'}
						<div>
							<label for="name" class="mb-1 block text-sm font-medium text-zinc-300"> Name </label>
							<input
								type="text"
								id="name"
								bind:value={formData.name}
								class="focus:border-brand w-full rounded-md border border-zinc-600 bg-zinc-800 px-3 py-2 text-zinc-300 placeholder:text-zinc-400 focus:outline-none"
								class:border-red-500={touched.name && errors.name}
								placeholder="Enter your name"
								disabled={loading}
								onblur={() => handleBlur('name')}
								oninput={() => handleInput('name')}
							/>
							{#if touched.name && errors.name}
								<p class="mt-1 text-xs text-red-400">{errors.name}</p>
							{/if}
						</div>
					{/if}

					<div>
						<label for="email" class="mb-1 block text-sm font-medium text-zinc-300"> Email </label>
						<input
							type="email"
							id="email"
							bind:value={formData.email}
							class="focus:border-brand w-full rounded-md border border-zinc-600 bg-zinc-800 px-3 py-2 text-zinc-300 placeholder:text-zinc-400 focus:outline-none"
							class:border-red-500={touched.email && errors.email}
							placeholder="Enter your email"
							disabled={loading}
							onblur={() => handleBlur('email')}
							oninput={() => handleInput('email')}
						/>
						{#if touched.email && errors.email}
							<p class="mt-1 text-xs text-red-400">{errors.email}</p>
						{/if}
					</div>

					<div>
						<label for="password" class="mb-1 block text-sm font-medium text-zinc-300"> Password </label>
						<input
							type="password"
							id="password"
							bind:value={formData.password}
							class="focus:border-brand w-full rounded-md border border-zinc-600 bg-zinc-800 px-3 py-2 text-zinc-300 placeholder:text-zinc-400 focus:outline-none"
							class:border-red-500={touched.password && errors.password}
							placeholder="Enter your password"
							disabled={loading}
							onblur={() => handleBlur('password')}
							oninput={() => handleInput('password')}
						/>
						{#if touched.password && errors.password}
							<p class="mt-1 text-xs text-red-400">{errors.password}</p>
						{/if}
					</div>

					{#if mode === 'register'}
						<div>
							<label for="confirmPassword" class="mb-1 block text-sm font-medium text-zinc-300">
								Confirm Password
							</label>
							<input
								type="password"
								id="confirmPassword"
								bind:value={formData.confirmPassword}
								class="focus:border-brand w-full rounded-md border border-zinc-600 bg-zinc-800 px-3 py-2 text-zinc-300 placeholder:text-zinc-400 focus:outline-none"
								class:border-red-500={touched.confirmPassword && errors.confirmPassword}
								placeholder="Confirm your password"
								disabled={loading}
								onblur={() => handleBlur('confirmPassword')}
								oninput={() => handleInput('confirmPassword')}
							/>
							{#if touched.confirmPassword && errors.confirmPassword}
								<p class="mt-1 text-xs text-red-400">{errors.confirmPassword}</p>
							{/if}
						</div>
					{/if}

					<button
						type="submit"
						disabled={loading}
						class="hover:shadow-brand-lg hover:bg-brand-hover bg-brand w-full cursor-pointer rounded-md py-2 font-medium text-zinc-900 transition-[background-color,box-shadow] focus:outline-none disabled:cursor-not-allowed disabled:opacity-50"
					>
						{#if loading}
							<div class="flex items-center justify-center">
								<svg
									class="mr-3 -ml-1 h-5 w-5 animate-spin text-zinc-300"
									xmlns="http://www.w3.org/2000/svg"
									fill="none"
									viewBox="0 0 24 24"
								>
									<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
									<path
										class="opacity-75"
										fill="currentColor"
										d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
									></path>
								</svg>
								Processing...
							</div>
						{:else}
							{mode === 'login' ? 'Sign In' : 'Create Account'}
						{/if}
					</button>
				</form>

				{#if mode === 'login'}
					<div class="mt-4">
						<div class="relative">
							<div class="absolute inset-0 flex items-center">
								<div class="w-full border-t border-zinc-600"></div>
							</div>
							<div class="relative flex justify-center text-sm">
								<span class="bg-zinc-900 px-2 text-zinc-400">or</span>
							</div>
						</div>

						<div class="mt-4 rounded-md border border-zinc-600 bg-zinc-800 p-3">
							<div class="mb-2 flex items-center space-x-2">
								<svg class="h-4 w-4 text-zinc-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
									/>
								</svg>
								<span class="text-sm font-medium text-zinc-300">Try Without Signing Up</span>
							</div>
							<p class="mb-3 text-xs text-zinc-400">
								Explore all features with pre-loaded sample palettes. Perfect for testing the app!
							</p>
							<button
								type="button"
								onclick={handleDemoLogin}
								disabled={loading}
								class="transition- hover:shadow-brand hover:border-brand/50 w-full cursor-pointer rounded-md bg-zinc-900 py-3 font-medium text-zinc-300 hover:border focus:outline-none disabled:cursor-not-allowed disabled:opacity-50"
							>
								{#if loading}
									<div class="flex items-center justify-center">
										<svg
											class="mr-3 -ml-1 h-5 w-5 animate-spin text-zinc-300"
											xmlns="http://www.w3.org/2000/svg"
											fill="none"
											viewBox="0 0 24 24"
										>
											<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
											<path
												class="opacity-75"
												fill="currentColor"
												d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
											></path>
										</svg>
										Processing...
									</div>
								{:else}
									<div class="flex items-center justify-center space-x-2">
										<span>🎨</span>
										<span>Launch Demo</span>
									</div>
								{/if}
							</button>
						</div>
					</div>
				{/if}

				<div class="mt-6 text-center">
					<p class="text-sm text-zinc-400">
						{mode === 'login' ? "Don't have an account?" : 'Already have an account?'}
						<button
							type="button"
							class="text-brand hover:text-brand-hover ml-1 cursor-pointer font-medium transition-colors"
							onclick={switchMode}
							disabled={loading}
						>
							{mode === 'login' ? 'Sign up' : 'Sign in'}
						</button>
					</p>
				</div>
			</div>
		</div>
	</div>
{/if}

<style>
	input:-webkit-autofill,
	input:-webkit-autofill:hover,
	input:-webkit-autofill:focus,
	input:-webkit-autofill:active {
		-webkit-box-shadow: 0 0 0 1000px rgb(39, 39, 42) inset !important;
		-webkit-text-fill-color: rgb(212, 212, 216) !important;
		box-shadow: 0 0 0 1000px rgb(39, 39, 42) inset !important;
	}
</style>
