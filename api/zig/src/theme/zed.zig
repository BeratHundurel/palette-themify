const std = @import("std");
const builtin = @import("builtin");
const theme_common = @import("common.zig");
const TerminalPalette = theme_common.TerminalPalette;
const types = @import("zed_types.zig");
const ThemeOverrides = @import("overrides.zig").ThemeOverrides;
const color_utils = @import("../color/utils.zig");
const GenerateOverridableRequest = theme_common.GenerateOverridableRequest;
const ThemeAppearance = theme_common.ThemeAppearance;

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
    appearance: ?ThemeAppearance,
) !ZedThemeResponse {
    const prepared = try theme_common.prepareThemeSelection(allocator, colors, overrides, appearance);
    const dark_base = prepared.dark_base;
    const has_background_override = overrides.background != null;
    const has_foreground_override = overrides.foreground != null;

    const background = if (overrides.background) |bg| bg else blk: {
        const bg_raw = prepared.background_seed;
        const base_luminance = color_utils.getLuminance(bg_raw);
        const darken_amount = if (dark_base) 0.5 + (base_luminance) * 1 else 0.0;
        const lighten_amount = if (dark_base) 0.0 else 0.2 + (1.0 - base_luminance) * 0.6;
        break :blk if (dark_base) color_utils.darkenColor(bg_raw, darken_amount) else color_utils.lightenColor(bg_raw, lighten_amount);
    };
    const bg_dark = if (dark_base) color_utils.darkenColor(background, 0.20) else color_utils.lightenColor(background, 0.05);
    const bg_very_dark = if (dark_base) color_utils.darkenColor(background, 0.30) else color_utils.lightenColor(background, 0.15);
    const bg_light = if (dark_base) color_utils.lightenColor(background, 0.10) else color_utils.darkenColor(background, 0.15);
    const bg_lighter = if (dark_base) color_utils.lightenColor(background, 0.20) else color_utils.darkenColor(background, 0.25);

    const proposed_foreground = overrides.foreground orelse prepared.foreground_seed;
    const foreground = if (has_foreground_override) proposed_foreground else color_utils.ensureReadableContrast(proposed_foreground, background, 7.0);

    const fg_muted = if (dark_base) color_utils.darkenColor(foreground, 0.30) else color_utils.lightenColor(foreground, 0.15);
    const fg_disabled = if (dark_base) color_utils.darkenColor(foreground, 0.40) else color_utils.lightenColor(foreground, 0.30);
    const fg_placeholder = if (dark_base) color_utils.darkenColor(foreground, 0.50) else color_utils.lightenColor(foreground, 0.40);

    const fg_12 = color_utils.addAlpha(foreground, "12");
    const fg_26 = color_utils.addAlpha(foreground, "26");
    const fg_40 = color_utils.addAlpha(foreground, "40");
    const fg_66 = color_utils.addAlpha(foreground, "66");
    const fg_80 = color_utils.addAlpha(foreground, "80");
    const fg_placeholder_30 = color_utils.addAlpha(fg_placeholder, "30");

    const c1 = theme_common.resolveAccent(prepared.c1_raw, background);
    const c2 = theme_common.resolveAccent(prepared.c2_raw, background);
    const c3 = theme_common.resolveAccent(prepared.c3_raw, background);
    const c4 = theme_common.resolveAccent(prepared.c4_raw, background);
    const c5 = theme_common.resolveAccent(prepared.c5_raw, background);
    const c6 = theme_common.resolveAccent(prepared.c6_raw, background);
    const c7 = theme_common.resolveAccent(prepared.c7_raw, background);
    const c8 = theme_common.resolveAccent(prepared.c8_raw, background);
    const c9 = theme_common.resolveAccent(prepared.c9_raw, background);

    const vim_normal_foreground = if (color_utils.contrastRatio(foreground, c2) >= color_utils.contrastRatio(bg_very_dark, c2)) foreground else bg_very_dark;
    const vim_visual_foreground = if (color_utils.contrastRatio(foreground, c6) >= color_utils.contrastRatio(bg_very_dark, c6)) foreground else bg_very_dark;
    const vim_insert_foreground = if (color_utils.contrastRatio(foreground, c4) >= color_utils.contrastRatio(bg_very_dark, c4)) foreground else bg_very_dark;
    const vim_visual_block_foreground = if (color_utils.contrastRatio(foreground, c3) >= color_utils.contrastRatio(bg_very_dark, c3)) foreground else bg_very_dark;
    const vim_replace_foreground = if (color_utils.contrastRatio(foreground, c8) >= color_utils.contrastRatio(bg_very_dark, c8)) foreground else bg_very_dark;

    const constants_raw = overrides.constants orelse color_utils.getHarmonicColor(c2, .@"split-complementary");
    const constants = theme_common.resolveAccent(constants_raw, background);

    const semantic_error = color_utils.boostAccentColor(color_utils.adjustForContrast(prepared.semantic.error_color, background, 3), background);
    const semantic_warning = color_utils.boostAccentColor(color_utils.adjustForContrast(prepared.semantic.warning_color, background, 3), background);
    const semantic_success = color_utils.boostAccentColor(color_utils.adjustForContrast(prepared.semantic.success_color, background, 3), background);
    const semantic_info = color_utils.boostAccentColor(color_utils.adjustForContrast(prepared.semantic.info_color, background, 3), background);

    const c2_33 = color_utils.addAlpha(c2, "33");
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

    const bright_semantic_warning = if (dark_base) color_utils.lightenColor(semantic_warning, 0.33) else color_utils.darkenColor(semantic_warning, 0.1);
    const accent_bright = if (dark_base) color_utils.lightenColor(c2, 0.33) else color_utils.darkenColor(c2, 0.33);

    const bg_lighter_4d = color_utils.addAlpha(bg_lighter, "4d");
    const bg_tab_inactive = if (dark_base) color_utils.darkenColor(bg_very_dark, 0.30) else color_utils.lightenColor(bg_very_dark, 0.30);
    const terminal: TerminalPalette = theme_common.buildTerminalPalette(
        background,
        foreground,
        semantic_error,
        semantic_success,
        semantic_warning,
        semantic_info,
        c2,
        c6,
        dark_base,
    );

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
    players[0] = .{ .cursor = c2, .selection = if (dark_base) color_utils.addAlpha(c2, "60") else color_utils.addAlpha(c2, "20"), .background = c2 };
    players[1] = .{ .cursor = c3, .selection = if (dark_base) color_utils.addAlpha(c3, "60") else color_utils.addAlpha(c3, "40"), .background = c3 };
    players[2] = .{ .cursor = c4, .selection = if (dark_base) color_utils.addAlpha(c4, "60") else color_utils.addAlpha(c4, "40"), .background = c4 };
    players[3] = .{ .cursor = c5, .selection = if (dark_base) color_utils.addAlpha(c5, "60") else color_utils.addAlpha(c5, "40"), .background = c5 };
    players[4] = .{ .cursor = c6, .selection = if (dark_base) color_utils.addAlpha(c6, "60") else color_utils.addAlpha(c6, "40"), .background = c6 };
    players[5] = .{ .cursor = c7, .selection = if (dark_base) color_utils.addAlpha(c7, "60") else color_utils.addAlpha(c7, "40"), .background = c7 };
    players[6] = .{ .cursor = c8, .selection = if (dark_base) color_utils.addAlpha(c8, "60") else color_utils.addAlpha(c8, "40"), .background = c8 };
    players[7] = .{ .cursor = c9, .selection = if (dark_base) color_utils.addAlpha(c9, "60") else color_utils.addAlpha(c9, "40"), .background = c9 };

    const style = ZedThemeStyle{
        .accents = accents,

        .@"vim.mode.text" = vim_normal_foreground,
        .@"vim.normal.foreground" = vim_normal_foreground,
        .@"vim.helix_normal.foreground" = vim_normal_foreground,
        .@"vim.visual.foreground" = vim_visual_foreground,
        .@"vim.helix_select.foreground" = vim_visual_foreground,
        .@"vim.insert.foreground" = vim_insert_foreground,
        .@"vim.visual_line.foreground" = vim_visual_foreground,
        .@"vim.visual_block.foreground" = vim_visual_block_foreground,
        .@"vim.replace.foreground" = vim_replace_foreground,

        .@"vim.normal.background" = c2,
        .@"vim.helix_normal.background" = c2,
        .@"vim.visual.background" = c6,
        .@"vim.helix_select.background" = c6,
        .@"vim.insert.background" = c4,
        .@"vim.visual_line.background" = c6,
        .@"vim.visual_block.background" = c3,
        .@"vim.replace.background" = c8,

        .@"background.appearance" = .@"opaque",

        .border = bg_lighter,
        .@"border.variant" = c2_88,
        .@"border.focused" = c2_88,
        .@"border.selected" = c2_88,
        .@"border.transparent" = c2_88,
        .@"border.disabled" = fg_disabled,

        .@"elevated_surface.background" = bg_dark,
        .@"surface.background" = bg_dark,
        .background = background,

        .@"element.background" = bg_very_dark,
        .@"element.hover" = bg_lighter,
        .@"element.active" = bg_lighter_4d,
        .@"element.selected" = bg_lighter_4d,
        .@"element.disabled" = fg_disabled,
        .@"drop_target.background" = color_utils.addAlpha(bg_lighter, "66"),

        .@"ghost_element.background" = "#00000000",
        .@"ghost_element.hover" = bg_light,
        .@"ghost_element.active" = bg_lighter,
        .@"ghost_element.selected" = c2_88,
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
        .@"title_bar.inactive_background" = bg_tab_inactive,
        .@"toolbar.background" = background,

        .@"tab_bar.background" = bg_very_dark,
        .@"tab.inactive_background" = bg_tab_inactive,
        .@"tab.active_background" = background,

        .@"search.match_background" = c3_33,

        .@"panel.background" = bg_dark,
        .@"panel.focused_border" = fg_40,
        .@"panel.indent_guide" = fg_placeholder_30,
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
        .@"editor.line_number" = fg_disabled,
        .@"editor.active_line_number" = c2,
        .@"editor.invisible" = fg_66,
        .@"editor.wrap_guide" = fg_placeholder,
        .@"editor.active_wrap_guide" = fg_placeholder,
        .@"editor.document_highlight.bracket_background" = color_utils.addAlpha(c2, "17"),
        .@"editor.document_highlight.read_background" = fg_26,
        .@"editor.document_highlight.write_background" = fg_26,
        .@"editor.indent_guide" = fg_placeholder_30,
        .@"editor.indent_guide_active" = fg_placeholder,

        .@"terminal.background" = background,
        .@"terminal.ansi.background" = background,
        .@"terminal.foreground" = terminal.foreground,
        .@"terminal.dim_foreground" = terminal.dim_foreground,
        .@"terminal.bright_foreground" = terminal.bright_foreground,
        .@"terminal.ansi.black" = terminal.ansi_black,
        .@"terminal.ansi.white" = terminal.ansi_white,
        .@"terminal.ansi.red" = terminal.ansi_red,
        .@"terminal.ansi.green" = terminal.ansi_green,
        .@"terminal.ansi.yellow" = terminal.ansi_yellow,
        .@"terminal.ansi.blue" = terminal.ansi_blue,
        .@"terminal.ansi.magenta" = terminal.ansi_magenta,
        .@"terminal.ansi.cyan" = terminal.ansi_cyan,
        .@"terminal.ansi.bright_black" = terminal.ansi_bright_black,
        .@"terminal.ansi.bright_white" = terminal.ansi_bright_white,
        .@"terminal.ansi.bright_red" = terminal.ansi_bright_red,
        .@"terminal.ansi.bright_green" = terminal.ansi_bright_green,
        .@"terminal.ansi.bright_yellow" = terminal.ansi_bright_yellow,
        .@"terminal.ansi.bright_blue" = terminal.ansi_bright_blue,
        .@"terminal.ansi.bright_magenta" = terminal.ansi_bright_magenta,
        .@"terminal.ansi.bright_cyan" = terminal.ansi_bright_cyan,
        .@"terminal.ansi.dim_black" = terminal.ansi_dim_black,
        .@"terminal.ansi.dim_white" = terminal.ansi_dim_white,
        .@"terminal.ansi.dim_red" = terminal.ansi_dim_red,
        .@"terminal.ansi.dim_green" = terminal.ansi_dim_green,
        .@"terminal.ansi.dim_yellow" = terminal.ansi_dim_yellow,
        .@"terminal.ansi.dim_blue" = terminal.ansi_dim_blue,
        .@"terminal.ansi.dim_magenta" = terminal.ansi_dim_magenta,
        .@"terminal.ansi.dim_cyan" = terminal.ansi_dim_cyan,

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
        .modified = bright_semantic_warning,
        .@"modified.border" = bright_semantic_warning,
        .@"modified.background" = if (dark_base) color_utils.lightenColor(semantic_warning_26, 0.33) else color_utils.darkenColor(semantic_warning_26, 0.2),
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
        .@"version_control.modified" = bright_semantic_warning,
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
            .comment = .{ .color = fg_disabled, .font_style = .italic, .font_weight = null },
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
        .background = if (has_background_override) overrides.background else background,
        .foreground = if (has_foreground_override) overrides.foreground else foreground,
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

    const raw_overrides = ThemeOverrides{
        .background = overrides.background,
        .foreground = overrides.foreground,
        .c1 = prepared.c1_raw,
        .c2 = prepared.c2_raw,
        .c3 = prepared.c3_raw,
        .c4 = prepared.c4_raw,
        .c5 = prepared.c5_raw,
        .c6 = prepared.c6_raw,
        .c7 = prepared.c7_raw,
        .c8 = prepared.c8_raw,
        .c9 = prepared.c9_raw,
        .constants = constants_raw,
    };

    return ZedThemeResponse{ .theme = theme, .themeOverrides = base_overrides, .rawThemeOverrides = raw_overrides, .colors = backupColors };
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

    return try generateZedTheme(allocator, palette, theme_name, overrides, request.appearance);
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
        .appearance = .light,
    };
    const response = try generateOverridableFromZedThemeValue(allocator, request);

    try std.testing.expect(response.themeOverrides.background != null);
    try std.testing.expect(response.themeOverrides.foreground != null);
    try std.testing.expect(response.themeOverrides.c1 != null);
    try std.testing.expectEqualStrings("#010203", response.rawThemeOverrides.c1.?);
    try std.testing.expect(!std.mem.eql(u8, response.themeOverrides.c1.?, response.rawThemeOverrides.c1.?));
}

