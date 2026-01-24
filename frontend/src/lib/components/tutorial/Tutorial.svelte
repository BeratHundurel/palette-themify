<script lang="ts">
	import { UI } from '$lib/constants';
	import { tutorialStore, type TutorialStep } from '$lib/stores/tutorial.svelte';
	import { appStore } from '$lib/stores/app.svelte';
	import { popoverStore } from '$lib/stores/popovers.svelte';
	import { tick } from 'svelte';
	import { fade, fly, scale } from 'svelte/transition';
	import { cn } from '$lib/utils';

	let highlightElement: HTMLElement | null = $state(null);
	let tooltipElement: HTMLElement | null = $state(null);

	$effect(() => {
		const currentStep = tutorialStore.getCurrentStep();
		if (currentStep?.id === 'upload-image' && appStore.state.imageLoaded && appStore.state.canvas) {
			tutorialStore.setImageUploaded(true);
		}

		if (
			currentStep?.id === 'canvas-interaction' &&
			appStore.state.selectors.some((s) => s.selection) &&
			!appStore.state.isDragging &&
			!appStore.state.isExtracting
		) {
			tutorialStore.setHasSelection(true);
		}

		if (
			currentStep?.id === 'selection-tools' &&
			appStore.state.activeSelectorId &&
			appStore.state.activeSelectorId !== UI.DEFAULT_SELECTOR_ID
		) {
			const nonGreenSelector = appStore.state.selectors.find(
				(s) => s.id === appStore.state.activeSelectorId && s.id !== UI.DEFAULT_SELECTOR_ID
			);

			if (nonGreenSelector?.selection && !appStore.state.isDragging && !appStore.state.isExtracting) {
				tutorialStore.setSelectorClicked(true);
			}
		}

		if (currentStep?.id === 'toolbar-features' && popoverStore.isOpen('saved')) {
			tutorialStore.setSavedPalettesPopoverOpen(true);
			const popover = document.querySelector('.palette-dropdown-base') as HTMLElement;
			if (popover) {
				popover.focus();
				popover.setAttribute('tabindex', '-1');
			}
		}
	});

	$effect(() => {
		const currentStep = tutorialStore.getCurrentStep();
		if (currentStep && currentStep.condition && tutorialStore.checkStepCondition()) {
			tutorialStore.next();
		}
	});

	$effect(() => {
		if (tutorialStore.state.isActive) {
			updateHighlight();
		}
	});

	async function updateHighlight() {
		await tick();
		const currentStep = tutorialStore.getCurrentStep();

		if (currentStep?.element) {
			const element = document.querySelector(currentStep.element) as HTMLElement;
			if (element) {
				highlightElement = element;
				return;
			}
		}

		highlightElement = null;
	}

	function getTooltipPosition(step: TutorialStep) {
		if (step.position === 'center') {
			return {
				styles: {
					top: '50%',
					left: '50%',
					transform: 'translate(-50%, -50%)'
				},
				actualPosition: 'center'
			};
		}

		if (!highlightElement)
			return {
				styles: {
					top: '50%',
					left: '50%',
					transform: 'translate(-50%, -50%)'
				},
				actualPosition: 'center'
			};

		const rect = highlightElement.getBoundingClientRect();
		const tooltipRect = tooltipElement?.getBoundingClientRect();
		const tooltipWidth = tooltipRect?.width || 320;
		const tooltipHeight = tooltipRect?.height || 200;
		const viewportWidth = window.innerWidth;
		const viewportHeight = window.innerHeight;

		let position = step.position;

		if (position === 'top' && rect.top - tooltipHeight - 20 < 0) {
			position = rect.left > viewportWidth / 2 ? 'left' : 'right';
		}
		if (position === 'bottom' && rect.bottom + tooltipHeight + 20 > viewportHeight) {
			position = rect.left > viewportWidth / 2 ? 'left' : 'right';
		}
		if (position === 'left' && rect.left - tooltipWidth - 20 < 0) {
			position = 'right';
		}
		if (position === 'right' && rect.right + tooltipWidth + 20 > viewportWidth) {
			position = 'left';
		}

		let styles;
		switch (position) {
			case 'top':
				styles = {
					top: `${Math.max(20, rect.top - tooltipHeight - 20)}px`,
					left: `${Math.max(20, Math.min(viewportWidth - tooltipWidth - 20, rect.left + rect.width / 2 - tooltipWidth / 2))}px`
				};
				break;
			case 'bottom':
				styles = {
					top: `${Math.min(viewportHeight - tooltipHeight - 20, rect.bottom + 20)}px`,
					left: `${Math.max(20, Math.min(viewportWidth - tooltipWidth - 20, rect.left + rect.width / 2 - tooltipWidth / 2))}px`
				};
				break;
			case 'left':
				styles = {
					top: `${Math.max(20, Math.min(viewportHeight - tooltipHeight - 20, rect.top + rect.height / 2 - tooltipHeight / 2))}px`,
					left: `${Math.max(20, rect.left - tooltipWidth - 20)}px`
				};
				break;
			case 'right':
				styles = {
					top: `${Math.max(20, Math.min(viewportHeight - tooltipHeight - 20, rect.top + rect.height / 2 - tooltipHeight / 2))}px`,
					left: `${Math.min(viewportWidth - tooltipWidth - 20, rect.right + 20)}px`
				};
				break;
			default:
				styles = {};
				break;
		}

		return { styles, actualPosition: position };
	}

	let currentStep = $derived(tutorialStore.getCurrentStep());
	let isFirstStep = $derived(tutorialStore.state.currentStepIndex === 0);
	let isLastStep = $derived(tutorialStore.state.currentStepIndex === tutorialStore.state.steps.length - 1);
	let tooltipPositionResult = $derived(
		currentStep ? getTooltipPosition(currentStep) : { styles: {}, actualPosition: 'center' }
	);
	let tooltipStyles = $derived(tooltipPositionResult.styles);
	let actualPosition = $derived(tooltipPositionResult.actualPosition);
