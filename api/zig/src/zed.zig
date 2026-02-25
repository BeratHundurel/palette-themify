const std = @import("std");
const builtin = @import("builtin");
const types = @import("zed_types.zig");
const ThemeOverrides = @import("theme_overrides.zig").ThemeOverrides;
const color_utils = @import("color_utils.zig");
const GenerateOverridableRequest = @import("palette_api.zig").GenerateOverridableRequest;

const ZedTheme = types.ZedTheme;
const ZedThemeStyle = types.ZedThemeStyle;
const ZedThemeEntry = types.ZedThemeEntry;
const ZedThemeResponse = types.ZedThemeResponse;
const ZedSyntax = types.ZedSyntax;
const SyntaxStyle = types.SyntaxStyle;
const Player = types.Player;
const ColorHex = types.ColorHex;

pub fn generateZedTheme(
    allocator: std.mem.Allocator,
    colors: []const []const u8,
    theme_name: []const u8,
    overrides: ThemeOverrides,
) !ZedThemeResponse {
    if (colors.len == 0) {
        return error.NotEnoughColors;
    }

    const semantic = color_utils.findSemanticColors(colors);
    const improved_colors = try color_utils.selectDiverseColors(allocator, colors, 11);
    defer allocator.free(improved_colors);

    var palette = std.ArrayList([]const u8){};
    defer palette.deinit(allocator);
    try palette.ensureTotalCapacity(allocator, 11);
    try palette.appendSlice(allocator, improved_colors);

    if (palette.items.len < 11) {
        const harmony_schemes = [_]color_utils.HarmonyScheme{
            .complementary,
            .triadic,
            .analogous,
            .@"split-complementary",
        };
        const base_count = palette.items.len;
        var scheme_index: usize = 0;
        var base_index: usize = 0;

        while (palette.items.len < 11) {
            const base_color = palette.items[base_index % base_count];
            const scheme = harmony_schemes[scheme_index % harmony_schemes.len];
            const harmonic = color_utils.getHarmonicColor(base_color, scheme);
            try palette.append(allocator, harmonic);
            scheme_index += 1;
            base_index += 1;
        }
    }

    var sum_luminance: f32 = 0.0;
    for (palette.items) |color| {
        sum_luminance += color_utils.getLuminance(color);
    }
    const average_luminance = sum_luminance / @as(f32, @floatFromInt(palette.items.len));
    const dark_base = average_luminance < 128.0;

    const selection = try color_utils.selectBackgroundAndForeground(allocator, palette.items, dark_base);
    defer allocator.free(selection.remaining_indices);

    const remaining = selection.remaining_indices;
    const c1_raw = overrides.c1 orelse palette.items[remaining[0]];
    const c2_raw = overrides.c2 orelse palette.items[remaining[1]];
    const c3_raw = overrides.c3 orelse palette.items[remaining[2]];
    const c4_raw = overrides.c4 orelse palette.items[remaining[3]];
    const c5_raw = overrides.c5 orelse palette.items[remaining[4]];
    const c6_raw = overrides.c6 orelse palette.items[remaining[5]];
    const c7_raw = overrides.c7 orelse palette.items[remaining[6]];
    const c8_raw = overrides.c8 orelse palette.items[remaining[7]];
    const c9_raw = overrides.c9 orelse palette.items[remaining[8]];

    const background = if (overrides.background) |bg| bg else blk: {
        const bg_raw = palette.items[selection.background_index];
        const base_luminance = color_utils.getLuminance(bg_raw);
        const darken_amount = if (dark_base) 0.6 + (base_luminance) * 0.2 else 0.0;
        const lighten_amount = if (dark_base) 0.0 else 0.6 + (1.0 - base_luminance) * 0.2;
        break :blk if (dark_base) color_utils.darkenColor(bg_raw, darken_amount) else color_utils.lightenColor(bg_raw, lighten_amount);
    };
    const bg_dark = if (dark_base) color_utils.darkenColor(background, 0.20) else color_utils.lightenColor(background, 0.20);
    const bg_very_dark = if (dark_base) color_utils.darkenColor(background, 0.30) else color_utils.lightenColor(background, 0.30);
    const bg_light = if (dark_base) color_utils.lightenColor(background, 0.10) else color_utils.darkenColor(background, 0.10);
    const bg_lighter = if (dark_base) color_utils.lightenColor(background, 0.20) else color_utils.darkenColor(background, 0.20);

    const proposed_foreground = overrides.foreground orelse palette.items[selection.foreground_index];
    const foreground = color_utils.ensureReadableContrast(proposed_foreground, background, 7.0);

    const fg_muted = if (dark_base) color_utils.darkenColor(foreground, 0.50) else color_utils.lightenColor(foreground, 0.50);
    const fg_disabled = if (dark_base) color_utils.darkenColor(foreground, 0.60) else color_utils.lightenColor(foreground, 0.60);
    const fg_placeholder = if (dark_base) color_utils.darkenColor(foreground, 0.70) else color_utils.lightenColor(foreground, 0.70);

    const fg_12 = color_utils.addAlpha(foreground, "12");
    const fg_26 = color_utils.addAlpha(foreground, "26");
    const fg_40 = color_utils.addAlpha(foreground, "40");
    const fg_66 = color_utils.addAlpha(foreground, "66");
    const fg_80 = color_utils.addAlpha(foreground, "80");

    const c1 = color_utils.adjustForContrast(c1_raw, background, 3);
    const c2 = color_utils.adjustForContrast(c2_raw, background, 3);
    const c3 = color_utils.adjustForContrast(c3_raw, background, 3);
    const c4 = color_utils.adjustForContrast(c4_raw, background, 3);
    const c5 = color_utils.adjustForContrast(c5_raw, background, 3);
    const c6 = color_utils.adjustForContrast(c6_raw, background, 3);
    const c7 = color_utils.adjustForContrast(c7_raw, background, 3);
    const c8 = color_utils.adjustForContrast(c8_raw, background, 3);
    const c9 = color_utils.adjustForContrast(c9_raw, background, 3);

    const constants_raw = overrides.constants orelse color_utils.getHarmonicColor(c2, .@"split-complementary");
    const constants = color_utils.adjustForContrast(constants_raw, background, 3);

    const semantic_error = color_utils.adjustForContrast(semantic.error_color, background, 3);
    const semantic_warning = color_utils.adjustForContrast(semantic.warning_color, background, 3);
    const semantic_success = color_utils.adjustForContrast(semantic.success_color, background, 3);
    const semantic_info = color_utils.adjustForContrast(semantic.info_color, background, 3);

    const c2_33 = color_utils.addAlpha(c2, "33");
    const c2_40 = color_utils.addAlpha(c2, "40");
    const c2_66 = color_utils.addAlpha(c2, "66");
    const c2_88 = color_utils.addAlpha(c2, "88");
    const c2_99 = color_utils.addAlpha(c2, "99");
    const c3_33 = color_utils.addAlpha(c3, "33");

    const semantic_error_26 = color_utils.addAlpha(semantic_error, "26");
    const semantic_error_1f = color_utils.addAlpha(semantic_error, "1f");
    const semantic_warning_26 = color_utils.addAlpha(semantic_warning, "26");
    const semantic_warning_1f = color_utils.addAlpha(semantic_warning, "1f");
    const semantic_success_26 = color_utils.addAlpha(semantic_success, "26");
    const semantic_success_1f = color_utils.addAlpha(semantic_success, "1f");
    const semantic_success_88 = color_utils.addAlpha(semantic_success, "88");

    const accent_bright = if (dark_base) color_utils.lightenColor(c2, 0.33) else color_utils.darkenColor(c2, 0.33);

    const accents = try allocator.alloc([]const u8, 8);
    accents[0] = c1;
    accents[1] = c2;
    accents[2] = c3;
    accents[3] = c4;
    accents[4] = c5;
    accents[5] = c6;
    accents[6] = c7;
    accents[7] = c8;

    const players = try allocator.alloc(Player, 8);
    players[0] = .{ .cursor = foreground, .selection = fg_40, .background = foreground };
    players[1] = .{ .cursor = c2, .selection = c2_40, .background = c2 };
    players[2] = .{ .cursor = c3, .selection = c3_33, .background = c3 };
    players[3] = .{ .cursor = c4, .selection = color_utils.addAlpha(c4, "40"), .background = c4 };
    players[4] = .{ .cursor = c5, .selection = color_utils.addAlpha(c5, "40"), .background = c5 };
    players[5] = .{ .cursor = c6, .selection = color_utils.addAlpha(c6, "40"), .background = c6 };
    players[6] = .{ .cursor = c7, .selection = color_utils.addAlpha(c7, "40"), .background = c7 };
    players[7] = .{ .cursor = c8, .selection = color_utils.addAlpha(c8, "40"), .background = c8 };

    const style = ZedThemeStyle{
        .accents = accents,

        .@"vim.mode.text" = bg_very_dark,
        .@"vim.normal.foreground" = bg_very_dark,
        .@"vim.helix_normal.foreground" = bg_very_dark,
        .@"vim.normal.background" = foreground,
        .@"vim.helix_normal.background" = foreground,
        .@"vim.visual.background" = c2,
        .@"vim.helix_select.background" = c2,
        .@"vim.insert.background" = semantic_success,
        .@"vim.visual_line.background" = c2,
        .@"vim.visual_block.background" = c3,
        .@"vim.replace.background" = semantic_error,

        .@"background.appearance" = .@"opaque",

        .border = bg_lighter,
        .@"border.variant" = c2_88,
        .@"border.focused" = c2_88,
        .@"border.selected" = c2_88,
        .@"border.transparent" = semantic_success_88,
        .@"border.disabled" = fg_disabled,

        .@"elevated_surface.background" = bg_dark,
        .@"surface.background" = bg_dark,
        .background = background,

        .@"element.background" = bg_very_dark,
        .@"element.hover" = bg_lighter,
        .@"element.active" = color_utils.addAlpha(bg_lighter, "4d"),
        .@"element.selected" = color_utils.addAlpha(bg_lighter, "4d"),
        .@"element.disabled" = fg_disabled,
        .@"drop_target.background" = color_utils.addAlpha(bg_lighter, "66"),

        .@"ghost_element.background" = "#00000000",
        .@"ghost_element.hover" = bg_light,
        .@"ghost_element.active" = bg_lighter,
        .@"ghost_element.selected" = fg_muted,
        .@"ghost_element.disabled" = fg_disabled,

        .text = foreground,
        .@"text.muted" = fg_muted,
        .@"text.placeholder" = fg_placeholder,
        .@"text.disabled" = fg_disabled,
        .@"text.accent" = c2,

        .icon = foreground,
        .@"icon.muted" = fg_muted,
        .@"icon.disabled" = fg_disabled,
        .@"icon.placeholder" = fg_placeholder,
        .@"icon.accent" = c2,

        .@"status_bar.background" = bg_very_dark,
        .@"title_bar.background" = bg_very_dark,
        .@"title_bar.inactive_background" = if (dark_base) color_utils.darkenColor(bg_very_dark, 0.30) else color_utils.lightenColor(bg_very_dark, 0.30),
        .@"toolbar.background" = background,

        .@"tab_bar.background" = bg_very_dark,
        .@"tab.inactive_background" = if (dark_base) color_utils.darkenColor(bg_very_dark, 0.30) else color_utils.lightenColor(bg_very_dark, 0.30),
        .@"tab.active_background" = background,

        .@"search.match_background" = c3_33,

        .@"panel.background" = bg_dark,
        .@"panel.focused_border" = foreground,
        .@"panel.indent_guide" = fg_placeholder,
        .@"panel.indent_guide_active" = fg_80,
        .@"panel.indent_guide_hover" = c2,
        .@"panel.overlay_background" = bg_very_dark,

        .@"pane.focused_border" = foreground,
        .@"pane_group.border" = bg_lighter,

        .@"scrollbar.thumb.background" = color_utils.addAlpha(fg_placeholder, "80"),
        .@"scrollbar.thumb.hover_background" = fg_muted,
        .@"scrollbar.thumb.active_background" = null,
        .@"scrollbar.thumb.border" = null,
        .@"scrollbar.track.background" = bg_very_dark,
        .@"scrollbar.track.border" = fg_12,

        .@"minimap.thumb.background" = c2_33,
        .@"minimap.thumb.hover_background" = c2_66,
        .@"minimap.thumb.active_background" = c2_99,
        .@"minimap.thumb.border" = null,

        .@"editor.foreground" = foreground,
        .@"editor.background" = background,
        .@"editor.gutter.background" = background,
        .@"editor.subheader.background" = bg_dark,
        .@"editor.active_line.background" = fg_12,
        .@"editor.highlighted_line.background" = null,
        .@"editor.line_number" = fg_muted,
        .@"editor.active_line_number" = c2,
        .@"editor.invisible" = fg_66,
        .@"editor.wrap_guide" = fg_placeholder,
        .@"editor.active_wrap_guide" = fg_placeholder,
        .@"editor.document_highlight.bracket_background" = color_utils.addAlpha(c2, "17"),
        .@"editor.document_highlight.read_background" = fg_26,
        .@"editor.document_highlight.write_background" = fg_26,
        .@"editor.indent_guide" = fg_placeholder,
        .@"editor.indent_guide_active" = fg_placeholder,

        .@"terminal.background" = background,
        .@"terminal.ansi.background" = background,
        .@"terminal.foreground" = foreground,
        .@"terminal.dim_foreground" = fg_muted,
        .@"terminal.bright_foreground" = foreground,
        .@"terminal.ansi.black" = if (dark_base) color_utils.darkenColor(foreground, 0.7) else color_utils.lightenColor(foreground, 0.7),
        .@"terminal.ansi.white" = fg_muted,
        .@"terminal.ansi.red" = semantic_error,
        .@"terminal.ansi.green" = semantic_success,
        .@"terminal.ansi.yellow" = semantic_warning,
        .@"terminal.ansi.blue" = semantic_info,
        .@"terminal.ansi.magenta" = c6,
        .@"terminal.ansi.cyan" = c7,
        .@"terminal.ansi.bright_black" = fg_placeholder,
        .@"terminal.ansi.bright_white" = fg_muted,
        .@"terminal.ansi.bright_red" = if (dark_base) color_utils.lightenColor(semantic_error, 0.1) else color_utils.darkenColor(semantic_error, 0.1),
        .@"terminal.ansi.bright_green" = if (dark_base) color_utils.lightenColor(semantic_success, 0.1) else color_utils.darkenColor(semantic_success, 0.1),
        .@"terminal.ansi.bright_yellow" = if (dark_base) color_utils.lightenColor(semantic_warning, 0.1) else color_utils.darkenColor(semantic_warning, 0.1),
        .@"terminal.ansi.bright_blue" = if (dark_base) color_utils.lightenColor(semantic_info, 0.1) else color_utils.darkenColor(semantic_info, 0.1),
        .@"terminal.ansi.bright_magenta" = if (dark_base) color_utils.lightenColor(c6, 0.1) else color_utils.darkenColor(c6, 0.1),
        .@"terminal.ansi.bright_cyan" = if (dark_base) color_utils.lightenColor(c7, 0.1) else color_utils.darkenColor(c7, 0.1),
        .@"terminal.ansi.dim_black" = if (dark_base) color_utils.darkenColor(foreground, 0.7) else color_utils.lightenColor(foreground, 0.7),
        .@"terminal.ansi.dim_white" = fg_muted,
        .@"terminal.ansi.dim_red" = semantic_error,
        .@"terminal.ansi.dim_green" = semantic_success,
        .@"terminal.ansi.dim_yellow" = semantic_warning,
        .@"terminal.ansi.dim_blue" = semantic_info,
        .@"terminal.ansi.dim_magenta" = c6,
        .@"terminal.ansi.dim_cyan" = c7,

        .@"link_text.hover" = accent_bright,

        .conflict = semantic_warning,
        .@"conflict.border" = semantic_warning,
        .@"conflict.background" = semantic_warning_26,
        .created = semantic_success,
        .@"created.border" = semantic_success,
        .@"created.background" = semantic_success_26,
        .deleted = semantic_error,
        .@"deleted.border" = semantic_error,
        .@"deleted.background" = semantic_error_26,
        .hidden = fg_disabled,
        .@"hidden.border" = fg_disabled,
        .@"hidden.background" = bg_dark,
        .hint = fg_placeholder,
        .@"hint.border" = fg_placeholder,
        .@"hint.background" = bg_dark,
        .ignored = fg_disabled,
        .@"ignored.border" = fg_disabled,
        .@"ignored.background" = color_utils.addAlpha(fg_disabled, "26"),
        .modified = color_utils.lightenColor(semantic_warning, 0.33),
        .@"modified.border" = color_utils.lightenColor(semantic_warning, 0.33),
        .@"modified.background" = color_utils.lightenColor(semantic_warning_26, 0.33),
        .predictive = fg_disabled,
        .@"predictive.border" = c2,
        .@"predictive.background" = bg_dark,
        .renamed = semantic_info,
        .@"renamed.border" = semantic_info,
        .@"renamed.background" = color_utils.addAlpha(semantic_info, "26"),
        .info = semantic_info,
        .@"info.border" = semantic_info,
        .@"info.background" = bg_light,
        .warning = semantic_warning,
        .@"warning.border" = semantic_warning,
        .@"warning.background" = semantic_warning_1f,
        .@"error" = semantic_error,
        .@"error.border" = semantic_error,
        .@"error.background" = semantic_error_1f,
        .success = semantic_success,
        .@"success.border" = semantic_success,
        .@"success.background" = semantic_success_1f,
        .@"unreachable" = semantic_error,
        .@"unreachable.border" = semantic_error,
        .@"unreachable.background" = semantic_error_1f,

        .players = players,

        .@"version_control.added" = semantic_success,
        .@"version_control.deleted" = semantic_error,
        .@"version_control.modified" = color_utils.lightenColor(semantic_warning, 0.33),
        .@"version_control.renamed" = semantic_info,
        .@"version_control.conflict" = semantic_warning,
        .@"version_control.conflict_marker.ours" = color_utils.addAlpha(semantic_success, "33"),
        .@"version_control.conflict_marker.theirs" = color_utils.addAlpha(semantic_info, "33"),
        .@"version_control.ignored" = fg_disabled,

        .@"debugger.accent" = semantic_error,
        .@"editor.debugger_active_line.background" = color_utils.addAlpha(semantic_warning, "12"),

        .syntax = ZedSyntax{
            .attribute = .{ .color = constants, .font_style = null, .font_weight = null },
            .boolean = .{ .color = constants, .font_style = null, .font_weight = null },
            .character = .{ .color = c7, .font_style = null, .font_weight = null },
            .@"character.special" = .{ .color = c4, .font_style = null, .font_weight = null },
            .comment = .{ .color = fg_muted, .font_style = .italic, .font_weight = null },
            .@"comment.documentation" = .{ .color = semantic_info, .font_style = .italic, .font_weight = null },
            .@"comment.error" = .{ .color = semantic_error, .font_style = .italic, .font_weight = null },
            .@"comment.hint" = .{ .color = semantic_info, .font_style = .italic, .font_weight = null },
            .@"comment.note" = .{ .color = foreground, .font_style = .italic, .font_weight = null },
            .@"comment.todo" = .{ .color = c8, .font_style = .italic, .font_weight = null },
            .@"comment.warning" = .{ .color = semantic_warning, .font_style = .italic, .font_weight = null },
            .concept = .{ .color = semantic_info, .font_style = null, .font_weight = null },
            .constant = .{ .color = constants, .font_style = null, .font_weight = null },
            .@"constant.macro" = .{ .color = c6, .font_style = null, .font_weight = null },
            .constructor = .{ .color = c8, .font_style = null, .font_weight = null },
            .@"diff.minus" = .{ .color = semantic_error, .font_style = null, .font_weight = null },
            .@"diff.plus" = .{ .color = semantic_success, .font_style = null, .font_weight = null },
            .embedded = .{ .color = c7, .font_style = null, .font_weight = null },
            .emphasis = .{ .color = c7, .font_style = null, .font_weight = null },
            .@"emphasis.strong" = .{ .color = c7, .font_style = null, .font_weight = 700 },
            .@"enum" = .{ .color = c7, .font_style = null, .font_weight = 700 },
            .field = .{ .color = c1, .font_style = null, .font_weight = null },
            .float = .{ .color = constants, .font_style = null, .font_weight = null },
            .function = .{ .color = c2, .font_style = null, .font_weight = null },
            .@"function.decorator" = .{ .color = constants, .font_style = null, .font_weight = null },
            .hint = .{ .color = fg_muted, .font_style = null, .font_weight = null },
            .keyword = .{ .color = c6, .font_style = null, .font_weight = null },
            .@"keyword.directive" = .{ .color = c4, .font_style = null, .font_weight = null },
            .@"keyword.export" = .{ .color = accent_bright, .font_style = null, .font_weight = null },
            .label = .{ .color = semantic_info, .font_style = null, .font_weight = null },
            .link_text = .{ .color = c1, .font_style = null, .font_weight = null },
            .link_uri = .{ .color = c2, .font_style = .italic, .font_weight = null },
            .module = .{ .color = c5, .font_style = null, .font_weight = null },
            .namespace = .{ .color = c5, .font_style = null, .font_weight = null },
            .number = .{ .color = constants, .font_style = null, .font_weight = null },
            .operator = .{ .color = accent_bright, .font_style = null, .font_weight = null },
            .parameter = .{ .color = c7, .font_style = null, .font_weight = null },
            .parent = .{ .color = constants, .font_style = null, .font_weight = null },
            .predictive = .{ .color = fg_disabled, .font_style = null, .font_weight = null },
            .predoc = .{ .color = c9, .font_style = null, .font_weight = null },
            .preproc = .{ .color = c6, .font_style = null, .font_weight = null },
            .primary = .{ .color = c7, .font_style = null, .font_weight = null },
            .property = .{ .color = c1, .font_style = null, .font_weight = null },
            .punctuation = .{ .color = fg_muted, .font_style = null, .font_weight = null },
            .@"punctuation.list_marker" = .{ .color = c7, .font_style = null, .font_weight = null },
            .@"punctuation.markup" = .{ .color = c9, .font_style = null, .font_weight = null },
            .@"punctuation.special" = .{ .color = c4, .font_style = null, .font_weight = null },
            .@"punctuation.special.symbol" = .{ .color = c8, .font_style = null, .font_weight = null },
            .selector = .{ .color = c5, .font_style = null, .font_weight = null },
            .@"selector.pseudo" = .{ .color = semantic_info, .font_style = null, .font_weight = null },
            .string = .{ .color = c3, .font_style = null, .font_weight = null },
            .@"string.documentation" = .{ .color = semantic_info, .font_style = null, .font_weight = null },
            .@"string.escape" = .{ .color = c4, .font_style = null, .font_weight = null },
            .@"string.regex" = .{ .color = constants, .font_style = null, .font_weight = null },
            .@"string.special" = .{ .color = c4, .font_style = null, .font_weight = null },
            .@"string.special.symbol" = .{ .color = c8, .font_style = null, .font_weight = null },
            .@"string.special.url" = .{ .color = foreground, .font_style = .italic, .font_weight = null },
            .symbol = .{ .color = c4, .font_style = null, .font_weight = null },
            .tag = .{ .color = c2, .font_style = null, .font_weight = null },
            .@"tag.attribute" = .{ .color = c5, .font_style = .italic, .font_weight = null },
            .@"tag.delimiter" = .{ .color = c7, .font_style = null, .font_weight = null },
            .@"tag.doctype" = .{ .color = c6, .font_style = null, .font_weight = null },
            .text = .{ .color = foreground, .font_style = null, .font_weight = null },
            .@"text.literal" = .{ .color = c3, .font_style = null, .font_weight = null },
            .title = .{ .color = foreground, .font_style = null, .font_weight = 800 },
            .type = .{ .color = c5, .font_style = null, .font_weight = null },
            .@"type.class.definition" = .{ .color = c5, .font_style = null, .font_weight = 700 },
            .variable = .{ .color = foreground, .font_style = null, .font_weight = null },
            .@"variable.builtin" = .{ .color = c9, .font_style = null, .font_weight = null },
            .@"variable.member" = .{ .color = c1, .font_style = null, .font_weight = null },
            .@"variable.parameter" = .{ .color = c7, .font_style = null, .font_weight = null },
            .@"variable.special" = .{ .color = c9, .font_style = null, .font_weight = null },
            .variant = .{ .color = c9, .font_style = null, .font_weight = null },
        },
    };

    const themes = try allocator.alloc(ZedThemeEntry, 1);
    const backupColors = try allocator.alloc(ColorHex, colors.len);
    for (colors, 0..) |c, i| backupColors[i] = .{ .hex = c };
    themes[0] = .{
        .name = theme_name,
        .appearance = if (dark_base) .dark else .light,
        .style = style,
    };

    const theme = ZedTheme{
        .@"$schema" = "https://zed.dev/schema/themes/v0.2.0.json",
        .name = theme_name,
        .author = "Palette Themify",
        .themes = themes,
    };

    const base_overrides = ThemeOverrides{
        .background = background,
        .foreground = foreground,
        .c1 = c1,
        .c2 = c2,
        .c3 = c3,
        .c4 = c4,
        .c5 = c5,
        .c6 = c6,
        .c7 = c7,
        .c8 = c8,
        .c9 = c9,
        .constants = constants,
    };

    return ZedThemeResponse{ .theme = theme, .themeOverrides = base_overrides, .colors = backupColors };
}

