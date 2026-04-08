export type ApplyPaletteSettings = {
	luminosity: number;
	nearest: number;
	power: number;
	maxDistance: number;
};

export type NumericConstraint = {
	min: number;
	max: number;
	step: number;
};

export const DEFAULT_APPLY_PALETTE_SETTINGS: ApplyPaletteSettings = {
	luminosity: 1,
	nearest: 30,
	power: 4,
	maxDistance: 0
};

export const APPLY_PALETTE_SETTINGS_CONSTRAINTS = {
	luminosity: { min: 0.1, max: 3, step: 0.1 },
	nearest: { min: 1, max: 30, step: 1 },
	power: { min: 0.5, max: 10, step: 0.5 },
	maxDistance: { min: 0, max: 200, step: 5 }
} as const satisfies Record<keyof ApplyPaletteSettings, NumericConstraint>;
