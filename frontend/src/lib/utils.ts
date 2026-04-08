import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { CANVAS } from './types/canvas';

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

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
