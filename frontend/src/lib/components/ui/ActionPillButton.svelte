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
		'text-brand hover:bg-brand/10 flex cursor-pointer items-center gap-1.5 rounded-md px-2.5 text-xs font-medium transition-[transform,background-color] hover:scale-105',
		className
	)}
>
	{@render children?.()}
</button>
