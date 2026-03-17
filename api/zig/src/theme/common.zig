const std = @import("std");

const color_utils = @import("../color/utils.zig");
const ThemeOverrides = @import("overrides.zig").ThemeOverrides;

const target_palette_size = 11;

pub const TerminalPalette = struct {
    foreground: []const u8,
    dim_foreground: []const u8,
    bright_foreground: []const u8,
    ansi_black: []const u8,
    ansi_white: []const u8,
    ansi_red: []const u8,
    ansi_green: []const u8,
    ansi_yellow: []const u8,
    ansi_blue: []const u8,
    ansi_magenta: []const u8,
    ansi_cyan: []const u8,
    ansi_bright_black: []const u8,
    ansi_bright_white: []const u8,
    ansi_bright_red: []const u8,
    ansi_bright_green: []const u8,
    ansi_bright_yellow: []const u8,
    ansi_bright_blue: []const u8,
    ansi_bright_magenta: []const u8,
    ansi_bright_cyan: []const u8,
    ansi_dim_black: []const u8,
    ansi_dim_white: []const u8,
    ansi_dim_red: []const u8,
    ansi_dim_green: []const u8,
    ansi_dim_yellow: []const u8,
    ansi_dim_blue: []const u8,
    ansi_dim_magenta: []const u8,
    ansi_dim_cyan: []const u8,
};

fn adjustTerminalVariant(base: []const u8, background: []const u8, amount: f32, make_brighter: bool, min_contrast: f32) []const u8 {
    const shifted = if (make_brighter)
        color_utils.lightenColor(base, amount)
    else
        color_utils.darkenColor(base, amount);

    return color_utils.ensureReadableContrast(shifted, background, min_contrast);
}

pub fn buildTerminalPalette(
    background: []const u8,
    foreground: []const u8,
    semantic_error: []const u8,
    semantic_success: []const u8,
    semantic_warning: []const u8,
    semantic_info: []const u8,
    c6: []const u8,
    c7: []const u8,
    dark_base: bool,
) TerminalPalette {
    const dim_foreground = color_utils.ensureReadableContrast(
        if (dark_base) color_utils.darkenColor(foreground, 0.22) else color_utils.lightenColor(foreground, 0.18),
        background,
        4.5,
    );
    const bright_foreground = color_utils.ensureReadableContrast(
        if (dark_base) color_utils.lightenColor(foreground, 0.10) else color_utils.darkenColor(foreground, 0.10),
        background,
        7.0,
    );

    const ansi_black = color_utils.ensureReadableContrast(
        if (dark_base) color_utils.darkenColor(foreground, 0.72) else color_utils.lightenColor(foreground, 0.72),
        background,
        2.2,
    );
    const ansi_white = dim_foreground;

    const ansi_red = color_utils.ensureReadableContrast(semantic_error, background, 3.0);
    const ansi_green = color_utils.ensureReadableContrast(semantic_success, background, 3.0);
    const ansi_yellow = color_utils.ensureReadableContrast(semantic_warning, background, 3.0);
    const ansi_blue = color_utils.ensureReadableContrast(semantic_info, background, 3.0);
    const ansi_magenta = color_utils.ensureReadableContrast(c6, background, 3.0);
    const ansi_cyan = color_utils.ensureReadableContrast(c7, background, 3.0);

    const ansi_bright_black = adjustTerminalVariant(ansi_black, background, if (dark_base) 0.28 else 0.20, dark_base, 3.0);
    const ansi_bright_white = bright_foreground;
    const ansi_bright_red = adjustTerminalVariant(ansi_red, background, 0.16, dark_base, 3.0);
    const ansi_bright_green = adjustTerminalVariant(ansi_green, background, 0.16, dark_base, 3.0);
    const ansi_bright_yellow = adjustTerminalVariant(ansi_yellow, background, 0.14, dark_base, 3.0);
    const ansi_bright_blue = adjustTerminalVariant(ansi_blue, background, 0.16, dark_base, 3.0);
    const ansi_bright_magenta = adjustTerminalVariant(ansi_magenta, background, 0.16, dark_base, 3.0);
    const ansi_bright_cyan = adjustTerminalVariant(ansi_cyan, background, 0.16, dark_base, 3.0);

    const ansi_dim_black = adjustTerminalVariant(ansi_black, background, if (dark_base) 0.06 else 0.10, !dark_base, 2.0);
    const ansi_dim_white = dim_foreground;
    const ansi_dim_red = adjustTerminalVariant(ansi_red, background, 0.18, !dark_base, 2.6);
    const ansi_dim_green = adjustTerminalVariant(ansi_green, background, 0.18, !dark_base, 2.6);
    const ansi_dim_yellow = adjustTerminalVariant(ansi_yellow, background, 0.16, !dark_base, 2.6);
    const ansi_dim_blue = adjustTerminalVariant(ansi_blue, background, 0.18, !dark_base, 2.6);
    const ansi_dim_magenta = adjustTerminalVariant(ansi_magenta, background, 0.18, !dark_base, 2.6);
    const ansi_dim_cyan = adjustTerminalVariant(ansi_cyan, background, 0.18, !dark_base, 2.6);

    return .{
        .foreground = foreground,
        .dim_foreground = dim_foreground,
        .bright_foreground = bright_foreground,
        .ansi_black = ansi_black,
        .ansi_white = ansi_white,
        .ansi_red = ansi_red,
        .ansi_green = ansi_green,
        .ansi_yellow = ansi_yellow,
        .ansi_blue = ansi_blue,
        .ansi_magenta = ansi_magenta,
        .ansi_cyan = ansi_cyan,
        .ansi_bright_black = ansi_bright_black,
        .ansi_bright_white = ansi_bright_white,
        .ansi_bright_red = ansi_bright_red,
        .ansi_bright_green = ansi_bright_green,
        .ansi_bright_yellow = ansi_bright_yellow,
        .ansi_bright_blue = ansi_bright_blue,
        .ansi_bright_magenta = ansi_bright_magenta,
        .ansi_bright_cyan = ansi_bright_cyan,
        .ansi_dim_black = ansi_dim_black,
        .ansi_dim_white = ansi_dim_white,
        .ansi_dim_red = ansi_dim_red,
        .ansi_dim_green = ansi_dim_green,
        .ansi_dim_yellow = ansi_dim_yellow,
        .ansi_dim_blue = ansi_dim_blue,
        .ansi_dim_magenta = ansi_dim_magenta,
        .ansi_dim_cyan = ansi_dim_cyan,
    };
}

