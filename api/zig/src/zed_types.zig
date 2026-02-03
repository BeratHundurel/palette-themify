const ThemeOverrides = @import("theme_overrides.zig").ThemeOverrides;

pub const ZedThemeResponse = struct {
    themeOverrides: ThemeOverrides,
    theme: ZedTheme,
};

pub const FontStyle = enum {
    normal,
    italic,
    oblique,
};

pub const SyntaxStyle = struct {
    color: []const u8,
    font_style: ?FontStyle = null,
    font_weight: ?u16 = null,
};

pub const Player = struct {
    cursor: []const u8,
    selection: []const u8,
    background: []const u8,
};

pub const Appearance = enum {
    dark,
    light,
};

pub const BackgroundAppearance = enum {
    @"opaque",
    blurred,
    transparent,
};

pub const ZedThemeStyle = struct {
    // Accents
    accents: []const []const u8,

    // Vim mode colors
    @"vim.mode.text": []const u8,
    @"vim.normal.background": []const u8,
    @"vim.helix_normal.background": []const u8,
    @"vim.visual.background": []const u8,
    @"vim.helix_select.background": []const u8,
    @"vim.insert.background": []const u8,
    @"vim.visual_line.background": []const u8,
    @"vim.visual_block.background": []const u8,
    @"vim.replace.background": []const u8,

    // Background appearance
    @"background.appearance": BackgroundAppearance,

    // Borders
    border: []const u8,
    @"border.variant": []const u8,
    @"border.focused": []const u8,
    @"border.selected": []const u8,
    @"border.transparent": []const u8,
    @"border.disabled": []const u8,

    // Surfaces and backgrounds
    @"elevated_surface.background": []const u8,
    @"surface.background": []const u8,
    background: []const u8,

    // Elements
    @"element.background": []const u8,
    @"element.hover": []const u8,
    @"element.active": []const u8,
    @"element.selected": []const u8,
    @"element.disabled": []const u8,
    @"drop_target.background": []const u8,

    // Ghost elements
    @"ghost_element.background": []const u8,
    @"ghost_element.hover": []const u8,
    @"ghost_element.active": []const u8,
    @"ghost_element.selected": []const u8,
    @"ghost_element.disabled": []const u8,

    // Text
    text: []const u8,
    @"text.muted": []const u8,
    @"text.placeholder": []const u8,
    @"text.disabled": []const u8,
    @"text.accent": []const u8,

    // Icons
    icon: []const u8,
    @"icon.muted": []const u8,
    @"icon.disabled": []const u8,
    @"icon.placeholder": []const u8,
    @"icon.accent": []const u8,

    // Status bar
    @"status_bar.background": []const u8,

    // Title bar
    @"title_bar.background": []const u8,
    @"title_bar.inactive_background": []const u8,

    // Toolbar
    @"toolbar.background": []const u8,

    // Tabs
    @"tab_bar.background": []const u8,
    @"tab.inactive_background": []const u8,
    @"tab.active_background": []const u8,

    // Search
    @"search.match_background": []const u8,

    // Panel
    @"panel.background": []const u8,
    @"panel.focused_border": []const u8,
    @"panel.indent_guide": []const u8,
    @"panel.indent_guide_active": []const u8,
    @"panel.indent_guide_hover": []const u8,
    @"panel.overlay_background": []const u8,

    // Pane
    @"pane.focused_border": []const u8,
    @"pane_group.border": []const u8,

    // Scrollbar
    @"scrollbar.thumb.background": []const u8,
    @"scrollbar.thumb.hover_background": []const u8,
    @"scrollbar.thumb.active_background": ?[]const u8,
    @"scrollbar.thumb.border": ?[]const u8,
    @"scrollbar.track.background": []const u8,
    @"scrollbar.track.border": []const u8,

    // Minimap
    @"minimap.thumb.background": []const u8,
    @"minimap.thumb.hover_background": []const u8,
    @"minimap.thumb.active_background": []const u8,
    @"minimap.thumb.border": ?[]const u8,

    // Editor
    @"editor.foreground": []const u8,
    @"editor.background": []const u8,
    @"editor.gutter.background": []const u8,
    @"editor.subheader.background": []const u8,
    @"editor.active_line.background": []const u8,
    @"editor.highlighted_line.background": ?[]const u8,
    @"editor.line_number": []const u8,
    @"editor.active_line_number": []const u8,
    @"editor.invisible": []const u8,
    @"editor.wrap_guide": []const u8,
    @"editor.active_wrap_guide": []const u8,
    @"editor.document_highlight.bracket_background": []const u8,
    @"editor.document_highlight.read_background": []const u8,
    @"editor.document_highlight.write_background": []const u8,
    @"editor.indent_guide": []const u8,
    @"editor.indent_guide_active": []const u8,

    // Terminal
    @"terminal.background": []const u8,
    @"terminal.ansi.background": []const u8,
    @"terminal.foreground": []const u8,
    @"terminal.dim_foreground": []const u8,
    @"terminal.bright_foreground": []const u8,
    @"terminal.ansi.black": []const u8,
    @"terminal.ansi.white": []const u8,
    @"terminal.ansi.red": []const u8,
    @"terminal.ansi.green": []const u8,
    @"terminal.ansi.yellow": []const u8,
    @"terminal.ansi.blue": []const u8,
    @"terminal.ansi.magenta": []const u8,
    @"terminal.ansi.cyan": []const u8,
    @"terminal.ansi.bright_black": []const u8,
    @"terminal.ansi.bright_white": []const u8,
    @"terminal.ansi.bright_red": []const u8,
    @"terminal.ansi.bright_green": []const u8,
    @"terminal.ansi.bright_yellow": []const u8,
    @"terminal.ansi.bright_blue": []const u8,
    @"terminal.ansi.bright_magenta": []const u8,
    @"terminal.ansi.bright_cyan": []const u8,
    @"terminal.ansi.dim_black": []const u8,
    @"terminal.ansi.dim_white": []const u8,
    @"terminal.ansi.dim_red": []const u8,
    @"terminal.ansi.dim_green": []const u8,
    @"terminal.ansi.dim_yellow": []const u8,
    @"terminal.ansi.dim_blue": []const u8,
    @"terminal.ansi.dim_magenta": []const u8,
    @"terminal.ansi.dim_cyan": []const u8,

    // Link
    @"link_text.hover": []const u8,

    // Status colors
    conflict: []const u8,
    @"conflict.border": []const u8,
    @"conflict.background": []const u8,
    created: []const u8,
    @"created.border": []const u8,
    @"created.background": []const u8,
    deleted: []const u8,
    @"deleted.border": []const u8,
    @"deleted.background": []const u8,
    hidden: []const u8,
    @"hidden.border": []const u8,
    @"hidden.background": []const u8,
    hint: []const u8,
    @"hint.border": []const u8,
    @"hint.background": []const u8,
    ignored: []const u8,
    @"ignored.border": []const u8,
    @"ignored.background": []const u8,
    modified: []const u8,
    @"modified.border": []const u8,
    @"modified.background": []const u8,
    predictive: []const u8,
    @"predictive.border": []const u8,
    @"predictive.background": []const u8,
    renamed: []const u8,
    @"renamed.border": []const u8,
    @"renamed.background": []const u8,
    info: []const u8,
    @"info.border": []const u8,
    @"info.background": []const u8,
    warning: []const u8,
    @"warning.border": []const u8,
    @"warning.background": []const u8,
    @"error": []const u8,
    @"error.border": []const u8,
    @"error.background": []const u8,
    success: []const u8,
    @"success.border": []const u8,
    @"success.background": []const u8,
    @"unreachable": []const u8,
    @"unreachable.border": []const u8,
    @"unreachable.background": []const u8,

    // Players (for collaborative editing)
    players: []const Player,

    // Version control
    @"version_control.added": []const u8,
    @"version_control.deleted": []const u8,
    @"version_control.modified": []const u8,
    @"version_control.renamed": []const u8,
    @"version_control.conflict": []const u8,
    @"version_control.conflict_marker.ours": []const u8,
    @"version_control.conflict_marker.theirs": []const u8,
    @"version_control.ignored": []const u8,

    // Debugger
    @"debugger.accent": []const u8,
    @"editor.debugger_active_line.background": []const u8,

    // Syntax highlighting
    syntax: ZedSyntax,
};

