<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLSelectAttributes } from 'svelte/elements';
	import { cn } from '$lib/utils';

	type Props = HTMLSelectAttributes & {
		class?: string;
		children?: Snippet;
		showChevron?: boolean;
		value?: string | string[];
	};

	let { class: className = '', children, showChevron = true, value = $bindable(''), ...rest }: Props = $props();
</script>

<div class="relative">
	<select
		{...rest}
		bind:value
		class={cn(
			showChevron ? 'appearance-none pr-9' : '',
			'transition-[border-color,box-shadow,background-color]',
			className
		)}
	>
		{@render children?.()}
	</select>
	{#if showChevron}
		<svg
			class="pointer-events-none absolute top-1/2 right-2.5 h-3.5 w-3.5 -translate-y-1/2 text-zinc-400"
			viewBox="0 0 20 20"
			fill="none"
			stroke="currentColor"
			stroke-width="1.8"
			aria-hidden="true"
		>
			<path d="M5 7.5L10 12.5L15 7.5" stroke-linecap="round" stroke-linejoin="round" />
		</svg>
	{/if}
</div>