pub const ThemeType = enum {
    vscode,
    zed,
};

pub const ThemeAppearance = enum {
    dark,
    light,
};

pub const GenerateOverridableRequest = struct {
    theme: std.json.Value,
    ThemeOverrides: ?ThemeOverrides = null,
    themeType: ThemeType = .zed,
    appearance: ?ThemeAppearance = null,
    boostCoefficient: ?f32 = null,
};

pub const PreparedThemeSelection = struct {
    semantic: color_utils.SemanticColors,
    dark_base: bool,
    background_seed: []const u8,
    foreground_seed: []const u8,
    c1_raw: []const u8,
    c2_raw: []const u8,
    c3_raw: []const u8,
    c4_raw: []const u8,
    c5_raw: []const u8,
    c6_raw: []const u8,
    c7_raw: []const u8,
    c8_raw: []const u8,
    c9_raw: []const u8,
};

pub fn resolveAccent(raw: []const u8, background: []const u8, boost_coefficient: f32) []const u8 {
    return color_utils.boostAccentColor(color_utils.adjustForContrast(raw, background, 3), background, boost_coefficient);
}

pub fn prepareThemeSelection(
    allocator: std.mem.Allocator,
    colors: []const []const u8,
    overrides: ThemeOverrides,
    appearance: ?ThemeAppearance,
) !PreparedThemeSelection {
    if (colors.len == 0) {
        return error.NotEnoughColors;
    }

    const semantic = color_utils.findSemanticColors(colors);
    const improved_colors = try color_utils.selectDiverseColors(allocator, colors, target_palette_size);
    defer allocator.free(improved_colors);

    var palette = std.ArrayList([]const u8){};
    defer palette.deinit(allocator);
    try palette.ensureTotalCapacity(allocator, target_palette_size);
    try palette.appendSlice(allocator, improved_colors);

    if (palette.items.len < target_palette_size) {
        const harmony_schemes = [_]color_utils.HarmonyScheme{
            .complementary,
            .triadic,
            .analogous,
            .@"split-complementary",
        };
        const base_count = palette.items.len;
        var scheme_index: usize = 0;
        var base_index: usize = 0;

        while (palette.items.len < target_palette_size) {
            const base_color = palette.items[base_index % base_count];
            const scheme = harmony_schemes[scheme_index % harmony_schemes.len];
            const harmonic = color_utils.getHarmonicColor(base_color, scheme);
            try palette.append(allocator, harmonic);
            scheme_index += 1;
            base_index += 1;
        }
    }

    const dark_base = if (appearance) |value| value == .dark else true;
    const selection = try color_utils.selectBackgroundAndForeground(allocator, palette.items, dark_base);
    defer allocator.free(selection.remaining_indices);

    const remaining = selection.remaining_indices;

    return .{
        .semantic = semantic,
        .dark_base = dark_base,
        .background_seed = palette.items[selection.background_index],
        .foreground_seed = palette.items[selection.foreground_index],
        .c1_raw = overrides.c1 orelse palette.items[remaining[0]],
        .c2_raw = overrides.c2 orelse palette.items[remaining[1]],
        .c3_raw = overrides.c3 orelse palette.items[remaining[2]],
        .c4_raw = overrides.c4 orelse palette.items[remaining[3]],
        .c5_raw = overrides.c5 orelse palette.items[remaining[4]],
        .c6_raw = overrides.c6 orelse palette.items[remaining[5]],
        .c7_raw = overrides.c7 orelse palette.items[remaining[6]],
        .c8_raw = overrides.c8 orelse palette.items[remaining[7]],
        .c9_raw = overrides.c9 orelse palette.items[remaining[8]],
    };
}
