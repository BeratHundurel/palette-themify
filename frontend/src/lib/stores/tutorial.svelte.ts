import { appStore } from './app.svelte';
import { popoverStore } from './popovers.svelte';
import { UI } from '$lib/constants';
import { TUTORIAL_APPLY_PALETTE } from '$lib/constants/tutorialPalette';
import type { Color } from '$lib/types/color';

export interface TutorialStep {
	id: string;
	title: string;
	description: string;
	element?: string;
	position: 'top' | 'bottom' | 'left' | 'right' | 'center';
	action?: 'click' | 'drag' | 'upload' | 'wait';
	condition?: () => boolean;
	onComplete?: () => void;
	skipable?: boolean;
}

interface TutorialState {
	isActive: boolean;
	currentStepIndex: number;
	steps: TutorialStep[];
	completedSteps: Set<string>;
	isCompleted: boolean;
	hasStarted: boolean;
}

const TUTORIAL_STEPS: TutorialStep[] = [
	{
		id: 'welcome',
		title: 'Welcome to Image to Palette! 🎨',
		description: "Let's do a quick tour. You can skip any step.",
		position: 'center',
		skipable: true
	},
	{
		id: 'upload-image',
		title: 'Upload Your First Image',
		description: 'Click or drag an image into the upload area to get started.',
		element: 'button[aria-label="Upload an image or drag and drop it here"]',
		position: 'bottom',
		action: 'upload',
		condition: () => tutorialStore.state.hasImageUploaded,
		skipable: true
	},
	{
		id: 'canvas-interaction',
		title: 'Select Areas of Interest',
		description: 'Click and drag on the image to extract colors from a selected area.',
		element: 'canvas',
		position: 'top',
		action: 'drag',
		condition: () => tutorialStore.state.hasSelection,
		skipable: true
	},
	{
		id: 'selection-tools',
		title: 'Use Different Selection Tools',
		description: 'Use another selector from the toolbar and make one more selection.',
		element: '[role="toolbar"] button[aria-label="Selector 2"]',
		position: 'left',
		action: 'drag',
		condition: () => tutorialStore.state.hasSelectorClicked,
		skipable: true
	},
	{
		id: 'extract-colors',
		title: 'Copy Colors from Your Palette',
		description: 'Great. Click the highlighted color to copy its hex code.',
		element: '#tutorial-color-swatch',
		position: 'top',
		condition: () => tutorialStore.state.hasColorCopied,
		skipable: true
	},
	{
		id: 'save-palette',
		title: 'Save Your Current Palette',
		description: 'Click "Save Palette" to keep this palette for later use.',
		element: '#save-palette',
		position: 'top',
		action: 'click',
		condition: () => tutorialStore.state.hasCurrentPaletteSaved,
		skipable: true
	},
	{
		id: 'toolbar-features',
		title: 'Explore the Toolbar',
		description: 'Open Saved Palettes from the toolbar (🎨). You can drag this toolbar anywhere.',
		element: 'button[aria-label="Show saved palettes"]',
		position: 'left',
		action: 'click',
		condition: () => tutorialStore.state.hasSavedPalettesPopoverOpen,
		skipable: true
	},
	{
		id: 'apply-palette',
		title: 'Apply a Palette',
		description: 'Apply the highlighted tutorial palette to recolor your current image.',
		element: '.tutorial-palette-apply',
		position: 'left',
		condition: () => tutorialStore.state.hasSavedPaletteApplied,
		skipable: true
	},
	{
		id: 'toolbar-themes',
		title: 'Saved Themes Panel',
		description: 'This panel stores full editor themes you generate and save.',
		element: 'button[aria-label="Show saved themes"]',
		position: 'left',
		skipable: true
	},
	{
		id: 'toolbar-settings',
		title: 'Settings Panels',
		description: 'Use these panels for Wallhaven search and palette apply behavior.',
		element: 'button[aria-label="Configure Wallhaven search settings"]',
		position: 'left',
		skipable: true
	},
	{
		id: 'toolbar-copy-export',
		title: 'Copy and Export',
		description: 'Copy palette formats or download your current image from here.',
		element: 'button[aria-label="Copy Palette"]',
		position: 'left',
		skipable: true
	},
	{
		id: 'completion',
		title: 'Tutorial Complete! 🎉',
		description: "You're all set. Explore the app and build palettes from any image.",
		position: 'center',
		skipable: false
	}
];

