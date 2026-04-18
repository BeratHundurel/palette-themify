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
			'focus:border-brand/60 focus:ring-brand/20 w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-300 transition-[border-color,box-shadow,background-color,color] duration-200 outline-none focus:ring-2 enabled:hover:border-zinc-600 disabled:cursor-not-allowed disabled:opacity-50',
			showChevron ? 'appearance-none pr-9' : '',
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