</script>

{#if tutorialStore.state.isActive}
	<div class="pointer-events-none fixed inset-0 z-30" transition:fade={{ duration: 300 }}>
		{#if highlightElement && currentStep?.element}
			{@const rect = highlightElement.getBoundingClientRect()}
			<div
				class="tutorial-highlight border-brand pointer-events-none absolute z-[10001] rounded-md border-[3px]"
				style={`
					top: ${rect.top - 8}px;
					left: ${rect.left - 8}px;
					width: ${rect.width + 16}px;
					height: ${rect.height + 16}px;
				`}
				transition:scale={{ duration: 300 }}
			></div>
		{/if}

		{#if currentStep}
			<div
				bind:this={tooltipElement}
				class={cn(
					'pointer-events-auto absolute z-[10002] max-w-[360px] min-w-[300px]',
					'max-md:!right-[5vw] max-md:!left-[5vw] max-md:max-w-[90vw] max-md:min-w-[280px]'
				)}
				style={Object.entries(tooltipStyles)
					.map(([key, value]) => `${key}: ${value}`)
					.join('; ')}
				transition:fly={{ y: 20, duration: 300 }}
			>
				<div class="border-brand/50 rounded-xl border bg-zinc-900 p-6 max-md:p-5">
					<div class="mb-4 flex items-center justify-between border-b border-zinc-600 pb-4">
						<div class="flex gap-2">
							{#each tutorialStore.state.steps as step, index (step.id)}
								<div
									class={cn(
										'h-2 w-2 rounded-full transition-[background-color,box-shadow,transform] duration-300',
										tutorialStore.state.completedSteps.has(step.id)
											? 'bg-emerald-500'
											: index === tutorialStore.state.currentStepIndex
												? 'bg-brand scale-[1.2] shadow-[0_0_8px_rgba(238,179,143,0.4)]'
												: 'bg-zinc-500'
									)}
								></div>
							{/each}
						</div>
						<span class="font-mono text-xs text-zinc-400">
							{tutorialStore.state.currentStepIndex + 1} of {tutorialStore.state.steps.length}
						</span>
					</div>

					<div class="mb-4">
						<h3 class="text-brand mb-2 text-lg font-semibold">{currentStep.title}</h3>
						<p class="mb-0 text-sm leading-relaxed text-zinc-300">{currentStep.description}</p>

						{#if currentStep.action === 'upload'}
							<div class="bg-brand/10 border-brand/20 mt-2 rounded-md border p-2 px-3 text-xs text-zinc-400">
								💡 Try uploading a photo with multiple colors for the best results!
							</div>
						{:else if currentStep.action === 'drag'}
							<div class="bg-brand/10 border-brand/20 mt-2 rounded-md border p-2 px-3 text-xs text-zinc-400">
								💡 Click and drag to create a selection rectangle on the image
							</div>
						{:else if currentStep.action === 'click'}
							<div class="bg-brand/10 border-brand/20 mt-2 rounded-md border p-2 px-3 text-xs text-zinc-400">
								💡 Click the highlighted element to continue
							</div>
						{/if}
					</div>

					<div class="flex flex-col gap-4">
						<div class="flex justify-between gap-3 max-md:flex-col">
							{#if !isFirstStep}
								<button
									class={cn(
										'hover:border-brand/50 cursor-pointer rounded-md border border-zinc-600 px-4 py-2 text-sm font-medium text-zinc-300 outline-0 transition-[background-color,border-color] duration-300 hover:bg-zinc-800'
									)}
									onclick={() => tutorialStore.previous()}
								>
									← Previous
								</button>
							{/if}

							{#if currentStep.condition && !tutorialStore.checkStepCondition()}
								<button
									class={cn(
										'animate-pulse cursor-not-allowed rounded-md border-0 bg-zinc-800 px-4 py-2 text-sm font-medium text-zinc-400 opacity-50 outline-0'
									)}
									disabled
								>
									Waiting for action...
								</button>
							{:else if !isLastStep}
								<button
									class={cn(
										'bg-brand hover:bg-brand-hover hover:shadow-brand-lg cursor-pointer rounded-md border-0 px-4 py-2 text-sm font-medium text-zinc-900 outline-0 transition-[background-color,box-shadow] duration-300 '
									)}
									onclick={() => tutorialStore.next()}
								>
									Next →
								</button>
							{:else}
								<button
									class={cn(
										'bg-brand hover:bg-brand-hover hover:shadow-brand-lg cursor-pointer rounded-md border-0 px-4 py-2 text-sm font-medium text-zinc-900 outline-0 transition-[background-color,box-shadow] duration-300 '
									)}
									onclick={() => tutorialStore.next()}
								>
									Complete 🎉
								</button>
							{/if}
						</div>

						<div class="flex justify-between gap-3">
							{#if currentStep.skipable}
								<button
									class={cn(
										'cursor-pointer rounded-md border-0 bg-transparent px-4 py-2 text-xs font-medium text-zinc-400 outline-0 transition-[color] duration-300 hover:text-zinc-300'
									)}
									onclick={() => tutorialStore.skip()}
								>
									Skip
								</button>
							{/if}

							<button
								class={cn(
									'cursor-pointer rounded-md border-0 bg-transparent px-4 py-2 text-xs font-medium text-zinc-400 outline-0 transition-[color] duration-300 hover:text-zinc-300'
								)}
								onclick={() => tutorialStore.exit()}
							>
								Exit Tutorial
							</button>
						</div>
					</div>
				</div>

				{#if currentStep.element && actualPosition !== 'center'}
					<div
						class="tutorial-arrow absolute h-0 w-0 border-[8px] border-transparent tutorial-arrow-{actualPosition}"
					></div>
				{/if}
			</div>
		{/if}
	</div>
{/if}

<style>
	.tutorial-highlight {
		box-shadow:
			0 0 0 9999px rgba(0, 0, 0, 0.7),
			0 0 20px rgba(238, 179, 143, 0.5),
			inset 0 0 20px rgba(238, 179, 143, 0.2);
	}

	.tutorial-arrow {
		position: absolute;
		width: 0;
		height: 0;
		border: 8px solid transparent;
	}

	.tutorial-arrow-top {
		bottom: -16px;
		left: 50%;
		transform: translateX(-50%);
		border-top-color: #27272a;
	}

	.tutorial-arrow-bottom {
		top: -16px;
		left: 50%;
		transform: translateX(-50%);
		border-bottom-color: #27272a;
	}

	.tutorial-arrow-left {
		right: -16px;
		top: 50%;
		transform: translateY(-50%);
		border-left-color: #27272a;
	}

	.tutorial-arrow-right {
		left: -16px;
		top: 50%;
		transform: translateY(-50%);
		border-right-color: #27272a;
	}
</style>