test "generateOverridableFromZedThemeValue preserves explicit override exactly" {
    var arena = std.heap.ArenaAllocator.init(std.testing.allocator);
    defer arena.deinit();
    const allocator = arena.allocator();

    const json =
        "{" ++
        "\"name\":\"ExactOverride\"," ++
        "\"themes\":[{" ++
        "\"name\":\"ExactOverride\"," ++
        "\"appearance\":\"dark\"," ++
        "\"style\":{\"background\":\"#111111\",\"text\":\"#EEEEEE\",\"accents\":[\"#CC99FF\"]}" ++
        "}]}";

    const parsed = try std.json.parseFromSlice(std.json.Value, allocator, json, .{});
    defer parsed.deinit();

    const request = GenerateOverridableRequest{
        .theme = parsed.value,
        .ThemeOverrides = ThemeOverrides{ .c6 = "#123456" },
    };
    const response = try generateOverridableFromZedThemeValue(allocator, request);

    try std.testing.expectEqualStrings("#123456", response.rawThemeOverrides.c6.?);
    try std.testing.expect(!std.mem.eql(u8, response.themeOverrides.c6.?, response.rawThemeOverrides.c6.?));
}

test "generateZedTheme preserves exact manual overrides" {
    var arena = std.heap.ArenaAllocator.init(std.testing.allocator);
    defer arena.deinit();
    const allocator = arena.allocator();

    const palette = [_][]const u8{ "#112233", "#445566", "#778899", "#AA5500", "#00AA55", "#5500AA", "#CC8844", "#44CC88", "#8844CC", "#EEDD88", "#88DDEE" };
    const response = try generateZedTheme(allocator, &palette, "Manual Overrides", ThemeOverrides{
        .background = "#101010",
        .foreground = "#F0F0F0",
        .c6 = "#123456",
        .constants = "#ABCDEF",
    }, .dark);

    try std.testing.expectEqualStrings("#101010", response.themeOverrides.background.?);
    try std.testing.expectEqualStrings("#F0F0F0", response.themeOverrides.foreground.?);
    try std.testing.expectEqualStrings("#123456", response.rawThemeOverrides.c6.?);
    try std.testing.expectEqualStrings("#ABCDEF", response.rawThemeOverrides.constants.?);
    try std.testing.expect(!std.mem.eql(u8, response.themeOverrides.c6.?, response.rawThemeOverrides.c6.?));
    try std.testing.expectEqualStrings("#ABCDEF", response.themeOverrides.constants.?);
}
