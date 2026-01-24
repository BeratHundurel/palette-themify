<script lang="ts">
	import { popoverStore } from '$lib/stores/popovers.svelte';
</script>

<button
	class="toolbar-button-base"
	aria-label="How your saved palettes are applied to current image"
	title="How your saved palettes are applied to current image"
	onclick={(e) => popoverStore.toggle('application', e)}
	type="button"
>
	<svg
		xmlns="http://www.w3.org/2000/svg"
		height="18px"
		viewBox="0 -960 960 960"
		width="18px"
		fill="currentColor"
		class="hover:text-brand text-zinc-400 transition-colors duration-300"
	>
		<path
			d="M360-120v-80h440v80H360Zm0-320v-80h440v80H360Zm0-320v-80h440v80H360ZM200-200q-33 0-56.5-23.5T120-280q0-33 23.5-56.5T200-360q33 0 56.5 23.5T280-280q0 33-23.5 56.5T200-200Zm0-320q-33 0-56.5-23.5T120-600q0-33 23.5-56.5T200-680q33 0 56.5 23.5T280-600q0 33-23.5 56.5T200-520Zm0-320q-33 0-56.5-23.5T120-920q0-33 23.5-56.5T200-1000q33 0 56.5 23.5T280-920q0 33-23.5 56.5T200-840Z"
		/>
	</svg>
</button>
