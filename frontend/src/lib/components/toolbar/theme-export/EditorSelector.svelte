<script lang="ts">
	import { cn } from '$lib/utils';
	import type { EditorThemeType } from '$lib/api/theme';

	type EditorOption = {
		value: EditorThemeType;
		label: string;
		description: string;
	};

	const EDITOR_OPTIONS: EditorOption[] = [
		{ value: 'vscode', label: 'VS Code', description: 'Generate theme for Visual Studio Code' },
		{ value: 'zed', label: 'Zed', description: 'Generate theme for Zed editor' }
	];

	type Props = {
		selected: EditorThemeType;
		onSelect: (type: EditorThemeType) => void;
	};

	let { selected, onSelect }: Props = $props();

	function getCardClass(option: EditorThemeType): string {
		return cn(
			'group relative overflow-hidden rounded-lg border px-4 py-3 text-left transition-[background-color,border-color,box-shadow] duration-300',
			selected === option
				? 'border-brand bg-brand/10 shadow-brand/20 shadow-lg'
				: 'hover:border-brand/50 border-zinc-600 hover:bg-zinc-800/50'
		);
	}
</script>

<div class="mb-8">
	<div class="mb-6 flex items-center gap-2">
		<h3 class="text-brand text-sm font-semibold tracking-wide uppercase">Editor Type</h3>
		<div class="from-brand/50 h-px flex-1 bg-linear-to-r to-transparent"></div>
	</div>
	<div id="tutorial-theme-editor-options" class="grid grid-cols-2 gap-6">
		{#each EDITOR_OPTIONS as option (option.value)}
			<button type="button" onclick={() => onSelect(option.value)} class={getCardClass(option.value)}>
				<div class="flex items-center gap-2">
					<div
						class={cn(
							'flex h-5 w-5 items-center justify-center rounded-full border transition-[background-color,border-color]',
							selected === option.value ? 'border-brand bg-brand/20' : 'border-zinc-500'
						)}
					>
						{#if selected === option.value}
							<div class="bg-brand h-2.5 w-2.5 rounded-full"></div>
						{/if}
					</div>
					<span class="text-sm font-medium {selected === option.value ? 'text-brand' : 'text-zinc-200'}">
						{option.label}
					</span>
				</div>
				<p class="mt-1.5 ml-7 text-xs text-zinc-400">{option.description}</p>
			</button>
		{/each}
	</div>
</div>
