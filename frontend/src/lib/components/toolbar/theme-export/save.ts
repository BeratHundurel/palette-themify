import type { EditorThemeType } from '$lib/api/theme';
import { appStore } from '$lib/stores/app.svelte';
import { popoverStore } from '$lib/stores/popovers.svelte';
import type { SavedThemeItem, ThemeColorWithUsage, ThemeResponse } from '$lib/types/theme';
import toast from 'svelte-french-toast';

import { getThemeSignature, normalizeThemeName, validateThemeName } from './utils';

type SaveThemeArgs = {
	name: string;
	editorType: EditorThemeType;
	themeResult: ThemeResponse | null;
	themeColorsWithUsage: ThemeColorWithUsage[];
};

type ExportThemeArgs = SaveThemeArgs & {
	saveOnCopy: boolean;
};

function buildSavedTheme({
	id,
	name,
	editorType,
	themeResult,
	themeColorsWithUsage
}: {
	id?: string;
	name: string;
	editorType: EditorThemeType;
	themeResult: ThemeResponse;
	themeColorsWithUsage: ThemeColorWithUsage[];
}): SavedThemeItem {
	return {
		id: id ?? `local_${Date.now()}`,
		name,
		editorType,
		themeResult,
		themeColorsWithUsage,
		createdAt: new Date().toISOString(),
		signature: getThemeSignature(themeResult)
	};
}

export function saveTheme({ name, editorType, themeResult, themeColorsWithUsage }: SaveThemeArgs) {
	if (!themeResult || !name) return;

	const trimmedName = name.trim();
	if (!trimmedName) return;

	const normalizedName = normalizeThemeName(trimmedName);
	const signature = getThemeSignature(themeResult);
	const existingThemes = appStore.state.savedThemes;

	const identicalTheme = existingThemes.find((item) => {
		const existingSignature = item.signature ?? getThemeSignature(item.themeResult);
		return item.editorType === editorType && existingSignature === signature;
	});
	if (identicalTheme) {
		return;
	}

	const nameMatch = existingThemes.find((item) => normalizeThemeName(item.name) === normalizedName);
	if (nameMatch) {
		const replace = confirm(`A theme named "${trimmedName}" already exists. Replace it?`);
		if (replace) {
			const saved = buildSavedTheme({
				id: nameMatch.id,
				name: trimmedName,
				editorType,
				themeResult,
				themeColorsWithUsage
			});
			appStore.replaceSavedTheme(nameMatch.id, saved);
			return;
		}

		const renamed = prompt('Enter a new name for the theme:', trimmedName);
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
			themeResult,
			themeColorsWithUsage
		});
		appStore.saveThemeToLocal(saved);
		return;
	}

	const saved = buildSavedTheme({
		name: trimmedName,
		editorType,
		themeResult,
		themeColorsWithUsage
	});
	appStore.saveThemeToLocal(saved);
}

export async function exportTheme({
	name,
	editorType,
	themeResult,
	themeColorsWithUsage,
	saveOnCopy
}: ExportThemeArgs) {
	if (!themeResult) return;

	try {
		const themeJson = JSON.stringify(themeResult.theme, null, 2);
		await navigator.clipboard.writeText(themeJson);

		if (saveOnCopy) {
			saveTheme({
				name: name.trim(),
				editorType,
				themeResult,
				themeColorsWithUsage
			});
		}

		toast.success('Theme JSON copied to clipboard!');
		popoverStore.close('themeExport');
	} catch {
		toast.error('Could not copy the theme. Please try again.');
	}
}
