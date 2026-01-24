export const CANVAS = {
	MAX_WIDTH: 800,
	MAX_HEIGHT: 400,
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
