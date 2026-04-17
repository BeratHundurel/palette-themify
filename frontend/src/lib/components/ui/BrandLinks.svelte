<script lang="ts">
	import type { Snippet } from 'svelte';
	import { resolve } from '$app/paths';
	import { cn } from '$lib/utils';

	type InternalRoute = '/' | '/shared';

	type Props = {
		class?: string;
		children?: Snippet;
		href: InternalRoute;
	};

	let { class: className = '', href, children }: Props = $props();
</script>

{#if href === '/'}
	<a
		href={resolve('/')}
		class={cn(
			'border-brand/50 hover:shadow-brand-lg inline-flex cursor-pointer items-center justify-center rounded-md border bg-zinc-900 px-3 py-2 text-sm font-medium transition-[background-color,border-color,box-shadow,color] duration-300',
			className
		)}
	>
		{@render children?.()}
	</a>
{:else}
	<a
		href={resolve('/shared')}
		class={cn(
			'border-brand/50 hover:shadow-brand-lg inline-flex cursor-pointer items-center justify-center rounded-md border bg-zinc-900 px-3 py-2 text-sm font-medium transition-[background-color,border-color,box-shadow,color] duration-300',
			className
		)}
	>
		{@render children?.()}
	</a>
{/if}