pub const ZedSyntax = struct {
    attribute: SyntaxStyle,
    boolean: SyntaxStyle,
    character: SyntaxStyle,
    @"character.special": SyntaxStyle,
    comment: SyntaxStyle,
    @"comment.doc": SyntaxStyle,
    @"comment.documentation": SyntaxStyle,
    @"comment.error": SyntaxStyle,
    @"comment.hint": SyntaxStyle,
    @"comment.note": SyntaxStyle,
    @"comment.todo": SyntaxStyle,
    @"comment.warning": SyntaxStyle,
    concept: SyntaxStyle,
    constant: SyntaxStyle,
    @"constant.macro": SyntaxStyle,
    constructor: SyntaxStyle,
    @"diff.minus": SyntaxStyle,
    @"diff.plus": SyntaxStyle,
    embedded: SyntaxStyle,
    emphasis: SyntaxStyle,
    @"emphasis.strong": SyntaxStyle,
    @"enum": SyntaxStyle,
    field: SyntaxStyle,
    float: SyntaxStyle,
    function: SyntaxStyle,
    @"function.decorator": SyntaxStyle,
    hint: SyntaxStyle,
    keyword: SyntaxStyle,
    @"keyword.directive": SyntaxStyle,
    @"keyword.export": SyntaxStyle,
    label: SyntaxStyle,
    link_text: SyntaxStyle,
    link_uri: SyntaxStyle,
    module: SyntaxStyle,
    namespace: SyntaxStyle,
    number: SyntaxStyle,
    @"number.float": SyntaxStyle,
    operator: SyntaxStyle,
    parameter: SyntaxStyle,
    parent: SyntaxStyle,
    predictive: SyntaxStyle,
    predoc: SyntaxStyle,
    preproc: SyntaxStyle,
    primary: SyntaxStyle,
    property: SyntaxStyle,
    punctuation: SyntaxStyle,
    @"punctuation.bracket": SyntaxStyle,
    @"punctuation.delimiter": SyntaxStyle,
    @"punctuation.list_marker": SyntaxStyle,
    @"punctuation.markup": SyntaxStyle,
    @"punctuation.special": SyntaxStyle,
    @"punctuation.special.symbol": SyntaxStyle,
    selector: SyntaxStyle,
    @"selector.pseudo": SyntaxStyle,
    string: SyntaxStyle,
    @"string.doc": SyntaxStyle,
    @"string.documentation": SyntaxStyle,
    @"string.escape": SyntaxStyle,
    @"string.regex": SyntaxStyle,
    @"string.regexp": SyntaxStyle,
    @"string.special": SyntaxStyle,
    @"string.special.path": SyntaxStyle,
    @"string.special.symbol": SyntaxStyle,
    @"string.special.url": SyntaxStyle,
    symbol: SyntaxStyle,
    tag: SyntaxStyle,
    @"tag.attribute": SyntaxStyle,
    @"tag.delimiter": SyntaxStyle,
    @"tag.doctype": SyntaxStyle,
    text: SyntaxStyle,
    @"text.literal": SyntaxStyle,
    title: SyntaxStyle,
    type: SyntaxStyle,
    @"type.class.definition": SyntaxStyle,
    variable: SyntaxStyle,
    @"variable.builtin": SyntaxStyle,
    @"variable.member": SyntaxStyle,
    @"variable.parameter": SyntaxStyle,
    @"variable.special": SyntaxStyle,
    variant: SyntaxStyle,
};

pub const ZedThemeEntry = struct {
    name: []const u8,
    appearance: Appearance,
    style: ZedThemeStyle,
};

pub const ZedTheme = struct {
    @"$schema": []const u8,
    name: []const u8,
    author: []const u8,
    themes: []const ZedThemeEntry,
};
