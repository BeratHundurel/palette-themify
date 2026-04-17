export type ConfirmDialogVariant = 'default' | 'danger';

export type ConfirmDialogOptions = {
	title?: string;
	message: string;
	confirmLabel?: string;
	cancelLabel?: string;
	variant?: ConfirmDialogVariant;
};

export type PromptDialogOptions = {
	title?: string;
	message: string;
	placeholder?: string;
	initialValue?: string;
	confirmLabel?: string;
	cancelLabel?: string;
	variant?: ConfirmDialogVariant;
};

type ConfirmDialogRequest = {
	id: number;
	title: string;
	message: string;
	confirmLabel: string;
	cancelLabel: string;
	variant: ConfirmDialogVariant;
	resolve: (value: boolean) => void;
};

type DialogState = {
	confirm: ConfirmDialogRequest | null;
	prompt: PromptDialogRequest | null;
};

type PromptDialogRequest = {
	id: number;
	title: string;
	message: string;
	placeholder: string;
	initialValue: string;
	confirmLabel: string;
	cancelLabel: string;
	variant: ConfirmDialogVariant;
	resolve: (value: string | null) => void;
};

function createDialogStore() {
	const state = $state<DialogState>({
		confirm: null,
		prompt: null
	});

	let nextId = 0;

	return {
		get state() {
			return state;
		},

		confirm(options: ConfirmDialogOptions): Promise<boolean> {
			if (state.confirm) {
				state.confirm.resolve(false);
			}
			if (state.prompt) {
				state.prompt.resolve(null);
				state.prompt = null;
			}

			return new Promise<boolean>((resolve) => {
				state.confirm = {
					id: ++nextId,
					title: options.title ?? 'Confirm action',
					message: options.message,
					confirmLabel: options.confirmLabel ?? 'Confirm',
					cancelLabel: options.cancelLabel ?? 'Cancel',
					variant: options.variant ?? 'default',
					resolve
				};
			});
		},

		prompt(options: PromptDialogOptions): Promise<string | null> {
			if (state.prompt) {
				state.prompt.resolve(null);
			}
			if (state.confirm) {
				state.confirm.resolve(false);
				state.confirm = null;
			}

			return new Promise<string | null>((resolve) => {
				state.prompt = {
					id: ++nextId,
					title: options.title ?? 'Enter value',
					message: options.message,
					placeholder: options.placeholder ?? '',
					initialValue: options.initialValue ?? '',
					confirmLabel: options.confirmLabel ?? 'Save',
					cancelLabel: options.cancelLabel ?? 'Cancel',
					variant: options.variant ?? 'default',
					resolve
				};
			});
		},

		resolveConfirm(value: boolean) {
			if (!state.confirm) return;
			const current = state.confirm;
			state.confirm = null;
			current.resolve(value);
		},

		resolvePrompt(value: string | null) {
			if (!state.prompt) return;
			const current = state.prompt;
			state.prompt = null;
			current.resolve(value);
		},

		close() {
			this.resolveConfirm(false);
			this.resolvePrompt(null);
		}
	};
}

export const dialogStore = createDialogStore();
