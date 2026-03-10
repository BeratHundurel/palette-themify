const std = @import("std");

const color_utils = @import("../color/utils.zig");
const ThemeOverrides = @import("overrides.zig").ThemeOverrides;

const target_palette_size = 11;

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
