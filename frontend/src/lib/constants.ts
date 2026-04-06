export const CANVAS = {
	MAX_WIDTH: 800,
	MAX_HEIGHT: 400
} as const;

export const SELECTION = {
	MIN_WIDTH: 1,
	MIN_HEIGHT: 1,
	STROKE_WIDTH: {
		OUTER: 4,
		MIDDLE: 2,
		INNER: 3,
		ANIMATED: 1
	},
	ANIMATION_SPEED: 150,
	DASH_PATTERN: [5, 5] as const
} as const;

export const IMAGE = {
	OUTPUT_FORMAT: 'image/png' as const,
	MAX_FILE_SIZE: 50 * 1024 * 1024
} as const;

export const UI = {
	SORT_BUTTON_PADDING: 5,
	DEFAULT_SELECTOR_ID: 'green'
} as const;

export const COLOR_REGEX = /^#[0-9a-fA-F]{6}([0-9a-fA-F]{2})?$/i;

/**
 * Desktop viewport breakpoints (mobile not supported)
 * - SM: Small desktop (1024px) - tablets in landscape, small laptops
 * - MD: Medium desktop (1440px) - standard monitors
 * - LG: Large desktop (1920px) - full HD displays
 * - XL: Extra large desktop (2560px) - 2K/4K displays
 */
export const DESKTOP_BREAKPOINTS = {
	SM: 1024,
	MD: 1440,
	LG: 1920,
	XL: 2560
} as const;

/**
 * Responsive canvas sizing
 * Returns the appropriate max width for the canvas based on viewport
 */
export function getResponsiveCanvasMaxWidth(): number {
	if (typeof window === 'undefined') return CANVAS.MAX_WIDTH;
	const vw = window.innerWidth;
	// Use 90% of viewport width on smaller screens, max 800px
	return Math.min(CANVAS.MAX_WIDTH, vw * 0.9);
}
