<script lang="ts">
	import '../app.css';
	import DesktopWindowBar from '$lib/components/desktop/DesktopWindowBar.svelte';
	import { isDesktopApp, isMacDesktop } from '$lib/platform';

	let { children } = $props();
</script>

{#if isDesktopApp && !isMacDesktop}
	<DesktopWindowBar />
{/if}

<div class={isDesktopApp && !isMacDesktop ? 'desktop-app-content' : undefined}>
	{@render children()}
</div>
