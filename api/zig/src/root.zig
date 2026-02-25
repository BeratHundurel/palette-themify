const std = @import("std");
pub const vscode = @import("vscode.zig");
pub const zed = @import("zed.zig");
pub const color_utils = @import("color_utils.zig");
pub const palette_api = @import("palette_api.zig");
test {
    _ = color_utils;
    _ = palette_api;
    _ = zed;
    _ = vscode;
}