function createTutorialStore() {
	const state = $state<TutorialState>({
		isActive: false,
		currentStepIndex: 0,
		steps: TUTORIAL_STEPS,
		completedSteps: new Set(),
		isCompleted: false,
		hasStarted: false
	});

	let hasImageUploaded = $state(false);
	let hasSelection = $state(false);
	let hasSavedPalette = $state(false);
	let hasSelectorClicked = $state(false);
	let hasColorCopied = $state(false);
	let hasCurrentPaletteSaved = $state(false);
	let hasSavedPalettesPopoverOpen = $state(false);
	let hasSavedPaletteApplied = $state(false);

	function resetToUploadState(store: typeof appStore) {
		store.state.image = null;
		store.state.imageLoaded = false;
		store.state.colors = [];
	}

	function clearNonGreenSelections(store: typeof appStore) {
		store.state.selectors.forEach((selector) => {
			if (selector.id !== UI.DEFAULT_SELECTOR_ID) {
				selector.selection = undefined;
			}
		});
	}

	return {
		get state() {
			return {
				...state,
				hasImageUploaded,
				hasSelection,
				hasSavedPalette,
				hasSelectorClicked,
				hasColorCopied,
				hasCurrentPaletteSaved,
				hasSavedPalettesPopoverOpen,
				hasSavedPaletteApplied
			};
		},

		start() {
			const hasCompletedBefore = localStorage.getItem('tutorial-completed') === 'true';
			const hasSkippedBefore = localStorage.getItem('tutorial-skipped') === 'true';
			if (hasCompletedBefore || hasSkippedBefore) {
				return;
			}

			state.isActive = true;
			state.hasStarted = true;
			state.currentStepIndex = 0;
			state.completedSteps.clear();

			hasSelectorClicked = false;
			hasColorCopied = false;
			hasCurrentPaletteSaved = false;
			hasSavedPalettesPopoverOpen = false;
			hasSavedPaletteApplied = false;
		},

		next() {
			if (state.currentStepIndex < state.steps.length - 1) {
				const currentStep = state.steps[state.currentStepIndex];
				state.completedSteps.add(currentStep.id);

				if (currentStep.onComplete) {
					currentStep.onComplete();
				}

				state.currentStepIndex++;
			} else {
				this.complete();
			}
		},

		previous() {
			if (state.currentStepIndex > 0) {
				state.currentStepIndex--;

				const targetStep = this.getCurrentStep();
				if (targetStep) {
					state.completedSteps.delete(targetStep.id);
				}
				if (targetStep) {
					switch (targetStep.id) {
						case 'upload-image':
							hasImageUploaded = false;
							resetToUploadState(appStore);
							break;

						case 'canvas-interaction':
							hasSelection = false;
							appStore.clearAllSelections();
							appStore.redrawCanvas();
							break;

						case 'selection-tools':
							hasSelectorClicked = false;
							clearNonGreenSelections(appStore);
							appStore.redrawCanvas();
							break;

						case 'extract-colors':
							hasColorCopied = false;
							break;

						case 'save-palette':
							hasCurrentPaletteSaved = false;
							break;

						case 'toolbar-features':
							hasSavedPalettesPopoverOpen = false;
							popoverStore.close();
							break;

						case 'apply-palette':
							hasSavedPaletteApplied = false;
							break;
					}
				}
			}
		},

		skip() {
			const currentStep = this.getCurrentStep();
			if (currentStep?.skipable) {
				this.next();
			}
		},

		complete() {
			state.isActive = false;
			state.isCompleted = true;
			localStorage.setItem('tutorial-completed', 'true');
		},

		exit() {
			state.isActive = false;
			localStorage.setItem('tutorial-skipped', 'true');
		},

		reset() {
			state.isActive = false;
			state.currentStepIndex = 0;
			state.completedSteps.clear();
			state.isCompleted = false;
			state.hasStarted = false;

			hasImageUploaded = false;
			hasSelection = false;
			hasSavedPalette = false;
			hasSelectorClicked = false;
			hasColorCopied = false;
			hasCurrentPaletteSaved = false;
			hasSavedPalettesPopoverOpen = false;
			hasSavedPaletteApplied = false;

			localStorage.removeItem('tutorial-completed');
			localStorage.removeItem('tutorial-skipped');
		},

		getCurrentStep(): TutorialStep | null {
			return state.steps[state.currentStepIndex] || null;
		},

		goToStep(stepId: string) {
			const stepIndex = state.steps.findIndex((step) => step.id === stepId);
			if (stepIndex !== -1) {
				state.currentStepIndex = stepIndex;
			}
		},

		checkStepCondition(): boolean {
			const currentStep = this.getCurrentStep();
			if (!currentStep || !currentStep.condition) {
				return true;
			}
			return currentStep.condition();
		},

		setImageUploaded(uploaded: boolean) {
			hasImageUploaded = uploaded;
		},

		setHasSelection(selection: boolean) {
			hasSelection = selection;
		},

		setSelectorClicked(clicked: boolean) {
			const currentStep = this.getCurrentStep();
			if (currentStep?.id === 'selection-tools') {
				hasSelectorClicked = clicked;
			}
		},

		setColorCopied(copied: boolean) {
			const currentStep = this.getCurrentStep();
			if (currentStep?.id === 'extract-colors') {
				hasColorCopied = copied;
			}
		},

		setHasSavedPalette(saved: boolean) {
			hasSavedPalette = saved;
		},

		setCurrentPaletteSaved(saved: boolean) {
			const currentStep = this.getCurrentStep();
			if (currentStep?.id === 'save-palette') {
				hasCurrentPaletteSaved = saved;
			}
		},

		setSavedPalettesPopoverOpen(open: boolean) {
			const currentStep = this.getCurrentStep();
			if (currentStep?.id === 'toolbar-features') {
				hasSavedPalettesPopoverOpen = open;
			}
		},

		setSavedPaletteApplied(applied: boolean) {
			const currentStep = this.getCurrentStep();
			if (currentStep?.id === 'apply-palette') {
				hasSavedPaletteApplied = applied;
			}
		},

		getTutorialApplyPalette(): Color[] {
			return TUTORIAL_APPLY_PALETTE;
		},

		shouldShowTutorial(): boolean {
			const hasCompletedBefore = localStorage.getItem('tutorial-completed') === 'true';
			const hasSkippedBefore = localStorage.getItem('tutorial-skipped') === 'true';
			return !hasCompletedBefore && !hasSkippedBefore;
		}
	};
}

export const tutorialStore = createTutorialStore();
