<script lang="ts">
	import '../app.css';
	import DesktopWindowBar from '$lib/components/desktop/DesktopWindowBar.svelte';
	import { isDesktopApp } from '$lib/platform';

	let { children } = $props();
</script>

{#if isDesktopApp}
	<DesktopWindowBar />
{/if}

<div class={isDesktopApp ? 'desktop-app-content' : undefined}>
	{@render children()}
</div>
