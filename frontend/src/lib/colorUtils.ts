import toast from 'svelte-french-toast';
import type { Color } from './types/color';
import type { Theme, ThemeColorWithUsage } from './types/theme';
import type { EditorThemeType } from './api/theme';

export function detectThemeType(theme: Theme): EditorThemeType {
	if ('themes' in theme && Array.isArray((theme as Record<string, unknown>).themes)) return 'zed';
	if ('colors' in theme && 'tokenColors' in theme) return 'vscode';
	return 'zed';
}

export function hexToRgb(hex: string): { r: number; g: number; b: number } {
	if (!hex || typeof hex !== 'string') {
		toast.error('Invalid color value.');
		return { r: 0, g: 0, b: 0 };
	}
	const cleanHex = hex.startsWith('#') ? hex.slice(1) : hex;

	if (cleanHex.length !== 6 || !/^[a-f\d]{6}$/i.test(cleanHex)) {
		toast.error('Invalid color format. Use a hex value like #AABBCC.');
		return { r: 0, g: 0, b: 0 };
	}

	const num = parseInt(cleanHex, 16);
	return {
		r: (num >> 16) & 0xff,
		g: (num >> 8) & 0xff,
		b: num & 0xff
	};
}

export function rgbToHsl(r: number, g: number, b: number): { h: number; s: number; l: number } {
	r /= 255;
	g /= 255;
	b /= 255;

	const max = Math.max(r, g, b);
	const min = Math.min(r, g, b);
	let h = 0;
	let s = 0;
	const l = (max + min) / 2;

	if (max !== min) {
		const d = max - min;
		s = l > 0.5 ? d / (2 - max - min) : d / (max + min);

		switch (max) {
			case r:
				h = ((g - b) / d + (g < b ? 6 : 0)) / 6;
				break;
			case g:
				h = ((b - r) / d + 2) / 6;
				break;
			case b:
				h = ((r - g) / d + 4) / 6;
				break;
		}
	}

	return { h, s, l };
}

export function hexToHsl(hex: string): { h: number; s: number; l: number } {
	const { r, g, b } = hexToRgb(hex);
	return rgbToHsl(r, g, b);
}

export function getLuminance(hex: string): number {
	const { r, g, b } = hexToRgb(hex);

	const rsRGB = r / 255;
	const gsRGB = g / 255;
	const bsRGB = b / 255;

	const rLinear = rsRGB <= 0.03928 ? rsRGB / 12.92 : Math.pow((rsRGB + 0.055) / 1.055, 2.4);
	const gLinear = gsRGB <= 0.03928 ? gsRGB / 12.92 : Math.pow((gsRGB + 0.055) / 1.055, 2.4);
	const bLinear = bsRGB <= 0.03928 ? bsRGB / 12.92 : Math.pow((bsRGB + 0.055) / 1.055, 2.4);

	return 0.2126 * rLinear + 0.7152 * gLinear + 0.0722 * bLinear;
}

export type SortMethod = 'hue' | 'saturation' | 'lightness' | 'luminance' | 'none';

export interface SortResult {
	colors: Array<Color>;
	hadNoChange: boolean;
}

export function sortColorsByMethod(colors: Array<Color>, method: SortMethod): SortResult {
	if (method === 'none') return { colors, hadNoChange: false };

	const sorted = [...colors].sort((a, b) => {
		switch (method) {
			case 'hue': {
				const hslA = hexToHsl(a.hex);
				const hslB = hexToHsl(b.hex);
				return hslA.h - hslB.h;
			}
			case 'saturation': {
				const hslA = hexToHsl(a.hex);
				const hslB = hexToHsl(b.hex);
				return hslB.s - hslA.s;
			}
			case 'lightness': {
				const hslA = hexToHsl(a.hex);
				const hslB = hexToHsl(b.hex);
				return hslA.l - hslB.l;
			}
			case 'luminance': {
				const lumA = getLuminance(a.hex);
				const lumB = getLuminance(b.hex);
				return lumA - lumB;
			}
			default:
				return 0;
		}
	});

	const hadNoChange = checkSortChange(colors, sorted);

	return { colors: sorted, hadNoChange };
}

function checkSortChange(original: Array<Color>, sorted: Array<Color>): boolean {
	if (original.length !== sorted.length || original.length === 0) return false;

	for (let i = 0; i < original.length; i++) {
		if (original[i].hex !== sorted[i].hex) {
			return false;
		}
	}

	return true;
}

const COLOR_REGEX = /^#[0-9a-fA-F]{6}([0-9a-fA-F]{2})?$/i;

export function extractThemeColorsWithUsage(theme: Theme): ThemeColorWithUsage[] {
	const colorMap = new Map<string, Map<string, Set<string>>>();

	function traverse(value: unknown, prefix: string) {
		if (typeof value === 'string' && COLOR_REGEX.test(value)) {
			const normalizedColor = value.toUpperCase();
			const baseColor = normalizedColor.substring(0, 7);

			let variantsMap = colorMap.get(baseColor);
			if (!variantsMap) {
				variantsMap = new Map();
				colorMap.set(baseColor, variantsMap);
			}

			let usagesSet = variantsMap.get(normalizedColor);
			if (!usagesSet) {
				usagesSet = new Set();
				variantsMap.set(normalizedColor, usagesSet);
			}

			usagesSet.add(prefix);
			return;
		}

		if (typeof value !== 'object' || value === null) return;

		if (Array.isArray(value)) {
			for (let i = 0; i < value.length; i++) {
				traverse(value[i], `${prefix}[${i}]`);
			}
			return;
		}

		for (const [key, entry] of Object.entries(value)) {
			traverse(entry, prefix ? `${prefix}.${key}` : key);
		}
	}

	traverse(theme, '');

	const result: ThemeColorWithUsage[] = [];
	result.length = colorMap.size;

	let resultIndex = 0;
	for (const [baseColor, variants] of colorMap) {
		const variantArray = [] as ThemeColorWithUsage['variants'];
		variantArray.length = variants.size;

		let variantIndex = 0;
		let totalUsages = 0;

		for (const [color, usages] of variants) {
			const sortedUsages = [...usages].sort();
			variantArray[variantIndex++] = {
				color,
				usages: sortedUsages
			};
			totalUsages += sortedUsages.length;
		}

		variantArray.sort((a, b) => b.usages.length - a.usages.length);

		result[resultIndex++] = {
			baseColor,
			label: baseColor,
			variants: variantArray,
			totalUsages
		};
	}

	return result;
}
