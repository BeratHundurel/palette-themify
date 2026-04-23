import { getDesktopSaveErrorMessage, isDesktopApp, saveThemeToEditorTarget } from '$lib/platform';
import { appStore } from '$lib/stores/app/store.svelte';
import { dialogStore } from '$lib/stores/dialog.svelte';
import { popoverStore } from '$lib/stores/popovers.svelte';
import type { EditorThemeType } from '$lib/types/themeApi';
import type { SavedThemeItem, ThemeGenerationResponse } from '$lib/types/theme';
import toast from 'svelte-french-toast';

import { normalizeThemeName, validateThemeName } from './utils';

type SaveThemeArgs = {
	name: string;
	editorType: EditorThemeType;
	themeResult: ThemeGenerationResponse | null;
};

type ExportThemeArgs = SaveThemeArgs & {
	saveOnCopy: boolean;
	onExported?: () => void;
};

function buildSavedTheme({
	id,
	name,
	editorType,
	themeResult
}: {
	id?: string;
	name: string;
	editorType: EditorThemeType;
	themeResult: ThemeGenerationResponse;
}): SavedThemeItem {
	return {
		id: id ?? `local_${Date.now()}`,
		name,
		editorType,
		themeResult,
		createdAt: new Date().toISOString(),
		signature: appStore.getThemeSignature(themeResult)
	};
}

export async function saveTheme({ name, editorType, themeResult }: SaveThemeArgs) {
	if (!themeResult || !name) return;

	const trimmedName = name.trim();
	if (!trimmedName) return;

	const normalizedName = normalizeThemeName(trimmedName);
	const signature = appStore.getThemeSignature(themeResult);
	const existingThemes = appStore.state.savedThemes;

	const identicalTheme = existingThemes.find((item) => {
		const existingSignature = item.signature ?? appStore.getThemeSignature(item.themeResult);
		return item.editorType === editorType && existingSignature === signature;
	});
	if (identicalTheme) {
		return;
	}

	const nameMatch = existingThemes.find((item) => normalizeThemeName(item.name) === normalizedName);
	if (nameMatch) {
		const replace = await dialogStore.confirm({
			title: 'Theme name already exists',
			message: `A theme named "${trimmedName}" already exists. Replace it?`,
			confirmLabel: 'Replace theme',
			variant: 'danger'
		});
		if (replace) {
			const saved = buildSavedTheme({
				id: nameMatch.id,
				name: trimmedName,
				editorType,
				themeResult
			});
			appStore.replaceSavedTheme(nameMatch.id, saved);
			return;
		}

		const renamed = await dialogStore.prompt({
			title: 'Rename theme',
			message: 'Enter a new name for the theme.',
			initialValue: trimmedName,
			confirmLabel: 'Save as new theme'
		});
		if (renamed === null) return;

		const renameError = validateThemeName(renamed);
		if (renameError) {
			toast.error(renameError);
			return;
		}

		const normalizedRename = normalizeThemeName(renamed);
		if (existingThemes.find((item) => normalizeThemeName(item.name) === normalizedRename)) {
			toast.error('A theme with that name already exists.');
			return;
		}

		const saved = buildSavedTheme({
			name: renamed.trim(),
			editorType,
			themeResult
		});
		appStore.saveThemeToLocal(saved);
		return;
	}

	const saved = buildSavedTheme({
		name: trimmedName,
		editorType,
		themeResult
	});
	appStore.saveThemeToLocal(saved);
}

export async function exportTheme({ name, editorType, themeResult, saveOnCopy, onExported }: ExportThemeArgs) {
	if (!themeResult) return;

	try {
		const themeJson = JSON.stringify(themeResult.theme, null, 2);
		await navigator.clipboard.writeText(themeJson);

		if (saveOnCopy) {
			await saveTheme({
				name: name.trim(),
				editorType,
				themeResult
			});
		}

		toast.success('Theme JSON copied to clipboard!');
		onExported?.();
		popoverStore.close('themeExport');
	} catch {
		toast.error('Could not copy the theme. Please try again.');
	}
}

export async function exportThemeToEditorFolder({ name, editorType, themeResult, saveOnCopy }: ExportThemeArgs) {
	if (!themeResult) return;
	if (!isDesktopApp) {
		toast.error('Save to editor folder is only available in desktop app.');
		return;
	}

	const trimmedName = name.trim();
	if (!trimmedName) {
		toast.error('Theme name cannot be empty.');
		return;
	}

	try {
		const themeJson = JSON.stringify(themeResult.theme, null, 2);
		await saveThemeToEditorTarget({
			editorType,
			themeName: trimmedName,
			themeJSON: themeJson
		});

		if (saveOnCopy) {
			await saveTheme({
				name: trimmedName,
				editorType,
				themeResult
			});
		}

		toast.success('Theme saved to editor folder');
		popoverStore.close('themeExport');
	} catch (error) {
		console.error('Error saving theme to editor folder:', error);
		toast.error(getDesktopSaveErrorMessage(error));
	}
}
