<script lang="ts">
	import { isDesktopApp } from '$lib/platform';
	import { onMount } from 'svelte';
	import type * as WailsRuntime from '@wailsio/runtime';

	let isMaximised = $state(false);
	type RuntimeWindow = typeof WailsRuntime.Window;

	async function withWindow<T>(callback: (windowApi: RuntimeWindow) => Promise<T>) {
		if (!isDesktopApp) return;
		const runtimeModule = await import('@wailsio/runtime');
		await callback(runtimeModule.Window);
	}

	async function syncWindowState() {
		await withWindow(async (windowApi) => {
			isMaximised = await windowApi.IsMaximised();
		});
	}

	async function minimiseWindow() {
		await withWindow(async (windowApi) => {
			await windowApi.Minimise();
		});
	}

	async function toggleMaximiseWindow() {
		await withWindow(async (windowApi) => {
			await windowApi.ToggleMaximise();
		});
		await syncWindowState();
	}

	async function closeWindow() {
		await withWindow(async (windowApi) => {
			await windowApi.Close();
		});
	}

	onMount(() => {
		if (!isDesktopApp) return;

		void syncWindowState();
		const onResize = () => void syncWindowState();
		window.addEventListener('resize', onResize);

		return () => {
			window.removeEventListener('resize', onResize);
		};
	});
</script>

<header
	class="desktop-window-bar"
	style="--wails-draggable: drag"
	ondblclick={toggleMaximiseWindow}
	role="presentation"
>
	<div class="desktop-window-brand">
		<span class="desktop-window-logo" aria-hidden="true"></span>
		<span class="desktop-window-title">ThemeSmith</span>
	</div>

	<div class="desktop-window-controls" style="--wails-draggable: no-drag">
		<button
			type="button"
			class="desktop-window-control"
			onclick={minimiseWindow}
			aria-label="Minimize window"
			title="Minimize"
		>
			<svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
				<path d="M2 6h8"></path>
			</svg>
		</button>

		<button
			type="button"
			class="desktop-window-control"
			onclick={toggleMaximiseWindow}
			aria-label={isMaximised ? 'Restore window' : 'Maximize window'}
			title={isMaximised ? 'Restore' : 'Maximize'}
		>
			{#if isMaximised}
				<svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.3" aria-hidden="true">
					<rect x="3" y="4" width="5" height="5"></rect>
					<path d="M4 3h5v5"></path>
				</svg>
			{:else}
				<svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.3" aria-hidden="true">
					<rect x="2.5" y="2.5" width="7" height="7"></rect>
				</svg>
			{/if}
		</button>

		<button
			type="button"
			class="desktop-window-control desktop-window-control-close"
			onclick={closeWindow}
			aria-label="Close window"
			title="Close"
		>
			<svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
				<path d="M3 3l6 6M9 3L3 9"></path>
			</svg>
		</button>
	</div>
</header>

<style>
	.desktop-window-bar {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		height: 46px;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0 10px 0 14px;
		z-index: 90;
		background: linear-gradient(180deg, rgba(24, 24, 27, 0.98) 0%, rgba(17, 17, 19, 0.96) 100%);
		border-bottom: 1px solid rgba(238, 179, 143, 0.28);
	}

	.desktop-window-brand {
		display: flex;
		align-items: center;
		gap: 10px;
		min-width: 0;
	}

	.desktop-window-logo {
		width: 10px;
		height: 10px;
		border-radius: 999px;
		background: radial-gradient(circle at 30% 30%, #ffd2b8, #eeb38f 70%);
		box-shadow: 0 0 12px rgba(238, 179, 143, 0.4);
	}

	.desktop-window-title {
		font-size: 13px;
		font-weight: 600;
		letter-spacing: 0.02em;
		color: rgb(212, 212, 216);
		user-select: none;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.desktop-window-controls {
		display: flex;
		align-items: center;
		gap: 6px;
	}

	.desktop-window-control {
		width: 30px;
		height: 28px;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		border-radius: 8px;
		border: 1px solid transparent;
		background: transparent;
		color: rgb(161, 161, 170);
		cursor: pointer;
		transition:
			background-color 150ms ease,
			color 150ms ease,
			border-color 150ms ease;
	}

	.desktop-window-control:hover {
		background: rgba(39, 39, 42, 0.95);
		border-color: rgba(82, 82, 91, 0.8);
		color: rgb(228, 228, 231);
	}

	.desktop-window-control-close:hover {
		background: rgba(185, 28, 28, 0.9);
		border-color: rgba(248, 113, 113, 0.55);
		color: rgb(255, 255, 255);
	}

	.desktop-window-control svg {
		width: 12px;
		height: 12px;
	}
</style>
