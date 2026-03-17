<script lang="ts">
	import type { Color, NamedColor } from '$lib/types/color';
	import { cn } from '$lib/utils';
	import toast from 'svelte-french-toast';
	import { appStore } from '$lib/stores/app.svelte';
	import { popoverStore } from '$lib/stores/popovers.svelte';

	const copy_options = [
		{ label: 'JSON', value: 'json' },
		{ label: 'CSS Variables', value: 'css_variables' },
		{ label: 'SCSS Variables', value: 'scss_variables' },
		{ label: 'Less Variables', value: 'less_variables' },
		{ label: 'Tailwind Config', value: 'tailwind_config' },
		{ label: 'Bootstrap Variables', value: 'bootstrap_variables' }
	];

	function handleCopyFormatChange(format: string) {
		if (appStore.state.colors.length > 0) {
			copyPaletteAs(format, appStore.state.colors);
			popoverStore.close('copy');
		} else {
			toast.error('Extract a palette before copying.');
		}
	}

	async function copyPaletteAs(format: string, palette: Color[]) {
		let output = '';
		const hexValues = palette.map((c) => c.hex);
		const namedPalette = await getNamedPalette(hexValues);
		switch (format) {
			case 'json':
				output = JSON.stringify(namedPalette, null, 2);
				break;
			case 'css_variables':
				output = namedPalette.map((c) => `--color-${c.name}: ${c.hex};`).join('\n');
				break;
			case 'scss_variables':
				output = generateScssVariables(namedPalette);
				break;
			case 'less_variables':
				output = generateLessVariables(namedPalette);
				break;
			case 'tailwind_config':
				output = generateTailwindThemeBlock(namedPalette);
				break;
			case 'bootstrap_variables':
				output = generateBootstrapVariables(namedPalette);
				break;
		}
		navigator.clipboard.writeText(output);
		toast.success(`${format.replace('_', ' ').toUpperCase()} copied to clipboard`);
	}

	async function getNamedPalette(hexValues: string[]): Promise<NamedColor[]> {
		try {
			const url = `https://api.color.pizza/v1/?values=${hexValues.map((h) => h.replace('#', '')).join(',')}`;
			const res = await fetch(url);
			if (!res.ok) return [];
			const data = await res.json();
			return data.colors.map((c: NamedColor) => ({
				name: slugifyName(c.name),
				hex: c.hex.toLowerCase()
			}));
		} catch {
			return hexValues.map((hex, i) => ({
				name: `color-${i + 1}`,
				hex: hex.toLowerCase()
			}));
		}
	}

	function generateTailwindThemeBlock(colors: NamedColor[]) {
		return `@theme {\n${colors.map((c) => `  --color-${c.name}: ${c.hex};`).join('\n')}\n}`;
	}

	function generateBootstrapVariables(colors: NamedColor[]) {
		return `:root {\n${colors.map((c) => `  --bs-${c.name}: ${c.hex};`).join('\n')}\n}`;
	}

	function generateScssVariables(colors: NamedColor[]) {
		const variables = colors.map((c) => `$color-${c.name}: ${c.hex};`).join('\n');
		return `// SCSS Variables\n${variables}`;
	}

	function generateLessVariables(colors: NamedColor[]) {
		const variables = colors.map((c) => `@color-${c.name}: ${c.hex};`).join('\n');
		return `// Less Variables\n${variables}`;
	}

	function slugifyName(name: string): string {
		return name
			.toLowerCase()
			.replace(/[^a-z0-9]+/g, '-')
			.replace(/^-+|-+$/g, '');
	}
</script>

<div
	class={cn(
		'palette-dropdown-base w-80',
		popoverStore.state.direction === 'right' ? 'left-full ml-2' : 'right-full mr-2'
	)}
	style={popoverStore.state.direction === 'right' ? 'left: calc(100% + 0.5rem);' : 'right: calc(100% + 0.5rem);'}
>
	<p class="text-brand mb-2 text-xs font-medium">Copy Palette</p>

	{#each copy_options as option (option.value)}
		<button
			class="text-brand hover:bg-brand mb-2 w-full cursor-pointer rounded bg-zinc-800 px-2 py-1 text-xs font-medium hover:text-zinc-900"
			onclick={() => handleCopyFormatChange(option.value)}
			type="button"
		>
			{option.label}
		</button>
	{/each}
</div>
