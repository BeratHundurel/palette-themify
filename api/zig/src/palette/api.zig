const std = @import("std");
const zigimg = @import("zigimg");
const color_utils = @import("../color/utils.zig");
const ThemeOverrides = @import("../theme/overrides.zig").ThemeOverrides;
const theme_common = @import("../theme/common.zig");
const vscode = @import("../theme/vscode.zig");
const VSCodeThemeResponse = @import("../theme/vscode_types.zig").VSCodeThemeResponse;
const zed = @import("../theme/zed.zig");
const ZedThemeResponse = @import("../theme/zed_types.zig").ZedThemeResponse;
const ZedTheme = @import("../theme/zed_types.zig").ZedTheme;

pub const ExtractError = error{
    NotAnImage,
    NoColors,
    OutOfMemory,
    InvalidImageData,
    InvalidContentType,
    NoImageProvided,
};

pub const ColorAndCount = struct {
    color: zigimg.color.Colorf32,
    count: usize,

    /// Two colors are considered similar if all RGB components differ by less than 10%
    fn colorSimilar(a: zigimg.color.Colorf32, b: zigimg.color.Colorf32) bool {
        const threshold = 0.1;
        return @abs(a.r - b.r) < threshold and
            @abs(a.g - b.g) < threshold and
            @abs(a.b - b.b) < threshold;
    }

    fn lessThan(_: void, a: ColorAndCount, b: ColorAndCount) bool {
        return a.count > b.count;
    }
};

const QuantizedColorAndCount = struct {
    r: u8,
    g: u8,
    b: u8,
    count: usize,

    fn similar(self: QuantizedColorAndCount, r: u8, g: u8, b: u8) bool {
        return absDiffU8(self.r, r) <= 25 and
            absDiffU8(self.g, g) <= 25 and
            absDiffU8(self.b, b) <= 25;
    }

    fn lessThan(_: void, a: QuantizedColorAndCount, b: QuantizedColorAndCount) bool {
        return a.count > b.count;
    }
};

inline fn absDiffU8(a: u8, b: u8) u8 {
    return if (a >= b) a - b else b - a;
}

pub const PaletteResult = struct {
    colors: [][]const u8,
    allocator: std.mem.Allocator,

    pub fn deinit(self: *PaletteResult) void {
        for (self.colors) |color| {
            self.allocator.free(color);
        }
        self.allocator.free(self.colors);
    }
};

pub fn extractPaletteFromBytes(
    allocator: std.mem.Allocator,
    image_data: []const u8,
) ExtractError!PaletteResult {
    var img = zigimg.Image.fromMemory(allocator, image_data) catch {
        return ExtractError.NotAnImage;
    };
    defer img.deinit(allocator);

    return processImage(allocator, &img);
}

fn processImage(
    allocator: std.mem.Allocator,
    img: *zigimg.Image,
) ExtractError!PaletteResult {
    return switch (img.pixels) {
        .rgb24 => |pixels| processRgb24Pixels(allocator, pixels),
        .rgba32 => |pixels| processRgba32Pixels(allocator, pixels),
        else => processImageGeneric(allocator, img),
    };
}

fn processRgb24Pixels(
    allocator: std.mem.Allocator,
    pixels: []const zigimg.color.Rgb24,
) ExtractError!PaletteResult {
    var color_map = std.ArrayList(QuantizedColorAndCount){};
    defer color_map.deinit(allocator);
    color_map.ensureTotalCapacity(allocator, 512) catch return ExtractError.OutOfMemory;

    for (pixels) |pixel| {
        var found = false;
        for (color_map.items) |*item| {
            if (item.similar(pixel.r, pixel.g, pixel.b)) {
                item.count += 1;
                found = true;
                break;
            }
        }

        if (!found) {
            color_map.append(allocator, .{
                .r = pixel.r,
                .g = pixel.g,
                .b = pixel.b,
                .count = 1,
            }) catch return ExtractError.OutOfMemory;
        }
    }

    if (color_map.items.len == 0) {
        return ExtractError.NoColors;
    }

    std.mem.sort(QuantizedColorAndCount, color_map.items, {}, QuantizedColorAndCount.lessThan);

    const num_colors = @min(20, color_map.items.len);
    const colors = allocator.alloc([]const u8, num_colors) catch return ExtractError.OutOfMemory;
    errdefer {
        for (colors) |c| {
            allocator.free(c);
        }
        allocator.free(colors);
    }

    for (0..num_colors) |i| {
        const c = color_map.items[i];
        colors[i] = color_utils.rgbToHexAlloc(allocator, c.r, c.g, c.b) catch return ExtractError.OutOfMemory;
    }

    return PaletteResult{ .colors = colors, .allocator = allocator };
}

