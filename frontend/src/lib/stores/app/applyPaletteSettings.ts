import { browser } from '$app/environment';
import type { ApplyPaletteSettings } from '$lib/types/applyPaletteSettings';

const APPLY_PALETTE_SETTINGS_KEY = 'applyPaletteSettings';
export const DEFAULT_APPLY_PALETTE_SETTINGS: ApplyPaletteSettings = {
	luminosity: 1,
	nearest: 30,
	power: 4,
	maxDistance: 0
};

function asNumber(value: unknown, fallback: number): number {
	return typeof value === 'number' ? value : fallback;
}

export function parseApplyPaletteSettings(value: unknown): ApplyPaletteSettings {
	if (!value || typeof value !== 'object') return { ...DEFAULT_APPLY_PALETTE_SETTINGS };
	const parsed = value as Partial<ApplyPaletteSettings>;
	return {
		luminosity: asNumber(parsed.luminosity, DEFAULT_APPLY_PALETTE_SETTINGS.luminosity),
		nearest: asNumber(parsed.nearest, DEFAULT_APPLY_PALETTE_SETTINGS.nearest),
		power: asNumber(parsed.power, DEFAULT_APPLY_PALETTE_SETTINGS.power),
		maxDistance: asNumber(parsed.maxDistance, DEFAULT_APPLY_PALETTE_SETTINGS.maxDistance)
	};
}

export function loadApplyPaletteSettings(): ApplyPaletteSettings {
	if (!browser) return { ...DEFAULT_APPLY_PALETTE_SETTINGS };
	const stored = localStorage.getItem(APPLY_PALETTE_SETTINGS_KEY);
	if (stored) {
		try {
			return parseApplyPaletteSettings(JSON.parse(stored));
		} catch {
			return { ...DEFAULT_APPLY_PALETTE_SETTINGS };
		}
	}
	return { ...DEFAULT_APPLY_PALETTE_SETTINGS };
}

export function saveApplyPaletteSettings(settings: ApplyPaletteSettings) {
	if (!browser) return;
	localStorage.setItem(APPLY_PALETTE_SETTINGS_KEY, JSON.stringify(settings));
}

export function clearApplyPaletteSettings() {
	if (!browser) return;
	localStorage.removeItem(APPLY_PALETTE_SETTINGS_KEY);
}
