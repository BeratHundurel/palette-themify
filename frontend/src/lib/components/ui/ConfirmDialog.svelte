<script lang="ts">
	import { dialogStore } from '$lib/stores/dialog.svelte';

	let confirm = $derived(dialogStore.state.confirm);

	function closeWith(value: boolean) {
		dialogStore.resolveConfirm(value);
	}

	function onBackdropKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			event.preventDefault();
			closeWith(false);
		}
	}
</script>

{#if confirm}
	<div
		class="fixed inset-0 z-[120] flex items-center justify-center p-4"
		role="presentation"
		onkeydown={onBackdropKeydown}
	>
		<button
			type="button"
			class="animate-fade-in absolute inset-0 cursor-default border-0 bg-black/75 p-0"
			onclick={() => closeWith(false)}
			aria-label="Close dialog"
		></button>

		<div
			role="dialog"
			aria-modal="true"
			aria-labelledby="confirm-dialog-title"
			class="animate-scale-in border-brand/60 relative z-10 w-full max-w-md rounded-xl border bg-zinc-900 p-6 shadow-2xl"
		>
			<div class="mb-5">
				<h2 id="confirm-dialog-title" class="text-lg font-semibold text-zinc-100">{confirm.title}</h2>
				<p class="mt-2 text-sm leading-6 text-zinc-300">{confirm.message}</p>
			</div>

			<div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
				<button
					type="button"
					onclick={() => closeWith(false)}
					class="hover:border-brand/50 rounded-lg border border-zinc-600 px-4 py-2 text-sm font-medium text-zinc-200 transition-[background-color,border-color] duration-200 hover:bg-zinc-800/70"
				>
					{confirm.cancelLabel}
				</button>

				<button
					type="button"
					onclick={() => closeWith(true)}
					class={confirm.variant === 'danger'
						? 'rounded-lg bg-red-500 px-4 py-2 text-sm font-semibold text-zinc-950 transition-[background-color,box-shadow] duration-200 hover:bg-red-400 hover:shadow-[0_0_18px_rgba(248,113,113,0.45)]'
						: 'bg-brand shadow-brand/20 hover:shadow-brand/40 hover:bg-brand-hover rounded-lg px-4 py-2 text-sm font-semibold text-zinc-900 transition-[background-color,box-shadow] duration-200'}
				>
					{confirm.confirmLabel}
				</button>
			</div>
		</div>
	</div>
{/if}