fn processRgba32Pixels(
    allocator: std.mem.Allocator,
    pixels: []const zigimg.color.Rgba32,
) ExtractError!PaletteResult {
    var color_map = std.ArrayList(QuantizedColorAndCount){};
    defer color_map.deinit(allocator);
    color_map.ensureTotalCapacity(allocator, 512) catch return ExtractError.OutOfMemory;

    for (pixels) |pixel| {
        if (pixel.a <= 2) continue;

        var found = false;
        for (color_map.items) |*item| {
            if (item.similar(pixel.r, pixel.g, pixel.b)) {
                item.count += 1;
                found = true;
                break;
            }
        }

        if (!found) {
            color_map.append(allocator, .{
                .r = pixel.r,
                .g = pixel.g,
                .b = pixel.b,
                .count = 1,
            }) catch return ExtractError.OutOfMemory;
        }
    }

    if (color_map.items.len == 0) {
        return ExtractError.NoColors;
    }

    std.mem.sort(QuantizedColorAndCount, color_map.items, {}, QuantizedColorAndCount.lessThan);

    const num_colors = @min(20, color_map.items.len);
    const colors = allocator.alloc([]const u8, num_colors) catch return ExtractError.OutOfMemory;
    errdefer {
        for (colors) |c| {
            allocator.free(c);
        }
        allocator.free(colors);
    }

    for (0..num_colors) |i| {
        const c = color_map.items[i];
        colors[i] = color_utils.rgbToHexAlloc(allocator, c.r, c.g, c.b) catch return ExtractError.OutOfMemory;
    }

    return PaletteResult{ .colors = colors, .allocator = allocator };
}

fn processImageGeneric(
    allocator: std.mem.Allocator,
    img: *zigimg.Image,
) ExtractError!PaletteResult {
    var color_map = std.ArrayList(ColorAndCount){};
    defer color_map.deinit(allocator);
    color_map.ensureTotalCapacity(allocator, 512) catch return ExtractError.OutOfMemory;

    var color_it = img.iterator();
    while (color_it.next()) |color| {
        if (color.a < 0.01) continue;

        var found = false;
        for (color_map.items) |*item| {
            if (ColorAndCount.colorSimilar(item.color, color)) {
                item.count += 1;
                found = true;
                break;
            }
        }

        if (!found) {
            color_map.append(allocator, .{
                .color = color,
                .count = 1,
            }) catch return ExtractError.OutOfMemory;
        }
    }

    if (color_map.items.len == 0) {
        return ExtractError.NoColors;
    }

    std.mem.sort(ColorAndCount, color_map.items, {}, ColorAndCount.lessThan);

    const num_colors = @min(20, color_map.items.len);
    const colors = allocator.alloc([]const u8, num_colors) catch return ExtractError.OutOfMemory;
    errdefer {
        for (colors) |c| {
            allocator.free(c);
        }
        allocator.free(colors);
    }

    for (0..num_colors) |i| {
        const c = color_map.items[i].color;
        colors[i] = color_utils.colorf32ToHex(allocator, c.r, c.g, c.b) catch return ExtractError.OutOfMemory;
    }

    return PaletteResult{
        .colors = colors,
        .allocator = allocator,
    };
}

pub const ThemeType = theme_common.ThemeType;
pub const ThemeAppearance = theme_common.ThemeAppearance;

pub fn returnThemeJson(
    allocator: std.mem.Allocator,
    colors: []const []const u8,
    theme_type: ThemeType,
    theme_name: []const u8,
    overrides: ThemeOverrides,
    appearance: ?ThemeAppearance,
    boost_coefficient: ?f32,
) ![]const u8 {
    const effective_boost_coefficient = boost_coefficient orelse 1.0;

    switch (theme_type) {
        .vscode => {
            const response = try vscode.generateVSCodeTheme(allocator, colors, theme_name, overrides, appearance, effective_boost_coefficient);
            return try std.json.Stringify.valueAlloc(allocator, response, .{ .whitespace = .minified, .emit_null_optional_fields = false });
        },
        .zed => {
            const response = try zed.generateZedTheme(allocator, colors, theme_name, overrides, appearance, effective_boost_coefficient);
            return try std.json.Stringify.valueAlloc(allocator, response, .{ .whitespace = .minified, .emit_null_optional_fields = false });
        },
    }
}

