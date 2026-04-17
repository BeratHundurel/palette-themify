import { hexToHsl } from '$lib/colorUtils';
import { COLOR_REGEX } from '$lib/types/color';
import type { Color } from '$lib/types/color';
import type { ThemeOverrides } from '$lib/types/theme';

const HARMONY_OFFSETS = [30, 60, 120, 150, 180, 210, 240, 330] as const;
const MIN_USED_DISTANCE = 52;
const MIN_RESULT_DISTANCE = 34;

function toHexUpper(value: string | null | undefined): string | null {
	if (!value || !COLOR_REGEX.test(value)) return null;
	return value.slice(0, 7).toUpperCase();
}

function hslToHex(h: number, s: number, l: number): string {
	const hueToRgb = (p: number, q: number, tInput: number): number => {
		let t = tInput;
		if (t < 0) t += 1;
		if (t > 1) t -= 1;
		if (t < 1 / 6) return p + (q - p) * 6 * t;
		if (t < 1 / 2) return q;
		if (t < 2 / 3) return p + (q - p) * (2 / 3 - t) * 6;
		return p;
	};

	const normalizedH = ((h % 1) + 1) % 1;
	let r = l;
	let g = l;
	let b = l;

	if (s !== 0) {
		const q = l < 0.5 ? l * (1 + s) : l + s - l * s;
		const p = 2 * l - q;
		r = hueToRgb(p, q, normalizedH + 1 / 3);
		g = hueToRgb(p, q, normalizedH);
		b = hueToRgb(p, q, normalizedH - 1 / 3);
	}

	const toByte = (value: number) => Math.max(0, Math.min(255, Math.round(value * 255)));
	const toHexPart = (value: number) => value.toString(16).padStart(2, '0').toUpperCase();
	return `#${toHexPart(toByte(r))}${toHexPart(toByte(g))}${toHexPart(toByte(b))}`;
}

function rgbDistance(hexA: string, hexB: string): number {
	const parse = (hex: string): [number, number, number] => {
		const r = Number.parseInt(hex.slice(1, 3), 16);
		const g = Number.parseInt(hex.slice(3, 5), 16);
		const b = Number.parseInt(hex.slice(5, 7), 16);
		return [r, g, b];
	};

	const [r1, g1, b1] = parse(hexA);
	const [r2, g2, b2] = parse(hexB);
	const dr = r1 - r2;
	const dg = g1 - g2;
	const db = b1 - b2;
	return Math.sqrt(dr * dr + dg * dg + db * db);
}

function buildHarmonyCandidates(seed: string): string[] {
	const hsl = hexToHsl(seed);
	const saturation = Math.min(0.82, Math.max(0.32, hsl.s * 1.08));
	const lightness = Math.min(0.74, Math.max(0.26, hsl.l));

	return HARMONY_OFFSETS.map((offset) => {
		const hue = hsl.h + offset / 360;
		return hslToHex(hue, saturation, lightness);
	});
}

type RecommendationArgs = {
	key: keyof ThemeOverrides;
	currentValue: string;
	themeOverrides: ThemeOverrides;
	baseOverrides: ThemeOverrides;
	paletteColors: Color[];
	overrideKeys: Array<keyof ThemeOverrides>;
};

export function getRecommendedOverrideColors({
	key,
	currentValue,
	themeOverrides,
	baseOverrides,
	paletteColors,
	overrideKeys
}: RecommendationArgs): string[] {
	if (key === 'background' || key === 'foreground') return [];

	const currentHex = toHexUpper(currentValue);
	if (!currentHex) return [];

	const used: string[] = [];
	for (const overrideKey of overrideKeys) {
		if (overrideKey === key) continue;
		const value = toHexUpper(themeOverrides[overrideKey] ?? baseOverrides[overrideKey]);
		if (value && !used.includes(value)) used.push(value);
	}

	const seeds: string[] = [];
	for (const color of paletteColors) {
		const value = toHexUpper(color.hex);
		if (value && !seeds.includes(value)) seeds.push(value);
	}

	const rankedSeeds = seeds
		.filter((hex) => hex !== currentHex)
		.sort((a, b) => rgbDistance(b, currentHex) - rgbDistance(a, currentHex));

	const pool: string[] = [...buildHarmonyCandidates(currentHex)];
	for (const seed of rankedSeeds.slice(0, 6)) {
		pool.push(...buildHarmonyCandidates(seed));
	}

	const results: string[] = [];
	for (const candidate of pool) {
		const normalized = toHexUpper(candidate);
		if (!normalized || normalized === currentHex) continue;
		if (used.some((value) => rgbDistance(value, normalized) < MIN_USED_DISTANCE)) continue;
		if (results.some((value) => rgbDistance(value, normalized) < MIN_RESULT_DISTANCE)) continue;
		results.push(normalized);
		if (results.length >= 4) break;
	}

	if (results.length < 4) {
		for (const seed of rankedSeeds) {
			if (used.some((value) => rgbDistance(value, seed) < MIN_USED_DISTANCE)) continue;
			if (results.some((value) => rgbDistance(value, seed) < MIN_RESULT_DISTANCE)) continue;
			results.push(seed);
			if (results.length >= 4) break;
		}
	}

	return results;
}
