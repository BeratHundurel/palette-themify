import type { ApplyPaletteSettings } from '$lib/types/applyPaletteSettings';
import type { WallhavenSettings } from '$lib/types/wallhaven';
import type { ThemeExportPreferences } from './theme';

export type AppPreferencesPayload = Partial<{
	applyPaletteSettings: Partial<ApplyPaletteSettings>;
	wallhavenSettings: Partial<WallhavenSettings>;
	themeExport: Partial<ThemeExportPreferences>;
}>;

export type AppPreferences = {
	applyPaletteSettings: ApplyPaletteSettings;
	wallhavenSettings: WallhavenSettings;
	themeExport: ThemeExportPreferences;
};
