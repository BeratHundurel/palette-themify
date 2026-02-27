export type PopoverName =
	| 'palette'
	| 'saved'
	| 'copy'
	| 'application'
	| 'themeExport'
	| 'wallhaven'
	| 'themes';
export type Direction = 'left' | 'right';

interface PopoverState {
	current: PopoverName | null;
	direction: Direction;
}

function createPopoverStore() {
	const state = $state<PopoverState>({
		current: null,
		direction: 'right'
	});

	let clickHandler: ((e: Event) => void) | null = null;

	function computeDirection(anchor?: HTMLElement | EventTarget | null): Direction {
		if (typeof window === 'undefined' || !anchor) return 'right';

		const el = anchor as HTMLElement;
		const rect = el.getBoundingClientRect?.();
		if (!rect) return 'right';

		const spaceRight = window.innerWidth - rect.right;
		const popoverWidth = 320;

		return spaceRight >= popoverWidth ? 'right' : 'left';
	}

	function setupClickOutside() {
		if (typeof window === 'undefined' || !state.current) return;

		cleanup();

		clickHandler = (e: Event) => {
			const target = e.target as HTMLElement;
			const isInsidePopover = target.closest('.palette-dropdown-base');
			const isInsideButton = target.closest('.toolbar-button-base');
			const isInsideModal = target.closest('.share-modal-content') || target.closest('.share-modal-backdrop');

			if (!isInsidePopover && !isInsideButton && !isInsideModal) {
				state.current = null;
			}
		};

		setTimeout(() => {
			if (clickHandler) {
				document.addEventListener('click', clickHandler, true);
			}
		}, 10);
	}

	function cleanup() {
		if (clickHandler) {
			document.removeEventListener('click', clickHandler, true);
			clickHandler = null;
		}
	}

	return {
		get state() {
			return state;
		},

		toggle(name: PopoverName, e: MouseEvent) {
			e.stopPropagation();

			if (state.current === name) {
				state.current = null;
				cleanup();
			} else {
				state.current = name;
				state.direction = computeDirection(e.currentTarget);
				setupClickOutside();
			}
		},

		close(name?: PopoverName) {
			if (!name || state.current === name) {
				state.current = null;
				cleanup();
			}
		},

		isOpen(name: PopoverName) {
			return state.current === name;
		}
	};
}

export const popoverStore = createPopoverStore();

if (typeof window !== 'undefined') {
	window.addEventListener('beforeunload', () => popoverStore.close());
}
