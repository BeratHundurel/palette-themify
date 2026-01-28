<script lang="ts">
	import { appStore } from '$lib/stores/app.svelte';
	import toast from 'svelte-french-toast';

	function generateCoolorsUrl(colors: Array<{ hex: string }>): string {
		const hexValues = colors.map((c) => c.hex.replace('#', '')).join('-');
		return `https://coolors.co/${hexValues}`;
	}

	function handleOpenInCoolors() {
		if (appStore.state.colors.length === 0) {
			toast.error('Extract a palette before sharing.');
			return;
		}

		const colorsToUse = appStore.state.colors.length > 10 ? appStore.state.colors.slice(0, 10) : appStore.state.colors;

		if (appStore.state.colors.length > 10) {
			toast.error('Coolors supports up to 10 colors. Using the first 10.');
		}

		const url = generateCoolorsUrl(colorsToUse);
		window.open(url, '_blank');

		if (appStore.state.colors.length <= 10) {
			toast.success('Opening in Coolors.co');
		}
	}
</script>

<button
	class="toolbar-button-base"
	aria-label="Open in Coolors.co"
	onclick={handleOpenInCoolors}
	type="button"
	title="Open palette in Coolors.co"
>
	<svg
		xmlns="http://www.w3.org/2000/svg"
		height="18px"
		viewBox="0 0 24 24"
		width="18px"
		fill="currentColor"
		class="hover:text-brand text-zinc-400 transition-colors duration-300"
	>
		<path d="M0 0h24v24H0z" fill="none" />
		<path
			d="M12 3c-4.97 0-9 4.03-9 9s4.03 9 9 9c.83 0 1.5-.67 1.5-1.5 0-.39-.15-.74-.39-1.01-.23-.26-.38-.61-.38-.99 0-.83.67-1.5 1.5-1.5H16c2.76 0 5-2.24 5-5 0-4.42-4.03-8-9-8zm-5.5 9c-.83 0-1.5-.67-1.5-1.5S5.67 9 6.5 9 8 9.67 8 10.5 7.33 12 6.5 12zm3-4C8.67 8 8 7.33 8 6.5S8.67 5 9.5 5s1.5.67 1.5 1.5S10.33 8 9.5 8zm5 0c-.83 0-1.5-.67-1.5-1.5S13.67 5 14.5 5s1.5.67 1.5 1.5S15.33 8 14.5 8zm3 4c-.83 0-1.5-.67-1.5-1.5S16.67 9 17.5 9s1.5.67 1.5 1.5-.67 1.5-1.5 1.5z"
		/>
	</svg>
</button>
