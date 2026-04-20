import { appStore } from './app/store.svelte';
import { popoverStore } from './popovers.svelte';

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
		title: 'Welcome to ThemeSmith! 🎨',
		description: "Let's do a quick tour of palette extraction and theme export. You can skip any step.",
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
		id: 'open-theme-inspector',
		title: 'Open Theme Inspector',
		description: 'Open Theme Inspector to generate a full editor theme from your extracted palette.',
		element: '#open-theme-inspector',
		position: 'top',
		action: 'click',
		condition: () => tutorialStore.state.hasThemeInspectorOpened,
		skipable: true
	},
	{
		id: 'theme-appearance',
		title: 'Choose Theme Appearance',
		description: 'Switch between Dark and Light to see how the same palette adapts to different backgrounds.',
		element: '#tutorial-theme-appearance-options',
		position: 'top',
		action: 'click',
		condition: () => tutorialStore.state.hasThemeAppearanceChanged,
		skipable: true
	},
	{
		id: 'theme-editor',
		title: 'Target VS Code or Zed',
		description: 'Toggle editor type to generate the correct theme schema for your editor.',
		element: '#tutorial-theme-editor-options',
		position: 'top',
		action: 'click',
		condition: () => tutorialStore.state.hasThemeEditorChanged,
		skipable: true
	},
	{
		id: 'theme-overrides',
		title: 'Customize Color Roles',
		description: 'Edit at least one override to control background, foreground, or syntax accent assignments.',
		element: '#tutorial-theme-overrides-root',
		position: 'top',
		action: 'click',
		condition: () => tutorialStore.state.hasThemeOverridesEdited,
		skipable: true
	},
	{
		id: 'theme-copy-json',
		title: 'Export Theme JSON',
		description: 'Copy Theme JSON to export your generated theme. Keep Save on copy enabled to store it locally too.',
		element: '#tutorial-theme-copy-json',
		position: 'top',
		action: 'click',
		condition: () => tutorialStore.state.hasThemeJsonCopied,
		skipable: true
	},
	{
		id: 'toolbar-themes',
		title: 'Saved Themes Panel',
		description: 'Open Saved Themes (🧩) to load, save, or share themes you exported.',
		element: '#tutorial-open-saved-themes',
		position: 'left',
		action: 'click',
		condition: () => tutorialStore.state.hasSavedThemesPopoverOpen,
		skipable: true
	},
	{
		id: 'completion',
		title: 'Tutorial Complete! 🎉',
		description: "You're all set. Build palettes, generate themes, and export them for your editor.",
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
	let hasThemeInspectorOpened = $state(false);
	let hasThemeAppearanceChanged = $state(false);
	let hasThemeEditorChanged = $state(false);
	let hasThemeOverridesEdited = $state(false);
	let hasThemeJsonCopied = $state(false);
	let hasSavedThemesPopoverOpen = $state(false);

	function resetToUploadState(store: typeof appStore) {
		store.state.image = null;
		store.state.imageLoaded = false;
		store.state.colors = [];
	}

	return {
		get state() {
			return {
				...state,
				hasImageUploaded,
				hasSelection,
				hasThemeInspectorOpened,
				hasThemeAppearanceChanged,
				hasThemeEditorChanged,
				hasThemeOverridesEdited,
				hasThemeJsonCopied,
				hasSavedThemesPopoverOpen
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

			hasImageUploaded = false;
			hasSelection = false;
			hasThemeInspectorOpened = false;
			hasThemeAppearanceChanged = false;
			hasThemeEditorChanged = false;
			hasThemeOverridesEdited = false;
			hasThemeJsonCopied = false;
			hasSavedThemesPopoverOpen = false;
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

						case 'open-theme-inspector':
							hasThemeInspectorOpened = false;
							break;

						case 'theme-appearance':
							hasThemeAppearanceChanged = false;
							break;

						case 'theme-editor':
							hasThemeEditorChanged = false;
							break;

						case 'theme-overrides':
							hasThemeOverridesEdited = false;
							break;

						case 'theme-copy-json':
							hasThemeJsonCopied = false;
							break;

						case 'toolbar-themes':
							hasSavedThemesPopoverOpen = false;
							popoverStore.close('themes');
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
			hasThemeInspectorOpened = false;
			hasThemeAppearanceChanged = false;
			hasThemeEditorChanged = false;
			hasThemeOverridesEdited = false;
			hasThemeJsonCopied = false;
			hasSavedThemesPopoverOpen = false;

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

		setThemeInspectorOpened(opened: boolean) {
			const currentStep = this.getCurrentStep();
			if (currentStep?.id === 'open-theme-inspector') {
				hasThemeInspectorOpened = opened;
			}
		},

		setThemeAppearanceChanged(changed: boolean) {
			const currentStep = this.getCurrentStep();
			if (currentStep?.id === 'theme-appearance') {
				hasThemeAppearanceChanged = changed;
			}
		},

		setThemeEditorChanged(changed: boolean) {
			const currentStep = this.getCurrentStep();
			if (currentStep?.id === 'theme-editor') {
				hasThemeEditorChanged = changed;
			}
		},

		setThemeOverridesEdited(edited: boolean) {
			const currentStep = this.getCurrentStep();
			if (currentStep?.id === 'theme-overrides') {
				hasThemeOverridesEdited = edited;
			}
		},

		setThemeJsonCopied(copied: boolean) {
			const currentStep = this.getCurrentStep();
			if (currentStep?.id === 'theme-copy-json') {
				hasThemeJsonCopied = copied;
			}
		},

		setSavedThemesPopoverOpen(open: boolean) {
			const currentStep = this.getCurrentStep();
			if (currentStep?.id === 'toolbar-themes') {
				hasSavedThemesPopoverOpen = open;
			}
		},

		shouldShowTutorial(): boolean {
			const hasCompletedBefore = localStorage.getItem('tutorial-completed') === 'true';
			const hasSkippedBefore = localStorage.getItem('tutorial-skipped') === 'true';
			return !hasCompletedBefore && !hasSkippedBefore;
		}
	};
}

export const tutorialStore = createTutorialStore();
