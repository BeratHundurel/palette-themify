<script lang="ts">
	import { cn } from '$lib/utils';
	import type { ThemeAppearance } from '$lib/types/themeApi';

	type AppearanceOption = {
		value: ThemeAppearance;
		label: string;
		description: string;
		hint: string;
		previewColors: string[];
	};

	const APPEARANCE_OPTIONS: AppearanceOption[] = [
		{
			value: 'dark',
			label: 'Dark',
			description: 'Deep surfaces, bold accents',
			hint: 'Optimized for low-light setups',
			previewColors: [
				'bg-zinc-800',
				'bg-zinc-700',
				'bg-zinc-600',
				'bg-zinc-900',
				'bg-zinc-800',
				'bg-zinc-700',
				'bg-zinc-950',
				'bg-zinc-900',
				'bg-zinc-800'
			]
		},
		{
			value: 'light',
			label: 'Light',
			description: 'Airy surfaces, crisp contrast',
			hint: 'Balanced for daylight work',
			previewColors: [
				'bg-zinc-200',
				'bg-zinc-300',
				'bg-zinc-400',
				'bg-zinc-100',
				'bg-zinc-200',
				'bg-zinc-300',
				'bg-zinc-50',
				'bg-zinc-100',
				'bg-zinc-200'
			]
		}
	];

	type Props = {
		selected: ThemeAppearance;
		onSelect: (appearance: ThemeAppearance) => void;
	};

	let { selected, onSelect }: Props = $props();

	function getCardClass(option: ThemeAppearance): string {
		return cn(
			'group relative overflow-hidden rounded-xl border px-4 py-4 text-left transition-[background-color,border-color,box-shadow] duration-300',
			selected === option
				? option === 'dark'
					? 'border-brand/70 shadow-brand/20 bg-zinc-900/80 shadow-lg'
					: 'border-brand/40 shadow-brand/10 bg-zinc-800/40 shadow-lg'
				: option === 'dark'
					? 'hover:border-brand/50 border-zinc-700 bg-zinc-950/60 hover:bg-zinc-900/60'
					: 'hover:border-brand/40 border-zinc-700 bg-zinc-950/60 hover:bg-zinc-900/40'
		);
	}
</script>

<div class="mb-8">
	<div class="mb-6 flex items-center gap-2">
		<h3 class="text-brand text-sm font-semibold tracking-wide uppercase">Theme Appearance</h3>
		<div class="from-brand/50 h-px flex-1 bg-linear-to-r to-transparent"></div>
	</div>
	<div id="tutorial-theme-appearance-options" class="grid grid-cols-2 gap-4">
		{#each APPEARANCE_OPTIONS as option (option.value)}
			<button type="button" onclick={() => onSelect(option.value)} class={getCardClass(option.value)}>
				<div class="flex items-start justify-between gap-3">
					<div>
						<div class="text-sm font-semibold {selected === option.value ? 'text-brand' : 'text-zinc-200'}">
							{option.label}
						</div>
						<div class="mt-1 text-xs text-zinc-400">{option.description}</div>
					</div>
					<div
						class={cn(
							'rounded-lg border border-zinc-700 p-2',
							option.value === 'dark' ? 'bg-zinc-950/80' : 'bg-zinc-900/60'
						)}
					>
						<div class="grid grid-cols-3 gap-1">
							{#each option.previewColors as previewColor, index (`${option.value}-${index}`)}
								<div class={cn('h-3 w-5 rounded', previewColor)}></div>
							{/each}
						</div>
					</div>
				</div>
				<div class="mt-3 flex items-center gap-2">
					<div
						class={cn(
							'flex h-5 w-5 items-center justify-center rounded-full border transition-[background-color,border-color]',
							selected === option.value
								? option.value === 'dark'
									? 'border-brand bg-brand/20'
									: 'border-brand bg-brand/15'
								: 'border-zinc-500'
						)}
					>
						{#if selected === option.value}
							<div class="bg-brand h-2.5 w-2.5 rounded-full"></div>
						{/if}
					</div>
					<span class="text-xs font-medium text-zinc-400">{option.hint}</span>
				</div>
			</button>
		{/each}
	</div>
</div>