pub fn generateOverridableFromZedThemeValue(allocator: std.mem.Allocator, request: GenerateOverridableRequest) !ZedThemeResponse {
    const root_obj = switch (request.theme) {
        .object => |obj| obj,
        else => return error.InvalidTheme,
    };

    const themes_value = root_obj.get("themes") orelse return error.InvalidTheme;
    const themes_array = switch (themes_value) {
        .array => |arr| arr,
        else => return error.InvalidTheme,
    };
    if (themes_array.items.len == 0) {
        return error.InvalidTheme;
    }

    const first_theme_value = themes_array.items[0];
    const theme_obj = switch (first_theme_value) {
        .object => |obj| obj,
        else => return error.InvalidTheme,
    };

    const style_obj = color_utils.getObjectField(theme_obj, "style") orelse return error.InvalidTheme;

    var theme_name: []const u8 = "Generated Theme";
    if (color_utils.getStringField(root_obj, "name")) |name| {
        if (name.len > 0) theme_name = name;
    } else if (color_utils.getStringField(theme_obj, "name")) |name| {
        if (name.len > 0) theme_name = name;
    }

    var overrides = request.ThemeOverrides orelse ThemeOverrides{};
    var colors = std.ArrayList([]const u8){};
    defer colors.deinit(allocator);

    if (color_utils.getStringField(style_obj, "background")) |background| {
        const value = overrides.background orelse background;
        overrides.background = value;
        try color_utils.addColor(allocator, &colors, value);
    } else if (overrides.background) |value| {
        overrides.background = value;
        try color_utils.addColor(allocator, &colors, value);
    }
    if (color_utils.getStringField(style_obj, "text")) |foreground| {
        const value = overrides.foreground orelse foreground;
        overrides.foreground = value;
        try color_utils.addColor(allocator, &colors, value);
    } else if (overrides.foreground) |value| {
        overrides.foreground = value;
        try color_utils.addColor(allocator, &colors, value);
    }

    if (color_utils.getArrayField(style_obj, "accents")) |accents| {
        const max_len = @min(accents.items.len, 8);
        for (accents.items[0..max_len], 0..) |item, idx| {
            const accent = switch (item) {
                .string => |str| str,
                else => continue,
            };
            const accent_value = switch (idx) {
                0 => overrides.c1 orelse accent,
                1 => overrides.c2 orelse accent,
                2 => overrides.c3 orelse accent,
                3 => overrides.c4 orelse accent,
                4 => overrides.c5 orelse accent,
                5 => overrides.c6 orelse accent,
                6 => overrides.c7 orelse accent,
                7 => overrides.c8 orelse accent,
                else => accent,
            };

            switch (idx) {
                0 => overrides.c1 = accent_value,
                1 => overrides.c2 = accent_value,
                2 => overrides.c3 = accent_value,
                3 => overrides.c4 = accent_value,
                4 => overrides.c5 = accent_value,
                5 => overrides.c6 = accent_value,
                6 => overrides.c7 = accent_value,
                7 => overrides.c8 = accent_value,
                else => {},
            }
            try color_utils.addColor(allocator, &colors, accent_value);
        }
    }

    if (color_utils.getFirstStringField(style_obj, &.{ "error", "deleted", "conflict" })) |value| {
        try color_utils.addColor(allocator, &colors, value);
    }
    if (color_utils.getFirstStringField(style_obj, &.{ "warning", "modified", "conflict" })) |value| {
        try color_utils.addColor(allocator, &colors, value);
    }
    if (color_utils.getFirstStringField(style_obj, &.{ "success", "created" })) |value| {
        try color_utils.addColor(allocator, &colors, value);
    }
    if (color_utils.getFirstStringField(style_obj, &.{ "info", "renamed" })) |value| {
        try color_utils.addColor(allocator, &colors, value);
    }

    if (color_utils.getObjectField(style_obj, "syntax")) |syntax_obj| {
        if (color_utils.getObjectField(syntax_obj, "constant")) |constant_obj| {
            if (color_utils.getStringField(constant_obj, "color")) |value| {
                if (overrides.constants == null) overrides.constants = value;
                try color_utils.addColor(allocator, &colors, value);
            }
        }
        if (color_utils.getObjectField(syntax_obj, "number")) |number_obj| {
            if (color_utils.getStringField(number_obj, "color")) |value| {
                try color_utils.addColor(allocator, &colors, value);
            }
        }
        if (overrides.c9 == null) {
            if (color_utils.getObjectField(syntax_obj, "variable.builtin")) |builtin_obj| {
                if (color_utils.getStringField(builtin_obj, "color")) |value| {
                    overrides.c9 = value;
                    try color_utils.addColor(allocator, &colors, value);
                }
            }
        }
        if (overrides.c9 == null) {
            if (color_utils.getObjectField(syntax_obj, "variable.special")) |special_obj| {
                if (color_utils.getStringField(special_obj, "color")) |value| {
                    overrides.c9 = value;
                    try color_utils.addColor(allocator, &colors, value);
                }
            }
        }
        if (overrides.c9 == null) {
            if (color_utils.getObjectField(syntax_obj, "predoc")) |predoc_obj| {
                if (color_utils.getStringField(predoc_obj, "color")) |value| {
                    overrides.c9 = value;
                    try color_utils.addColor(allocator, &colors, value);
                }
            }
        }
        if (overrides.c9 == null) {
            if (color_utils.getObjectField(syntax_obj, "variant")) |variant_obj| {
                if (color_utils.getStringField(variant_obj, "color")) |value| {
                    overrides.c9 = value;
                    try color_utils.addColor(allocator, &colors, value);
                }
            }
        }
    }

    if (colors.items.len == 0) {
        return error.InvalidTheme;
    }

    const palette = try allocator.dupe([]const u8, colors.items);
    defer allocator.free(palette);

    return try generateZedTheme(allocator, palette, theme_name, overrides);
}

