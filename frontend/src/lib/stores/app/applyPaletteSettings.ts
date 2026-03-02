import { browser } from '$app/environment';
import type { ApplyPaletteSettings } from '$lib/types/applyPaletteSettings';

const APPLY_PALETTE_SETTINGS_KEY = 'applyPaletteSettings';
export const DEFAULT_APPLY_PALETTE_SETTINGS: ApplyPaletteSettings = {
	luminosity: 1,
	nearest: 30,
	power: 4,
	maxDistance: 0
};

export function loadApplyPaletteSettings(): ApplyPaletteSettings {
	if (!browser) return DEFAULT_APPLY_PALETTE_SETTINGS;
	const stored = localStorage.getItem(APPLY_PALETTE_SETTINGS_KEY);
	if (stored) {
		const parsed = JSON.parse(stored);
		return {
			luminosity: parsed.luminosity ?? DEFAULT_APPLY_PALETTE_SETTINGS.luminosity,
			nearest: parsed.nearest ?? DEFAULT_APPLY_PALETTE_SETTINGS.nearest,
			power: parsed.power ?? DEFAULT_APPLY_PALETTE_SETTINGS.power,
			maxDistance: parsed.maxDistance ?? DEFAULT_APPLY_PALETTE_SETTINGS.maxDistance
		};
	}
	return DEFAULT_APPLY_PALETTE_SETTINGS;
}

export function saveApplyPaletteSettings(settings: ApplyPaletteSettings) {
	if (!browser) return;
	localStorage.setItem(APPLY_PALETTE_SETTINGS_KEY, JSON.stringify(settings));
}

export function clearApplyPaletteSettings() {
	if (!browser) return;
	localStorage.removeItem(APPLY_PALETTE_SETTINGS_KEY);
}
