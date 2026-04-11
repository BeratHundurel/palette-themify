<script lang="ts">
	import { cn } from '$lib/utils';

	type Props = {
		color: string;
		title?: string;
		class?: string;
	};

	let { color, title = color, class: className = '' }: Props = $props();
</script>

<span
	class={cn(
		'inline-block h-8 w-8 cursor-pointer rounded-md border border-zinc-700/50 shadow-md transition-[transform,box-shadow,border-color] duration-300 hover:scale-105 hover:ring-2 hover:ring-white/50',
		className
	)}
	style="background-color: {color}"
	{title}
></span>
