<script lang="ts">
	import type { ThemeOverrides } from '$lib/types/theme';

	type OverrideField = { key: keyof ThemeOverrides; label: string; hint: string };

	type Props = {
		overrideFields: OverrideField[];
		themeOverrides: ThemeOverrides;
		themeResultExists: boolean;
		manualOverrideBadges: Partial<Record<keyof ThemeOverrides, boolean>>;
		getOverrideValue: (key: keyof ThemeOverrides) => string;
		getOverrideRecommendations: (key: keyof ThemeOverrides, currentValue: string) => string[];
		onUpdateOverride: (key: keyof ThemeOverrides, value: string) => void;
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
		onShuffle,
		onReset
	}: Props = $props();
</script>

<div>
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

	<div class="grid grid-cols-2 gap-2">
		{#each overrideFields as field (field.key)}
			{@const isModified = manualOverrideBadges[field.key] === true}
			{@const currentValue = getOverrideValue(field.key)}
			{@const recommendations = getOverrideRecommendations(field.key, currentValue)}
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

								{#if recommendations.length > 0}
									<div
										class="mt-1.5 inline-flex items-center gap-1.5 rounded-md border border-zinc-700/50 bg-zinc-900/60 px-2 py-1"
									>
										<span class="text-[10px] text-zinc-500">Options</span>
										{#each recommendations as recommendation (recommendation)}
											<button
												type="button"
												onclick={() => onUpdateOverride(field.key, recommendation)}
												class="hover:border-brand/50 h-5 w-5 rounded border border-zinc-700/70 transition-colors hover:scale-105"
												style={`background-color: ${recommendation};`}
												title={`Use suggested option ${recommendation}`}
												aria-label={`Use suggested option ${recommendation} for ${field.label}`}
											></button>
										{/each}
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