test "generateOverridableFromZedThemeValue rejects invalid json" {
    var arena = std.heap.ArenaAllocator.init(std.testing.allocator);
    defer arena.deinit();
    const allocator = arena.allocator();
    const bad_root = std.json.Value{ .string = "not an object" };
    const request = GenerateOverridableRequest{ .theme = bad_root };
    try std.testing.expectError(error.InvalidTheme, generateOverridableFromZedThemeValue(allocator, request));
}

test "generateOverridableFromZedThemeValue builds overrides" {
    var arena = std.heap.ArenaAllocator.init(std.testing.allocator);
    defer arena.deinit();
    const allocator = arena.allocator();
    const json =
        "{" ++ "\"name\":\"Sample\"," ++ "\"themes\":[{" ++ "\"name\":\"Sample\"," ++ "\"appearance\":\"dark\"," ++ "\"style\":{" ++ "\"background\":\"#001122\"," ++ "\"text\":\"#DDEEFF\"," ++ "\"accents\":[\"#AA0000\",\"#00AA00\",\"#0000AA\",\"#AAAA00\",\"#00AAAA\",\"#AA00AA\",\"#666666\",\"#999999\"]," ++ "\"error\":\"#FF0000\"," ++ "\"warning\":\"#FFAA00\"," ++ "\"success\":\"#00FF00\"," ++ "\"info\":\"#0099FF\"," ++ "\"syntax\":{\"constant\":{\"color\":\"#FF00FF\"}}" ++ "}" ++ "}]" ++ "}";

    const parsed = try std.json.parseFromSlice(std.json.Value, allocator, json, .{});
    defer parsed.deinit();

    const request = GenerateOverridableRequest{
        .theme = parsed.value,
        .ThemeOverrides = ThemeOverrides{ .c1 = "#010203" },
    };
    const response = try generateOverridableFromZedThemeValue(allocator, request);

    try std.testing.expect(response.themeOverrides.background != null);
    try std.testing.expect(response.themeOverrides.foreground != null);
    try std.testing.expect(response.themeOverrides.c1 != null);

    const expected_c1 = color_utils.adjustForContrast("#010203", response.themeOverrides.background.?, 3);
    try std.testing.expectEqualStrings(expected_c1, response.themeOverrides.c1.?);
}
