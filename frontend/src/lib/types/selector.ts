export type Selector = {
	id: string;
	color: string;
	selected: boolean;
	selection?: { x: number; y: number; w: number; h: number };
};

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

export const DEFAULT_SELECTOR_ID = 'green';
