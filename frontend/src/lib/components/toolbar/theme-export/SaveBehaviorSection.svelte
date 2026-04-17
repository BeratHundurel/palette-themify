<script lang="ts">
	type Props = {
		saveOnCopy: boolean;
		accentBoostCoefficient: number;
		onSaveOnCopyChange: (checked: boolean) => void;
		onAccentBoostChange: (value: number) => void;
	};

	let { saveOnCopy, accentBoostCoefficient, onSaveOnCopyChange, onAccentBoostChange }: Props = $props();

	function handleAccentBoostInput(event: Event) {
		const target = event.target as HTMLInputElement;

		const parsed = Number(target.value);
		if (Number.isNaN(parsed)) return;

		const clamped = Math.min(3, Math.max(0, parsed));
		const normalized = Math.round(clamped * 100) / 100;

		onAccentBoostChange(normalized);
	}
</script>

<div class="mb-8">
	<div class="mb-4 flex items-center gap-2">
		<h3 class="text-brand text-sm font-semibold tracking-wide uppercase">Save Behavior</h3>
		<div class="from-brand/50 h-px flex-1 bg-linear-to-r to-transparent"></div>
	</div>
	<div class="space-y-3">
		<label
			class="flex cursor-pointer items-center justify-between rounded-lg border border-zinc-700/50 bg-zinc-900/60 px-4 py-3"
		>
			<div>
				<div class="text-sm font-medium text-zinc-200">Save on copy</div>
				<p class="mt-1 text-xs text-zinc-500">Store the generated theme locally when copying JSON</p>
			</div>
			<input
				id="save-on-copy"
				type="checkbox"
				checked={saveOnCopy}
				onchange={(e) => onSaveOnCopyChange((e.target as HTMLInputElement).checked)}
				class="text-brand focus:ring-brand h-4 w-4 rounded border-zinc-600 bg-zinc-800 focus:ring-2"
			/>
		</label>

		<div class="rounded-lg border border-zinc-700/50 bg-zinc-900/60 px-4 py-3">
			<div class="flex items-start justify-between gap-4">
				<div>
					<div class="text-sm font-medium text-zinc-200">Accent boost coefficient</div>
					<p class="mt-1 text-xs text-zinc-500">
						Set between 0.00 and 3.00 (default 1.00). Boosting mainly affects muted mid-tone accents; very dark/light or
						already vivid colors may not change much. Values above 1.00 push harder, but results can plateau once
						saturation/contrast safety limits are reached.
					</p>
				</div>
			</div>

			<div class="mt-3 flex items-center gap-3">
				<input
					type="range"
					min="0"
					max="3"
					step="0.01"
					value={accentBoostCoefficient}
					oninput={handleAccentBoostInput}
					class="accent-brand h-2 w-full cursor-pointer rounded-lg bg-zinc-800"
				/>
				<input
					type="number"
					min="0"
					max="3"
					step="0.01"
					value={accentBoostCoefficient}
					oninput={handleAccentBoostInput}
					class="focus:border-brand/50 w-24 rounded border border-zinc-700 bg-zinc-900 p-2 text-xs text-zinc-300 transition-[border-color,box-shadow,background-color] duration-300 focus:outline-none"
				/>
			</div>
		</div>
	</div>
</div>