const ThemeRequest = struct {
    colors: []const ColorInput,
    type: []const u8,
    name: ?[]const u8 = null,
    overrides: ?ThemeOverrides = null,
    appearance: ?ThemeAppearance = null,
    boostCoefficient: ?f32 = null,
};

const ColorInput = struct {
    hex: []const u8,
};

pub fn handleExtractPalette(allocator: std.mem.Allocator, request_body: []const u8) ![]const u8 {
    if (request_body.len == 0) return error.NoImageProvided;

    var result = try extractPaletteFromBytes(allocator, request_body);
    defer result.deinit();

    var json_array = std.ArrayList(u8){};
    errdefer json_array.deinit(allocator);

    try json_array.appendSlice(allocator, "{\"palette\":[");
    for (result.colors, 0..) |color, i| {
        if (i > 0) try json_array.appendSlice(allocator, ",");
        try json_array.appendSlice(allocator, "{\"hex\":\"");
        try json_array.appendSlice(allocator, color);
        try json_array.appendSlice(allocator, "\"}");
    }
    try json_array.appendSlice(allocator, "]}");

    return try json_array.toOwnedSlice(allocator);
}

pub fn handleGenerateTheme(allocator: std.mem.Allocator, request_body: []const u8) ![]const u8 {
    const parsed = try std.json.parseFromSlice(ThemeRequest, allocator, request_body, .{
        .ignore_unknown_fields = true,
    });
    defer parsed.deinit();

    const req = parsed.value;

    if (req.colors.len == 0) {
        return error.NoColors;
    }

    const colors = try allocator.alloc([]const u8, req.colors.len);
    defer allocator.free(colors);

    for (req.colors, 0..) |c, i| {
        colors[i] = c.hex;
    }

    const theme_type: ThemeType = if (std.mem.eql(u8, req.type, "zed")) .zed else .vscode;
    const theme_name = req.name orelse "Generated Theme";
    const overrides = req.overrides orelse ThemeOverrides{};
    const appearance = req.appearance;
    const boost_coefficient = req.boostCoefficient;

    return try returnThemeJson(allocator, colors, theme_type, theme_name, overrides, appearance, boost_coefficient);
}

pub const GenerateOverridableRequest = theme_common.GenerateOverridableRequest;

pub fn handleGenerateOverridable(allocator: std.mem.Allocator, request_body: []const u8) ![]const u8 {
    const parsed = std.json.parseFromSlice(GenerateOverridableRequest, allocator, request_body, .{
        .ignore_unknown_fields = true,
    }) catch |err| {
        return err;
    };
    defer parsed.deinit();

    const req: GenerateOverridableRequest = parsed.value;

    switch (req.themeType) {
        .vscode => {
            const response = vscode.generateOverridableFromVSCodeThemeValue(allocator, req) catch |err| {
                return err;
            };
            return try std.json.Stringify.valueAlloc(allocator, response, .{ .whitespace = .minified, .emit_null_optional_fields = false });
        },
        .zed => {
            const response = zed.generateOverridableFromZedThemeValue(allocator, req) catch |err| {
                return err;
            };
            return try std.json.Stringify.valueAlloc(allocator, response, .{ .whitespace = .minified, .emit_null_optional_fields = false });
        },
    }
}

test "handleGenerateTheme rejects empty colors" {
    var arena = std.heap.ArenaAllocator.init(std.testing.allocator);
    defer arena.deinit();
    const allocator = arena.allocator();
    const body = "{\"colors\":[],\"type\":\"vscode\"}";
    try std.testing.expectError(error.NoColors, handleGenerateTheme(allocator, body));
}

test "returnThemeJson produces vscode theme payload" {
    var arena = std.heap.ArenaAllocator.init(std.testing.allocator);
    defer arena.deinit();
    const allocator = arena.allocator();
    const colors = [_][]const u8{
        "#112233",
        "#445566",
        "#778899",
        "#AABBCC",
        "#DDEEFF",
        "#FFAA00",
        "#00AABB",
        "#BB00AA",
        "#33CC66",
        "#CC3366",
    };

    const json = try returnThemeJson(allocator, &colors, .vscode, "Test Theme", ThemeOverrides{}, null, null);

    const parsed = try std.json.parseFromSlice(std.json.Value, allocator, json, .{});
    defer parsed.deinit();

    const root_obj = switch (parsed.value) {
        .object => |obj| obj,
        else => return error.InvalidTheme,
    };

    try std.testing.expect(root_obj.get("theme") != null);
    try std.testing.expect(root_obj.get("themeOverrides") != null);
}
