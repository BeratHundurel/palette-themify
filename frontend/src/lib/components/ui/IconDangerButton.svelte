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
		'flex cursor-pointer items-center gap-1 rounded-md p-1.5 text-zinc-500 transition-[transform,background-color,color] hover:scale-110 hover:bg-red-500/10 hover:text-red-400',
		className
	)}
>
	{@render children?.()}
</button>
