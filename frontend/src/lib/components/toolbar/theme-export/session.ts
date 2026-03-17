import type { EditorThemeType, ThemeAppearance } from '$lib/api/theme';
import { extractThemeColorsWithUsage } from '$lib/colorUtils';
import { appStore } from '$lib/stores/app.svelte';
import type { ThemeGenerationResponse, ThemeOverrides } from '$lib/types/theme';

import { cloneThemeResponse, getThemeVersionKey } from './utils';

export function clearThemeVersions() {
	appStore.state.themeExport.themeVersions = {};
}

export function cacheThemeVersion(
	response: ThemeGenerationResponse,
	type: EditorThemeType,
	appearance: ThemeAppearance
) {
	appStore.state.themeExport.themeVersions = {
		...appStore.state.themeExport.themeVersions,
		[getThemeVersionKey(type, appearance)]: cloneThemeResponse(response)
	};
}

export function setActiveThemeResponse(
	response: ThemeGenerationResponse,
	type: EditorThemeType,
	appearance: ThemeAppearance
) {
	const nextResponse = cloneThemeResponse(response);
	appStore.state.themeExport.themeResult = nextResponse;
	appStore.state.themeExport.themeName = nextResponse.theme.name;
	appStore.state.themeExport.rawThemeOverrides = { ...nextResponse.rawThemeOverrides };
	appStore.state.themeExport.boostCoefficient = nextResponse.boostCoefficient;
	appStore.state.themeExport.themeColorsWithUsage = extractThemeColorsWithUsage(nextResponse.theme);
	cacheThemeVersion(nextResponse, type, appearance);
}

export function getCachedThemeVersion(
	type: EditorThemeType,
	appearance: ThemeAppearance
): ThemeGenerationResponse | null {
	return appStore.state.themeExport.themeVersions[getThemeVersionKey(type, appearance)] ?? null;
}

export function setManualOverrideFlag(key: keyof ThemeOverrides, enabled: boolean) {
	if (key === 'background') {
		appStore.state.themeExport.hasManualBackgroundOverride = enabled;
	}

	if (key === 'foreground') {
		appStore.state.themeExport.hasManualForegroundOverride = enabled;
	}
}

export function getCurrentOverrides(): ThemeOverrides {
	if (Object.keys(appStore.state.themeExport.rawThemeOverrides).length > 0) {
		return { ...appStore.state.themeExport.rawThemeOverrides };
	}

	if (appStore.state.themeExport.loadedThemeOverridesReference) {
		return { ...appStore.state.themeExport.loadedThemeOverridesReference };
	}

	return {};
}

export function buildRequestOverrides(
	nextAppearance: ThemeAppearance,
	currentAppearance: ThemeAppearance
): ThemeOverrides {
	const overrides = getCurrentOverrides();

	if (nextAppearance !== currentAppearance) {
		delete overrides.background;
		delete overrides.foreground;
	}

	return overrides;
}

export function resetThemeExportOverrideState(overrides: ThemeOverrides = {}) {
	clearThemeVersions();
	appStore.state.themeExport.rawThemeOverrides = { ...overrides };
	appStore.state.themeExport.hasManualBackgroundOverride = false;
	appStore.state.themeExport.hasManualForegroundOverride = false;
}

export function hydrateThemeExportResponse(
	response: ThemeGenerationResponse,
	type: EditorThemeType,
	appearance: ThemeAppearance
) {
	const activeResponse = cloneThemeResponse(response);
	const cachedResponse = cloneThemeResponse(response);

	appStore.setThemeExportEditorType(type);
	appStore.setThemeExportAppearance(appearance);
	appStore.state.themeExport.themeResult = activeResponse;
	appStore.state.themeExport.backupColors = activeResponse.colors;
	appStore.state.themeExport.themeName = activeResponse.theme.name;
	appStore.state.themeExport.themeVersions = {
		[getThemeVersionKey(type, appearance)]: cachedResponse
	};
	appStore.state.themeExport.rawThemeOverrides = { ...activeResponse.rawThemeOverrides };
	appStore.state.themeExport.boostCoefficient = activeResponse.boostCoefficient;
	appStore.state.themeExport.hasManualBackgroundOverride = false;
	appStore.state.themeExport.hasManualForegroundOverride = false;
	appStore.state.themeExport.loadedThemeOverridesReference = { ...activeResponse.rawThemeOverrides };
	appStore.state.themeExport.themeColorsWithUsage = extractThemeColorsWithUsage(activeResponse.theme);
}
