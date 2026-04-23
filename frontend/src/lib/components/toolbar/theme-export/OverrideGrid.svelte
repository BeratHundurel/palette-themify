<script lang="ts">
	import type { ThemeOverrides } from '$lib/types/theme';
	import Select from '$lib/components/ui/Select.svelte';

	type OverrideField = { key: keyof ThemeOverrides; label: string; hint: string };

	type Props = {
		overrideFields: OverrideField[];
		themeOverrides: ThemeOverrides;
		themeResultExists: boolean;
		manualOverrideBadges: Partial<Record<keyof ThemeOverrides, boolean>>;
		getOverrideValue: (key: keyof ThemeOverrides) => string;
		getOverrideRecommendations: (key: keyof ThemeOverrides, currentValue: string) => string[];
		onUpdateOverride: (key: keyof ThemeOverrides, value: string) => void;
		onSwitchOverride: (key: keyof ThemeOverrides, sourceKey: keyof ThemeOverrides) => void;
		onShuffle: () => void;
		onReset: () => void;
	};

	let {
		overrideFields,
		themeOverrides,
		themeResultExists,
		manualOverrideBadges,
		getOverrideValue,
		getOverrideRecommendations,
		onUpdateOverride,
		onSwitchOverride,
		onShuffle,
		onReset
	}: Props = $props();

	const switchableKeys: Array<keyof ThemeOverrides> = [
		'c1',
		'c2',
		'c3',
		'c4',
		'c5',
		'c6',
		'c7',
		'c8',
		'c9',
		'constants'
	];

	function getSwitchableLabel(key: keyof ThemeOverrides): string {
		const field = overrideFields.find((overrideField) => overrideField.key === key);
		if (!field) return String(key).toUpperCase();
		return field.label;
	}
</script>

<div id="tutorial-theme-overrides-root">
	<div class="mb-3 flex items-center justify-between gap-4">
		<div class="flex items-center gap-2">
			<h3 class="text-brand text-sm font-semibold tracking-wide uppercase">Base Color Overrides</h3>
			<span class="text-xs text-zinc-500">Auto-regenerates variants</span>
		</div>
		<div class="flex items-center gap-2">
			<button
				type="button"
				onclick={onShuffle}
				disabled={!themeResultExists}
				class="hover:border-brand/50 rounded border border-zinc-700/50 px-3 py-1.5 text-xs text-zinc-400 transition-colors hover:bg-zinc-800/50 hover:text-zinc-300 disabled:cursor-not-allowed disabled:opacity-40"
				title="Shuffle C1-C9"
			>
				Shuffle
			</button>
			<button
				type="button"
				onclick={onReset}
				disabled={Object.values(themeOverrides).every((value) => value == null)}
				class="hover:border-brand/50 rounded border border-zinc-700/50 px-3 py-1.5 text-xs text-zinc-400 transition-colors hover:bg-zinc-800/50 hover:text-zinc-300 disabled:cursor-not-allowed disabled:opacity-40"
				title="Return to defaults"
			>
				Reset
			</button>
		</div>
	</div>

	<div id="tutorial-theme-overrides-grid" class="grid grid-cols-2 gap-2">
		{#each overrideFields as field (field.key)}
			{@const isModified = manualOverrideBadges[field.key] === true}
			{@const currentValue = getOverrideValue(field.key)}
			{@const recommendations = getOverrideRecommendations(field.key, currentValue)}
			{@const isSwitchableField = switchableKeys.includes(field.key)}
			<div
				class="group rounded-lg border border-zinc-800/80 bg-zinc-900/40 px-3 py-2 transition-colors hover:border-zinc-700/80 hover:bg-zinc-900/60"
			>
				<div class="flex items-center gap-4">
					<div class="relative shrink-0">
						<input
							type="color"
							value={currentValue}
							class="h-8 w-8 cursor-pointer rounded border-0 outline-none"
							oninput={(e) => onUpdateOverride(field.key, (e.target as HTMLInputElement).value)}
							title="Click to pick color"
						/>
						{#if isModified}
							<div class="bg-brand absolute -top-0.5 -right-0.5 h-2 w-2 rounded-full ring-2 ring-zinc-900"></div>
						{/if}
					</div>

					<div class="min-w-0 flex-1">
						<div class="flex min-w-0 items-center gap-2">
							<div class="min-w-0 flex-1">
								<span class="block min-w-0 truncate text-sm font-medium text-zinc-200 {isModified ? 'text-brand' : ''}">
									{field.label}
								</span>
								<span class="mt-0.5 block min-w-0 truncate text-xs text-zinc-500">{field.hint}</span>

								{#if recommendations.length > 0 || isSwitchableField}
									<div class="mt-1.5 flex items-center gap-2">
										{#if recommendations.length > 0}
											<div
												class="inline-flex h-7 min-w-0 items-center gap-1.5 rounded-md border border-zinc-700/60 bg-zinc-800/60 px-2"
											>
												<span class="text-[10px] tracking-wide text-zinc-400 uppercase">Options</span>
												{#each recommendations as recommendation (recommendation)}
													<button
														type="button"
														onclick={() => onUpdateOverride(field.key, recommendation)}
														class="hover:border-brand/60 h-4 w-4 shrink-0 rounded border border-zinc-600/80 shadow-[0_0_0_1px_rgba(0,0,0,0.25)] transition-[transform,border-color,box-shadow] hover:scale-105 hover:shadow-[0_0_0_1px_rgba(244,114,182,0.35)]"
														style={`background-color: ${recommendation};`}
														title={`Use suggested option ${recommendation}`}
														aria-label={`Use suggested option ${recommendation} for ${field.label}`}
													></button>
												{/each}
											</div>
										{/if}

										{#if isSwitchableField}
											<Select
												id={`switch-${field.key}`}
												class="h-7 w-28 shrink-0 rounded border-zinc-700/60 bg-zinc-800/60 px-2 py-0 text-xs text-zinc-400 focus:bg-zinc-800"
												onchange={(event) => {
													const selected = (event.target as HTMLSelectElement).value as keyof ThemeOverrides;
													if (!selected) return;
													onSwitchOverride(field.key, selected);
													(event.target as HTMLSelectElement).value = '';
												}}
											>
												<option value="">Swap</option>
												{#each switchableKeys.filter((key) => key !== field.key) as switchKey (switchKey)}
													<option value={switchKey}>{getSwitchableLabel(switchKey)}</option>
												{/each}
											</Select>
										{/if}
									</div>
								{/if}
							</div>

							<input
								type="text"
								value={currentValue}
								placeholder="#000000"
								class="focus:border-brand/60 mt-0.5 w-28 rounded border border-zinc-700/50 bg-zinc-800/60 px-2 py-1.5 font-mono text-xs text-zinc-300 placeholder-zinc-600 transition-colors focus:bg-zinc-800 focus:outline-none {isModified
									? 'border-brand/40 bg-zinc-800'
									: ''}"
								oninput={(e) => onUpdateOverride(field.key, (e.target as HTMLInputElement).value)}
							/>
						</div>
					</div>
				</div>
			</div>
		{/each}
	</div>
</div>
