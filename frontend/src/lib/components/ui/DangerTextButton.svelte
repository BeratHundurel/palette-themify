<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLButtonAttributes } from 'svelte/elements';
	import { cn } from '$lib/utils';

	type Props = HTMLButtonAttributes & {
		class?: string;
		children?: Snippet;
	};

	let { class: className = '', type = 'button', children, ...rest }: Props = $props();
</script>

<button
	{...rest}
	{type}
	class={cn(
		'rounded-md px-2 py-1 text-xs font-medium text-red-400 transition-[background-color,color] hover:bg-red-500/10 hover:text-red-300 disabled:cursor-not-allowed disabled:text-zinc-600 disabled:hover:bg-transparent',
		className
	)}
>
	{@render children?.()}
</button>
