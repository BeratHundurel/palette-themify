export type Color = {
	hex: string;
};

export type NamedColor = {
	name: string;
	hex: string;
};

export const COLOR_REGEX = /^#[0-9a-fA-F]{6}([0-9a-fA-F]{2})?$/i;
