import { browser } from '$app/environment';
import {
	type ApplyPaletteSettings,
	type NumericConstraint,
	DEFAULT_APPLY_PALETTE_SETTINGS,
	APPLY_PALETTE_SETTINGS_CONSTRAINTS
} from '$lib/types/applyPaletteSettings';

const APPLY_PALETTE_SETTINGS_KEY = 'applyPaletteSettings';

function clamp(value: number, constraint: NumericConstraint): number {
	return Math.min(constraint.max, Math.max(constraint.min, value));
}

function asFiniteNumber(value: unknown, fallback: number, constraint: NumericConstraint): number {
	if (typeof value !== 'number' || !Number.isFinite(value)) return fallback;
	return clamp(value, constraint);
}

function asFiniteInteger(value: unknown, fallback: number, constraint: NumericConstraint): number {
	if (typeof value !== 'number' || !Number.isFinite(value)) return fallback;
	return Math.round(clamp(value, constraint));
}

export function parseApplyPaletteSettings(value: unknown): ApplyPaletteSettings {
	if (!value || typeof value !== 'object') return { ...DEFAULT_APPLY_PALETTE_SETTINGS };
	const parsed = value as Partial<ApplyPaletteSettings>;
	return {
		luminosity: asFiniteNumber(
			parsed.luminosity,
			DEFAULT_APPLY_PALETTE_SETTINGS.luminosity,
			APPLY_PALETTE_SETTINGS_CONSTRAINTS.luminosity
		),
		nearest: asFiniteInteger(
			parsed.nearest,
			DEFAULT_APPLY_PALETTE_SETTINGS.nearest,
			APPLY_PALETTE_SETTINGS_CONSTRAINTS.nearest
		),
		power: asFiniteNumber(parsed.power, DEFAULT_APPLY_PALETTE_SETTINGS.power, APPLY_PALETTE_SETTINGS_CONSTRAINTS.power),
		maxDistance: asFiniteNumber(
			parsed.maxDistance,
			DEFAULT_APPLY_PALETTE_SETTINGS.maxDistance,
			APPLY_PALETTE_SETTINGS_CONSTRAINTS.maxDistance
		)
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
