<script lang="ts">
	import { dialogStore } from '$lib/stores/dialog.svelte';
	import { onMount } from 'svelte';

	let prompt = $derived(dialogStore.state.prompt);
	let inputValue = $state('');
	let inputEl = $state<HTMLInputElement | null>(null);

	$effect(() => {
		if (!prompt) return;
		inputValue = prompt.initialValue;
	});

	onMount(() => {
		if (!prompt || !inputEl) return;
		inputEl.focus();
		inputEl.select();
	});

	$effect(() => {
		if (prompt && inputEl) {
			inputEl.focus();
			inputEl.select();
		}
	});

	function onCancel() {
		dialogStore.resolvePrompt(null);
	}

	function onConfirm() {
		dialogStore.resolvePrompt(inputValue);
	}

	function onBackdropKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			event.preventDefault();
			onCancel();
		}
	}
</script>

{#if prompt}
	<div
		class="fixed inset-0 z-[121] flex items-center justify-center p-4"
		role="presentation"
		onkeydown={onBackdropKeydown}
	>
		<button
			type="button"
			class="animate-fade-in absolute inset-0 cursor-default border-0 bg-black/75 p-0"
			onclick={onCancel}
			aria-label="Close dialog"
		></button>

		<div
			class="animate-scale-in border-brand/60 relative z-10 w-full max-w-md rounded-xl border bg-zinc-900 p-6 shadow-2xl"
			role="dialog"
			aria-modal="true"
			aria-labelledby="prompt-dialog-title"
		>
			<form
				onsubmit={(event) => {
					event.preventDefault();
					onConfirm();
				}}
			>
				<div class="mb-5">
					<h2 id="prompt-dialog-title" class="text-lg font-semibold text-zinc-100">{prompt.title}</h2>
					<p class="mt-2 text-sm leading-6 text-zinc-300">{prompt.message}</p>
				</div>

				<input
					type="text"
					bind:this={inputEl}
					bind:value={inputValue}
					placeholder={prompt.placeholder}
					class="focus:border-brand/50 mb-5 w-full rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-200 placeholder-zinc-500 transition-[border-color,box-shadow] duration-200 focus:outline-none"
				/>

				<div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
					<button
						type="button"
						onclick={onCancel}
						class="hover:border-brand/50 rounded-lg border border-zinc-600 px-4 py-2 text-sm font-medium text-zinc-200 transition-[background-color,border-color] duration-200 hover:bg-zinc-800/70"
					>
						{prompt.cancelLabel}
					</button>

					<button
						type="submit"
						class={prompt.variant === 'danger'
							? 'rounded-lg bg-red-500 px-4 py-2 text-sm font-semibold text-zinc-950 transition-[background-color,box-shadow] duration-200 hover:bg-red-400 hover:shadow-[0_0_18px_rgba(248,113,113,0.45)]'
							: 'bg-brand shadow-brand/20 hover:shadow-brand/40 hover:bg-brand-hover rounded-lg px-4 py-2 text-sm font-semibold text-zinc-900 transition-[background-color,box-shadow] duration-200'}
					>
						{prompt.confirmLabel}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
