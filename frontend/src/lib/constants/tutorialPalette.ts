import type { Color, NamedColor } from '$lib/types/color';

export const TUTORIAL_APPLY_NAMED_PALETTE: NamedColor[] = [
	{ name: 'black-rock', hex: '#2c2d3c' },
	{ name: 'legendary-purple', hex: '#4e4e63' },
	{ name: 'little-theatre', hex: '#73778f' },
	{ name: 'frosty-nightfall', hex: '#9497b3' },
	{ name: 'back-in-black', hex: '#16141c' },
	{ name: 'lavender-dream', hex: '#b4aecc' },
	{ name: 'mount-sterling', hex: '#cad3d4' }
];

export const TUTORIAL_APPLY_PALETTE: Color[] = TUTORIAL_APPLY_NAMED_PALETTE.map(({ hex }) => ({ hex }));
