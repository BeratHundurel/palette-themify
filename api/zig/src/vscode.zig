const std = @import("std");

const color_utils = @import("color_utils.zig");
const types = @import("vscode_types.zig");
const ThemeOverrides = @import("theme_overrides.zig").ThemeOverrides;
const VSCodeTheme = types.VSCodeTheme;
const VSCodeThemeColors = types.VSCodeThemeColors;
const VSCodeTokenColor = types.VSCodeTokenColor;
const VSCodeThemeResponse = types.VSCodeThemeResponse;

/// Generates a complete VS Code theme from a palette of colors.
/// Strategy: Select 10 most diverse colors, pick bg/fg with good contrast,
/// then assign remaining colors to syntax tokens and UI elements.
pub fn generateVSCodeTheme(
    allocator: std.mem.Allocator,
    colors: []const []const u8,
    theme_name: []const u8,
    overrides: ThemeOverrides,
) !VSCodeThemeResponse {
    const semantic = color_utils.findSemanticColors(colors);
    const improved_colors = try color_utils.selectDiverseColors(allocator, colors, 10);
    defer allocator.free(improved_colors);

    var palette = std.ArrayList([]const u8){};
    defer palette.deinit(allocator);
    try palette.ensureTotalCapacity(allocator, 10);
    try palette.appendSlice(allocator, improved_colors);

    if (palette.items.len == 0) {
        return error.NotEnoughColors;
    }

    if (palette.items.len < 10) {
        const harmony_schemes = [_]color_utils.HarmonyScheme{
            .complementary,
            .triadic,
            .analogous,
            .@"split-complementary",
        };
        const base_count = palette.items.len;
        var scheme_index: usize = 0;
        var base_index: usize = 0;

        while (palette.items.len < 10) {
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
    const dark_base = average_luminance < 128.0; // Luminance range is 0-1, but comparison seems off - effectively always dark

    const selection = try color_utils.selectBackgroundAndForeground(allocator, palette.items, dark_base);
    defer allocator.free(selection.remaining_indices);

    const c0 = overrides.background orelse palette.items[selection.background_index];
    const remaining = selection.remaining_indices;
    const c1_raw = overrides.c1 orelse palette.items[remaining[0]];
    const c2_raw = overrides.c2 orelse palette.items[remaining[1]];
    const c3_raw = overrides.c3 orelse palette.items[remaining[2]];
    const c4_raw = overrides.c4 orelse palette.items[remaining[3]];
    const c5_raw = overrides.c5 orelse palette.items[remaining[4]];
    const c6_raw = overrides.c6 orelse palette.items[remaining[5]];
    const c7_raw = overrides.c7 orelse palette.items[remaining[6]];
    const c8_raw = overrides.c8 orelse palette.items[remaining[7]];

    const base_luminance = color_utils.getLuminance(c0);
    const darken_amount = if (dark_base) 0.75 + (base_luminance) * 0.20 else 0.0;
    const lighten_amount = if (dark_base) 0.0 else 0.75 + (1.0 - base_luminance) * 0.20;
    const background = if (dark_base) color_utils.darkenColor(c0, darken_amount) else color_utils.lightenColor(c0, lighten_amount);

    const proposed_foreground = overrides.foreground orelse palette.items[selection.foreground_index];
    const foreground = color_utils.ensureReadableContrast(proposed_foreground, background, 7.0);

    const bg_very_dark = if (dark_base) color_utils.darkenColor(background, 0.20) else color_utils.lightenColor(background, 0.20);
    const bg_dark = if (dark_base) color_utils.darkenColor(background, 0.15) else color_utils.lightenColor(background, 0.15);
    const bg_medium = if (dark_base) color_utils.darkenColor(background, 0.10) else color_utils.lightenColor(background, 0.10);
    const bg_light = if (dark_base) color_utils.lightenColor(background, 0.05) else color_utils.darkenColor(background, 0.05);
    const bg_lighter = if (dark_base) color_utils.lightenColor(background, 0.10) else color_utils.darkenColor(background, 0.10);
    const bg_inactive = if (dark_base) color_utils.darkenColor(background, 0.50) else color_utils.lightenColor(background, 0.50);

    // These are more often used against very dark backgrounds, so adjust accordingly
    const c1 = color_utils.adjustForContrast(c1_raw, bg_very_dark, 3.5);
    const c2 = color_utils.adjustForContrast(c2_raw, bg_very_dark, 3.5);

    const numbers_raw = color_utils.getHarmonicColor(c2, .complementary);
    const numbers = color_utils.adjustForContrast(numbers_raw, background, 3.5);

    const c3 = color_utils.adjustForContrast(c3_raw, background, 3.5);
    const c4 = color_utils.adjustForContrast(c4_raw, background, 3.5);
    const c5 = color_utils.adjustForContrast(c5_raw, background, 3.5);
    const c6 = color_utils.adjustForContrast(c6_raw, background, 3.5);
    const c7 = color_utils.adjustForContrast(c7_raw, background, 3.5);
    const c8 = color_utils.adjustForContrast(c8_raw, background, 3.5);

    const semantic_error = color_utils.adjustForContrast(semantic.error_color, background, 3.5);
    const semantic_warning = color_utils.adjustForContrast(semantic.warning_color, background, 3.5);
    const semantic_success = color_utils.adjustForContrast(semantic.success_color, background, 3.5);
    const semantic_info = color_utils.adjustForContrast(semantic.info_color, background, 3.5);

    const c3_dark = if (dark_base) color_utils.darkenColor(c3, 0.8) else color_utils.lightenColor(c3, 0.8);
    const semantic_error_dark = if (dark_base) color_utils.darkenColor(semantic_error, 0.8) else color_utils.lightenColor(semantic_error, 0.8);
    const semantic_warning_dark = if (dark_base) color_utils.darkenColor(semantic_warning, 0.8) else color_utils.lightenColor(semantic_warning, 0.8);
    const semantic_info_dark = if (dark_base) color_utils.darkenColor(semantic_info, 0.8) else color_utils.lightenColor(semantic_info, 0.8);
    const button_fg = if (dark_base) background else color_utils.darkenColor(c0, 0.9);

    const fg30 = color_utils.addAlpha(foreground, "30");
    const fg50 = color_utils.addAlpha(foreground, "50");
    const fg60 = color_utils.addAlpha(foreground, "60");
    const fg70 = color_utils.addAlpha(foreground, "70");
    const fg99 = color_utils.addAlpha(foreground, "99");
    const c1_20 = color_utils.addAlpha(c1, "20");
    const c1_30 = color_utils.addAlpha(c1, "30");
    const c1_40 = color_utils.addAlpha(c1, "40");
    const c1_60 = color_utils.addAlpha(c1, "60");
    const c1_80 = color_utils.addAlpha(c1, "80");
    const c2_20 = color_utils.addAlpha(c2, "20");
    const c2_30 = color_utils.addAlpha(c2, "30");
    const c2_40 = color_utils.addAlpha(c2, "40");
    const c2_50 = color_utils.addAlpha(c2, "50");
    const c2_60 = color_utils.addAlpha(c2, "60");
    const c2_80 = color_utils.addAlpha(c2, "80");
    const c3_30 = color_utils.addAlpha(c3, "30");
    const c3_80 = color_utils.addAlpha(c3, "80");
    const c4_80 = color_utils.addAlpha(c4, "80");
    const c5_20 = color_utils.addAlpha(c5, "20");
    const c5_30 = color_utils.addAlpha(c5, "30");
    const c5_40 = color_utils.addAlpha(c5, "40");
    const c5_80 = color_utils.addAlpha(c5, "80");
    const c6_30 = color_utils.addAlpha(c6, "30");
    const c6_80 = color_utils.addAlpha(c6, "80");
    const c7_30 = color_utils.addAlpha(c7, "30");
    const c7_80 = color_utils.addAlpha(c7, "80");
    const fg_aa = color_utils.addAlpha(foreground, "aa");
    const fg_bracket_dark = color_utils.addAlpha(foreground, if (dark_base) "90" else "80");
    const fg_punct_dark = color_utils.addAlpha(foreground, if (dark_base) "70" else "60");

    const theme_colors = VSCodeThemeColors{
        .@"editor.background" = background,
        .@"editor.foreground" = foreground,
        .foreground = foreground,
        .disabledForeground = fg60,
        .focusBorder = c2_60,
        .descriptionForeground = fg70,
        .errorForeground = c4,
        .@"icon.foreground" = foreground,
        .@"widget.border" = c1_40,
        .@"selection.background" = c2_50,
        .@"sash.hoverBorder" = c2,
        .@"activityBar.background" = bg_very_dark,
        .@"activityBar.foreground" = foreground,
        .@"activityBar.activeBorder" = c2,
        .@"activityBarBadge.background" = c2,
        .@"activityBarBadge.foreground" = foreground,
        .@"sideBar.background" = bg_dark,
        .@"sideBar.foreground" = foreground,
        .@"sideBar.border" = c1_20,
        .@"sideBarTitle.foreground" = foreground,
        .@"statusBar.background" = bg_very_dark,
        .@"statusBar.foreground" = foreground,
        .@"statusBar.noFolderBackground" = c3_dark,
        .@"statusBar.debuggingBackground" = c4,
        .@"titleBar.activeBackground" = bg_very_dark,
        .@"titleBar.activeForeground" = foreground,
        .@"titleBar.inactiveBackground" = bg_inactive,
        .@"titleBar.inactiveForeground" = fg99,
        .@"tab.activeBackground" = background,
        .@"tab.activeForeground" = foreground,
        .@"tab.inactiveBackground" = bg_very_dark,
        .@"tab.inactiveForeground" = fg_aa,
        .@"tab.activeBorder" = c2,
        .@"tab.border" = c1_20,
        .@"editorGroupHeader.tabsBackground" = bg_very_dark,
        .@"panel.background" = background,
        .@"panel.border" = c1_40,
        .@"panelTitle.activeBorder" = c2,
        .@"terminal.foreground" = foreground,
        .@"terminal.ansiBlack" = if (dark_base) color_utils.darkenColor(c0, 0.9) else color_utils.darkenColor(c0, 0.2),
        .@"terminal.ansiRed" = semantic_error,
        .@"terminal.ansiGreen" = semantic_success,
        .@"terminal.ansiYellow" = semantic_warning,
        .@"terminal.ansiBlue" = semantic_info,
        .@"terminal.ansiMagenta" = c6,
        .@"terminal.ansiCyan" = c7,
        .@"terminal.ansiWhite" = foreground,
        .@"terminal.ansiBrightBlack" = if (dark_base) color_utils.darkenColor(foreground, 0.3) else color_utils.lightenColor(foreground, 0.3),
        .@"terminal.ansiBrightRed" = if (dark_base) color_utils.lightenColor(semantic_error, 0.2) else color_utils.darkenColor(semantic_error, 0.2),
        .@"terminal.ansiBrightGreen" = if (dark_base) color_utils.lightenColor(semantic_success, 0.2) else color_utils.darkenColor(semantic_success, 0.2),
        .@"terminal.ansiBrightYellow" = if (dark_base) color_utils.lightenColor(semantic_warning, 0.2) else color_utils.darkenColor(semantic_warning, 0.2),
        .@"terminal.ansiBrightBlue" = if (dark_base) color_utils.lightenColor(semantic_info, 0.2) else color_utils.darkenColor(semantic_info, 0.2),
        .@"terminal.ansiBrightMagenta" = if (dark_base) color_utils.lightenColor(c6, 0.2) else color_utils.darkenColor(c6, 0.2),
        .@"terminal.ansiBrightCyan" = if (dark_base) color_utils.lightenColor(c7, 0.2) else color_utils.darkenColor(c7, 0.2),
        .@"terminal.ansiBrightWhite" = if (dark_base) color_utils.lightenColor(foreground, 0.2) else color_utils.darkenColor(foreground, 0.2),
        .@"input.background" = bg_lighter,
        .@"input.border" = c1_40,
        .@"input.foreground" = foreground,
        .@"input.placeholderForeground" = fg50,
        .@"inputOption.activeBorder" = c2,
        .@"inputOption.activeBackground" = c2_30,
        .@"inputOption.activeForeground" = foreground,
        .@"inputValidation.errorBackground" = semantic_error_dark,
        .@"inputValidation.errorBorder" = semantic_error,
        .@"inputValidation.errorForeground" = foreground,
        .@"inputValidation.warningBackground" = semantic_warning_dark,
        .@"inputValidation.warningBorder" = semantic_warning,
        .@"inputValidation.warningForeground" = foreground,
        .@"inputValidation.infoBackground" = semantic_info_dark,
        .@"inputValidation.infoBorder" = semantic_info,
        .@"inputValidation.infoForeground" = foreground,
        .@"dropdown.background" = bg_light,
        .@"dropdown.foreground" = foreground,
        .@"dropdown.border" = c1_40,
        .@"dropdown.listBackground" = bg_lighter,
        .@"quickInput.background" = bg_light,
        .@"quickInput.foreground" = foreground,
        .@"quickInputList.focusBackground" = c2_40,
        .@"quickInputList.focusForeground" = foreground,
        .@"quickInputList.focusIconForeground" = c2,
        .@"quickInputTitle.background" = bg_dark,
        .@"list.activeSelectionBackground" = c2_40,
        .@"list.activeSelectionForeground" = foreground,
        .@"list.inactiveSelectionBackground" = c1_30,
        .@"list.hoverBackground" = c1_20,
        .@"list.focusBackground" = c2_30,
        .@"list.highlightForeground" = c2,
        .@"pickerGroup.foreground" = c6,
        .@"pickerGroup.border" = c1_60,
        .@"button.background" = c2,
        .@"button.foreground" = button_fg,
        .@"button.hoverBackground" = if (dark_base) color_utils.lightenColor(c2, 0.1) else color_utils.darkenColor(c2, 0.1),
        .@"button.secondaryBackground" = bg_light,
        .@"button.secondaryForeground" = foreground,
        .@"button.secondaryHoverBackground" = bg_lighter,
        .@"badge.background" = c2,
        .@"badge.foreground" = button_fg,
        .@"breadcrumb.foreground" = fg70,
        .@"breadcrumb.focusForeground" = foreground,
        .@"breadcrumb.activeSelectionForeground" = c2,
        .@"breadcrumb.background" = background,
        .@"scrollbarSlider.background" = c1_40,
        .@"scrollbarSlider.hoverBackground" = c1_60,
        .@"scrollbarSlider.activeBackground" = c2_60,
        .@"editorLineNumber.foreground" = fg50,
        .@"editorLineNumber.activeForeground" = c2,
        .@"editorCursor.foreground" = c2,
        .@"editor.selectionBackground" = c2_40,
        .@"editor.inactiveSelectionBackground" = c1_30,
        .@"editor.findMatchBackground" = c5_40,
        .@"editor.findMatchHighlightBackground" = c5_20,
        .@"editorBracketMatch.background" = c2_20,
        .@"editorBracketMatch.border" = c2,
        .@"editorBracketHighlight.foreground1" = c2_80,
        .@"editorBracketHighlight.foreground2" = c3_80,
        .@"editorBracketHighlight.foreground3" = c5_80,
        .@"editorBracketHighlight.foreground4" = c6_80,
        .@"editorBracketHighlight.foreground5" = c7_80,
        .@"editorBracketHighlight.foreground6" = c1_80,
        .@"editorBracketPairGuide.activeBackground1" = c2,
        .@"editorBracketPairGuide.activeBackground2" = c3,
        .@"editorBracketPairGuide.activeBackground3" = c5,
        .@"editorBracketPairGuide.activeBackground4" = c6,
        .@"editorBracketPairGuide.activeBackground5" = c7,
        .@"editorBracketPairGuide.activeBackground6" = c1,
        .@"editorBracketPairGuide.background1" = c2_30,
        .@"editorBracketPairGuide.background2" = c3_30,
        .@"editorBracketPairGuide.background3" = c5_30,
        .@"editorBracketPairGuide.background4" = c6_30,
        .@"editorBracketPairGuide.background5" = c7_30,
        .@"editorBracketPairGuide.background6" = c1_30,
        .@"editorWhitespace.foreground" = fg30,
        .@"editorWidget.background" = bg_light,
        .@"editorWidget.foreground" = foreground,
        .@"editorWidget.border" = c1_40,
        .@"editorWidget.resizeBorder" = c2,
        .@"editorSuggestWidget.background" = bg_light,
        .@"editorSuggestWidget.foreground" = foreground,
        .@"editorSuggestWidget.border" = c1_40,
        .@"editorSuggestWidget.highlightForeground" = c2,
        .@"editorSuggestWidget.focusHighlightForeground" = c2,
        .@"editorSuggestWidget.selectedBackground" = c2_40,
        .@"editorSuggestWidget.selectedForeground" = foreground,
        .@"editorSuggestWidget.selectedIconForeground" = c2,
        .@"editorHoverWidget.background" = bg_light,
        .@"editorHoverWidget.foreground" = foreground,
        .@"editorHoverWidget.border" = c1_40,
        .@"editorHoverWidget.highlightForeground" = c2,
        .@"editorHoverWidget.statusBarBackground" = bg_dark,
        .@"editorError.foreground" = semantic_error,
        .@"editorWarning.foreground" = semantic_warning,
        .@"editorInfo.foreground" = semantic_info,
        .@"editorGutter.addedBackground" = semantic_success,
        .@"editorGutter.modifiedBackground" = semantic_warning,
        .@"editorGutter.deletedBackground" = semantic_error,
        .@"gitDecoration.addedResourceForeground" = semantic_success,
        .@"gitDecoration.modifiedResourceForeground" = semantic_warning,
        .@"gitDecoration.deletedResourceForeground" = semantic_error,
        .@"gitDecoration.untrackedResourceForeground" = c7,
        .@"gitDecoration.ignoredResourceForeground" = fg60,
        .@"peekView.border" = c2,
        .@"peekViewEditor.background" = bg_light,
        .@"peekViewResult.background" = bg_dark,
        .@"peekViewTitle.background" = bg_very_dark,
        .@"notificationCenter.border" = c1_40,
        .@"notificationCenterHeader.background" = bg_dark,
        .@"notifications.background" = bg_light,
        .@"notifications.border" = c1_40,
        .@"notificationLink.foreground" = c2,
        .@"settings.headerForeground" = foreground,
        .@"settings.modifiedItemIndicator" = c2,
        .@"settings.focusedRowBackground" = bg_medium,
        .@"settings.rowHoverBackground" = bg_dark,
        .@"settings.focusedRowBorder" = c2_60,
        .@"settings.numberInputBackground" = background,
        .@"settings.numberInputForeground" = c6,
        .@"settings.numberInputBorder" = c1_40,
        .@"settings.textInputBackground" = background,
        .@"settings.textInputForeground" = c2,
        .@"settings.textInputBorder" = c1_40,
        .@"settings.checkboxBackground" = background,
        .@"settings.checkboxForeground" = c5,
        .@"settings.checkboxBorder" = c1_40,
        .@"settings.dropdownBackground" = background,
        .@"settings.dropdownForeground" = foreground,
        .@"settings.dropdownBorder" = c1_40,
        .@"settings.dropdownListBorder" = c1_40,
        .@"textLink.foreground" = c2,
        .@"textLink.activeForeground" = if (dark_base) color_utils.lightenColor(c2, 0.15) else color_utils.darkenColor(c2, 0.15),
        .@"textBlockQuote.background" = bg_dark,
        .@"textBlockQuote.border" = c1_40,
        .@"textCodeBlock.background" = bg_dark,
        .@"textPreformat.foreground" = c5,
        .@"textSeparator.foreground" = fg50,
        .@"walkThrough.embeddedEditorBackground" = bg_very_dark,
        .@"welcomePage.background" = background,
    };

    var token_colors = [_]VSCodeTokenColor{
        .{
            .scope = &[_][]const u8{ "comment", "punctuation.definition.comment" },
            .settings = .{ .foreground = fg60, .fontStyle = "italic" },
        },
        .{
            .scope = &[_][]const u8{ "keyword", "keyword.control", "keyword.operator.new", "keyword.operator.expression", "keyword.other" },
            .settings = .{ .foreground = c6, .fontStyle = "bold" },
        },
        .{
            .scope = &[_][]const u8{ "storage", "storage.type", "storage.modifier", "entity.name.tag", "meta.tag" },
            .settings = .{ .foreground = c6 },
        },
        .{
            .scope = &[_][]const u8{ "string", "string.quoted", "string.template", "string.regexp", "punctuation.definition.string", "support.constant.property-value", "support.constant.property-value.css", "markup.inline.raw", "markup.fenced_code", "markup.inserted" },
            .settings = .{ .foreground = c3 },
        },
        .{
            .scope = &[_][]const u8{ "constant.numeric", "constant.character", "constant.language.boolean", "constant.language.null", "number" },
            .settings = .{ .foreground = numbers },
        },
        .{
            .scope = &[_][]const u8{ "constant.language", "constant.other", "entity.name.class", "entity.other.inherited-class", "entity.name.type", "variable.other.constant", "support.constant", "support.class", "support.type" },
            .settings = .{ .foreground = c5 },
        },
        .{
            .scope = &[_][]const u8{ "variable", "identifier", "variable.other.readwrite", "meta.definition.variable" },
            .settings = .{ .foreground = foreground },
        },
        .{
            .scope = &[_][]const u8{ "variable.other.property", "variable.other.object.property", "meta.object-literal.key", "support.variable", "support.other.variable", "support.type.property-name", "support.type.property-name.css" },
            .settings = .{ .foreground = c1 },
        },
        .{
            .scope = &[_][]const u8{ "entity.name.function", "meta.function-call", "meta.method-call", "meta.method", "entity.other.attribute-name", "entity.name.module", "support.module", "support.function", "support.node" },
            .settings = .{ .foreground = c2 },
        },
        .{
            .scope = &[_][]const u8{ "variable.parameter", "meta.parameter" },
            .settings = .{ .foreground = c7 },
        },
        .{
            .scope = &[_][]const u8{ "punctuation.definition.begin.bracket", "punctuation.definition.end.bracket", "punctuation.definition.begin.bracket.round", "punctuation.definition.end.bracket.round", "punctuation.definition.begin.bracket.square", "punctuation.definition.end.bracket.square", "punctuation.definition.begin.bracket.curly", "punctuation.definition.end.bracket.curly", "meta.brace", "punctuation.section.brackets", "punctuation.section.parens", "punctuation.section.braces" },
            .settings = .{ .foreground = fg_bracket_dark },
        },
        .{
            .scope = &[_][]const u8{ "punctuation", "punctuation.terminator", "punctuation.separator", "punctuation.separator.comma", "punctuation.definition" },
            .settings = .{ .foreground = fg_punct_dark },
        },
        .{
            .scope = &[_][]const u8{ "keyword.operator", "punctuation.operator" },
            .settings = .{ .foreground = c8 },
        },
        .{
            .scope = &[_][]const u8{ "markup.heading", "entity.name.section" },
            .settings = .{ .foreground = c2, .fontStyle = "bold" },
        },
        .{
            .scope = &[_][]const u8{"markup.italic"},
            .settings = .{ .fontStyle = "italic" },
        },
        .{
            .scope = &[_][]const u8{"markup.bold"},
            .settings = .{ .fontStyle = "bold" },
        },
        .{
            .scope = &[_][]const u8{ "markup.underline.link", "string.other.link" },
            .settings = .{ .foreground = c2, .fontStyle = "underline" },
        },
        .{
            .scope = &[_][]const u8{"markup.deleted"},
            .settings = .{ .foreground = c4 },
        },
        .{
            .scope = &[_][]const u8{ "invalid", "invalid.illegal" },
            .settings = .{ .foreground = c4, .fontStyle = "bold" },
        },
        .{
            .scope = &[_][]const u8{"invalid.deprecated"},
            .settings = .{ .foreground = c4_80, .fontStyle = "italic" },
        },
    };

    const token_colors_slice = try allocator.dupe(VSCodeTokenColor, &token_colors);

    const theme = VSCodeTheme{
        .@"$schema" = "vscode://schemas/color-theme",
        .name = theme_name,
        .type = if (dark_base) .dark else .light,
        .colors = theme_colors,
        .tokenColors = token_colors_slice,
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
    };

    return VSCodeThemeResponse{
        .theme = theme,
        .themeOverrides = base_overrides,
    };
}
