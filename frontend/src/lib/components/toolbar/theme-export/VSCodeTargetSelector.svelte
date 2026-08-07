<script lang="ts">
	import type { VSCodeFamilyTarget } from '$lib/types/theme';
	import { cn } from '$lib/utils';

	import { VSCODE_FAMILY_TARGETS } from './targets';

	type Props = {
		selected: VSCodeFamilyTarget;
		canInstall: boolean;
		onSelect: (target: VSCodeFamilyTarget) => void;
	};

	let { selected, canInstall, onSelect }: Props = $props();
</script>

<div class="mb-8 rounded-lg border border-zinc-700/70 bg-zinc-900/60 p-4">
	<div class="flex flex-col gap-1 sm:flex-row sm:items-end sm:justify-between">
		<div>
			<h4 class="text-sm font-medium text-zinc-200">Target editor</h4>
			<p class="mt-1 text-xs text-zinc-500">
				{canInstall
					? 'The desktop app installs the theme in your chosen editor.'
					: 'The copied theme uses the same VS Code format in every supported editor.'}
			</p>
		</div>
		<span class="text-brand text-xs font-medium">No conversion needed</span>
	</div>

	<div class="mt-4 grid gap-3 sm:grid-cols-3">
		{#each VSCODE_FAMILY_TARGETS as option (option.value)}
			<button
				type="button"
				onclick={() => onSelect(option.value)}
				class={cn(
					'rounded-lg border px-3 py-3 text-left transition-[background-color,border-color,box-shadow] duration-300',
					selected === option.value
						? 'border-brand bg-brand/10 shadow-brand/10 shadow-md'
						: 'hover:border-brand/50 border-zinc-700 bg-zinc-900 hover:bg-zinc-800/50'
				)}
				aria-pressed={selected === option.value}
			>
				<div class="flex items-center gap-2">
					<span class={cn('h-2 w-2 rounded-full', selected === option.value ? 'bg-brand' : 'bg-zinc-600')}></span>
					<span class={cn('text-sm font-medium', selected === option.value ? 'text-brand' : 'text-zinc-200')}>
						{option.label}
					</span>
				</div>
				<p class="mt-1.5 ml-4 text-xs text-zinc-500">{option.description}</p>
			</button>
		{/each}
	</div>

	{#if selected === 'cursor'}
		<p class="mt-3 text-xs leading-5 text-amber-300/80">
			Cursor applies VS Code themes to its Editor Window. Its separate Agent Window keeps Cursor's own appearance
			settings.
		</p>
	{:else if selected === 'antigravity'}
		<p class="mt-3 text-xs leading-5 text-zinc-500">
			Antigravity IDE uses the VS Code extension format and discovers the local theme after an IDE reload.
		</p>
	{/if}
</div>
