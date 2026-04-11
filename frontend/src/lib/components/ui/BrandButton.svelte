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
		'border-brand/50 hover:shadow-brand-lg cursor-pointer rounded-md border bg-zinc-900 py-2 text-sm font-medium transition-[background-color,border-color,box-shadow,color] duration-300',
		className
	)}
>
	{@render children?.()}
</button>
