const std = @import("std");
const types = @import("vscode_types.zig");

pub const HSL = struct {
    h: f32,
    s: f32,
    l: f32,
};

pub const RGB = struct {
    r: u8,
    g: u8,
    b: u8,
};

/// LAB color space for perceptual color calculations
pub const LAB = struct {
    l: f32,
    a: f32,
    b: f32,
};

/// Color harmony schemes based on color wheel relationships
pub const HarmonyScheme = enum {
    complementary,
    triadic,
    analogous,
    @"split-complementary",
};

pub const BackgroundForegroundSelection = struct {
    background_index: usize,
    foreground_index: usize,
    remaining_indices: []usize,
};

pub const SemanticColors = struct {
    error_color: []const u8,
    warning_color: []const u8,
    success_color: []const u8,
    info_color: []const u8,
};

const HexBuffer = struct {
    data: [10]u8 = undefined,
    len: u8 = 0,

    pub fn slice(self: *const HexBuffer) []const u8 {
        return self.data[0..self.len];
    }
};

var hex_buffer_pool: [512]HexBuffer = undefined;
var hex_buffer_index: usize = 0;

fn allocHexBuffer() *HexBuffer {
    const buf = &hex_buffer_pool[hex_buffer_index];
    hex_buffer_index = (hex_buffer_index + 1) % hex_buffer_pool.len;
    return buf;
}

fn formatHex(r: u8, g: u8, b: u8) []const u8 {
    const buf = allocHexBuffer();
    const result = std.fmt.bufPrint(&buf.data, "#{X:0>2}{X:0>2}{X:0>2}", .{ r, g, b }) catch return "#000000";
    buf.len = @intCast(result.len);
    return result;
}

fn formatHexWithAlpha(hex: []const u8, alpha: []const u8) []const u8 {
    const buf = allocHexBuffer();
    const result = std.fmt.bufPrint(&buf.data, "{s}{s}", .{ hex, alpha }) catch return "#00000000";
    buf.len = @intCast(result.len);
    return result;
}

pub fn findSemanticColors(colors: []const []const u8) SemanticColors {
    const target_red = RGB{ .r = 220, .g = 60, .b = 60 };
    const target_orange = RGB{ .r = 230, .g = 160, .b = 50 };
    const target_green = RGB{ .r = 80, .g = 180, .b = 80 };
    const target_blue = RGB{ .r = 80, .g = 160, .b = 220 };

    var best_red: []const u8 = colors[0];
    var best_orange: []const u8 = colors[0];
    var best_green: []const u8 = colors[0];
    var best_blue: []const u8 = colors[0];

    var min_red_dist: f32 = std.math.floatMax(f32);
    var min_orange_dist: f32 = std.math.floatMax(f32);
    var min_green_dist: f32 = std.math.floatMax(f32);
    var min_blue_dist: f32 = std.math.floatMax(f32);

    for (colors) |color| {
        const rgb = parseHexToRgb(color);
        const hsl = hexToHsl(color);

        if (hsl.s < 0.2 or hsl.l < 0.15 or hsl.l > 0.85) continue;

        const red_dist = rgbDistanceFromRgb(rgb, target_red);
        const orange_dist = rgbDistanceFromRgb(rgb, target_orange);
        const green_dist = rgbDistanceFromRgb(rgb, target_green);
        const blue_dist = rgbDistanceFromRgb(rgb, target_blue);

        if (red_dist < min_red_dist) {
            min_red_dist = red_dist;
            best_red = color;
        }
        if (orange_dist < min_orange_dist) {
            min_orange_dist = orange_dist;
            best_orange = color;
        }
        if (green_dist < min_green_dist) {
            min_green_dist = green_dist;
            best_green = color;
        }
        if (blue_dist < min_blue_dist) {
            min_blue_dist = blue_dist;
            best_blue = color;
        }
    }

    return SemanticColors{
        .error_color = best_red,
        .warning_color = best_orange,
        .success_color = best_green,
        .info_color = best_blue,
    };
}

inline fn parseHexToRgb(hex: []const u8) RGB {
    const r = std.fmt.parseInt(u8, hex[1..3], 16) catch 0;
    const g = std.fmt.parseInt(u8, hex[3..5], 16) catch 0;
    const b = std.fmt.parseInt(u8, hex[5..7], 16) catch 0;
    return RGB{ .r = r, .g = g, .b = b };
}

inline fn rgbDistanceFromRgb(rgb1: RGB, rgb2: RGB) f32 {
    const dr = @as(f32, @floatFromInt(@as(i32, rgb1.r) - @as(i32, rgb2.r)));
    const dg = @as(f32, @floatFromInt(@as(i32, rgb1.g) - @as(i32, rgb2.g)));
    const db = @as(f32, @floatFromInt(@as(i32, rgb1.b) - @as(i32, rgb2.b)));
    return @sqrt(dr * dr + dg * dg + db * db);
}

pub fn rgbDistance(hex_1: []const u8, hex_2: []const u8) f32 {
    const rgb1 = parseHexToRgb(hex_1);
    const rgb2 = parseHexToRgb(hex_2);
    return rgbDistanceFromRgb(rgb1, rgb2);
}

pub fn perceptualDistance(hex1: []const u8, hex2: []const u8) f32 {
    const rgb1 = parseHexToRgb(hex1);
    const rgb2 = parseHexToRgb(hex2);
    const lab1 = rgbToLab(rgb1);
    const lab2 = rgbToLab(rgb2);
    return deltaE94(lab1, lab2);
}

/// Calculates relative luminance per WCAG 2.0 spec for contrast ratio calculations.
inline fn getLuminanceFromRgb(rgb: RGB) f32 {
    const rsRGB = @as(f32, @floatFromInt(rgb.r)) / 255.0;
    const gsRGB = @as(f32, @floatFromInt(rgb.g)) / 255.0;
    const bsRGB = @as(f32, @floatFromInt(rgb.b)) / 255.0;

    const rLinear = if (rsRGB <= 0.03928) rsRGB / 12.92 else std.math.pow(f32, (rsRGB + 0.055) / 1.055, 2.4);
    const gLinear = if (gsRGB <= 0.03928) gsRGB / 12.92 else std.math.pow(f32, (gsRGB + 0.055) / 1.055, 2.4);
    const bLinear = if (bsRGB <= 0.03928) bsRGB / 12.92 else std.math.pow(f32, (bsRGB + 0.055) / 1.055, 2.4);

    return 0.2126 * rLinear + 0.7152 * gLinear + 0.0722 * bLinear;
}

pub fn getLuminance(hex: []const u8) f32 {
    const rgb = parseHexToRgb(hex);
    return getLuminanceFromRgb(rgb);
}

pub fn darkenColor(hex: []const u8, percent: f32) []const u8 {
    const rgb = parseHexToRgb(hex);
    const factor = 1.0 - percent;
    const r: u8 = @intFromFloat(@max(@as(f32, @floatFromInt(rgb.r)) * factor, 0.0));
    const g: u8 = @intFromFloat(@max(@as(f32, @floatFromInt(rgb.g)) * factor, 0.0));
    const b: u8 = @intFromFloat(@max(@as(f32, @floatFromInt(rgb.b)) * factor, 0.0));
    return formatHex(r, g, b);
}

pub fn lightenColor(hex: []const u8, percent: f32) []const u8 {
    const rgb = parseHexToRgb(hex);
    const r: u8 = @intFromFloat(@min(@as(f32, @floatFromInt(rgb.r)) + (255.0 - @as(f32, @floatFromInt(rgb.r))) * percent, 255.0));
    const g: u8 = @intFromFloat(@min(@as(f32, @floatFromInt(rgb.g)) + (255.0 - @as(f32, @floatFromInt(rgb.g))) * percent, 255.0));
    const b: u8 = @intFromFloat(@min(@as(f32, @floatFromInt(rgb.b)) + (255.0 - @as(f32, @floatFromInt(rgb.b))) * percent, 255.0));
    return formatHex(r, g, b);
}

pub fn addAlpha(hex: []const u8, alpha: []const u8) []const u8 {
    return formatHexWithAlpha(hex, alpha);
}

pub fn contrastRatio(hex1: []const u8, hex2: []const u8) f32 {
    const lum1 = getLuminance(hex1);
    const lum2 = getLuminance(hex2);

    const lighter = @max(lum1, lum2);
    const darker = @min(lum1, lum2);

    return (lighter + 0.05) / (darker + 0.05);
}

/// Uses YIQ luminance formula (weighted RGB based on human perception) to determine if a color is dark.
inline fn isDarkColorFromRgb(rgb: RGB) bool {
    const brightness = (@as(f32, @floatFromInt(rgb.r)) * 299.0 + @as(f32, @floatFromInt(rgb.g)) * 587.0 + @as(f32, @floatFromInt(rgb.b)) * 114.0) / 255000.0;
    return brightness < 0.5;
}

pub fn isDarkColor(hex: []const u8) bool {
    const rgb = parseHexToRgb(hex);
    return isDarkColorFromRgb(rgb);
}

/// Rules for reducing saturation/brightness
const VibrancyRule = struct {
    min_s: f32,
    min_l: f32,
    max_l: ?f32 = null,

    target_l_min: f32,
    target_l_max: f32,

    s_min_mul: f32,
    s_max_mul: f32,

    ease_range: f32,
};

/// Desired rules
const dark_bg_rules = [_]VibrancyRule{
    .{
        .min_s = 0.6,
        .min_l = 0.55,
        .max_l = null,
        .target_l_min = 0.52,
        .target_l_max = 0.65,
        .s_min_mul = 0.85,
        .s_max_mul = 0.70,
        .ease_range = 0.35,
    },
};

const light_bg_rules = [_]VibrancyRule{
    .{
        .min_s = 0.8,
        .min_l = 0.4,
        .max_l = 0.6,
        .target_l_min = 0.4,
        .target_l_max = 0.6,
        .s_min_mul = 0.7,
        .s_max_mul = 0.7,
        .ease_range = 0.2,
    },
};

inline fn easeOut(t: f32) f32 {
    return 1.0 - (1.0 - t) * (1.0 - t);
}

/// Adjusts a foreground color to meet minimum contrast ratio against a background.
/// First reduces saturation/brightness of overly vibrant colors, then iteratively
/// lightens or darkens until min_contrast is met.
pub fn adjustForContrast(fg: []const u8, bg: []const u8, min_contrast: f32) []const u8 {
    var color = fg;
    var iterations: u32 = 0;
    const dark_bg = isDarkColor(bg);

    const rules = if (dark_bg) &dark_bg_rules else &light_bg_rules;
    const initial_hsl = hexToHsl(color);
    for (rules) |rule| {
        if (initial_hsl.s <= rule.min_s) continue;
        if (initial_hsl.l <= rule.min_l) continue;
        if (rule.max_l) |max_l| {
            if (initial_hsl.l >= max_l) continue;
        }

        const t_raw = (initial_hsl.l - rule.min_l) / rule.ease_range;
        const t = @min(@max(t_raw, 0.0), 1.0);
        const eased = easeOut(t);

        const new_l = rule.target_l_min +
            eased * (rule.target_l_max - rule.target_l_min);

        const s_mul = rule.s_min_mul +
            eased * (rule.s_max_mul - rule.s_min_mul);

        const new_s = initial_hsl.s * s_mul;

        const rgb = hslToRgb(initial_hsl.h, new_s, new_l);
        color = rgbToHex(rgb.r, rgb.g, rgb.b);
        break;
    }

    while (contrastRatio(color, bg) < min_contrast and iterations < 20) : (iterations += 1) {
        const step: f32 = if (iterations < 4) 0.12 else 0.06;
        color = if (dark_bg) lightenColor(color, step) else darkenColor(color, step);
    }

    return color;
}

/// Multi-stage fallback to ensure readable contrast. Tries adjustment first, then
/// creates a tinted fallback preserving hue, finally falls back to near-neutral if needed.
pub fn ensureReadableContrast(proposed_color: []const u8, background: []const u8, min_contrast: f32) []const u8 {
    if (contrastRatio(proposed_color, background) >= min_contrast) {
        return proposed_color;
    }

    const adjusted = adjustForContrast(proposed_color, background, min_contrast);
    if (contrastRatio(adjusted, background) >= min_contrast) {
        return adjusted;
    }

    const dark_bg = isDarkColor(background);
    const bg_hsl = hexToHsl(background);
    const proposed_hsl = hexToHsl(proposed_color);

    const target_lightness: f32 = if (dark_bg) 0.85 else 0.15;
    const rgb = hslToRgb(proposed_hsl.h, @max(proposed_hsl.s * 0.7, 0.1), target_lightness);
    const tinted_fallback = rgbToHex(rgb.r, rgb.g, rgb.b);

    if (contrastRatio(tinted_fallback, background) >= min_contrast) {
        return tinted_fallback;
    }

    const neutral_lightness: f32 = if (dark_bg) 0.9 else 0.1;
    const neutral_rgb = hslToRgb(bg_hsl.h, 0.05, neutral_lightness);
    return rgbToHex(neutral_rgb.r, neutral_rgb.g, neutral_rgb.b);
}

pub fn hexToHsl(hex: []const u8) HSL {
    const rgb = parseHexToRgb(hex);
    const r = @as(f32, @floatFromInt(rgb.r)) / 255.0;
    const g = @as(f32, @floatFromInt(rgb.g)) / 255.0;
    const b = @as(f32, @floatFromInt(rgb.b)) / 255.0;

    const max_val = @max(@max(r, g), b);
    const min_val = @min(@min(r, g), b);
    const delta = max_val - min_val;

    var h: f32 = 0.0;
    var s: f32 = 0.0;
    const l: f32 = (max_val + min_val) / 2.0;

    if (delta != 0.0) {
        s = if (l > 0.5) delta / (2.0 - max_val - min_val) else delta / (max_val + min_val);

        if (max_val == r) {
            h = ((g - b) / delta + (if (g < b) @as(f32, 6.0) else @as(f32, 0.0))) / 6.0;
        } else if (max_val == g) {
            h = ((b - r) / delta + 2.0) / 6.0;
        } else {
            h = ((r - g) / delta + 4.0) / 6.0;
        }
    }

    return HSL{ .h = h, .s = s, .l = l };
}

fn hueToRgb(p: f32, q: f32, t_input: f32) f32 {
    var t = t_input;
    if (t < 0.0) t += 1.0;
    if (t > 1.0) t -= 1.0;
    if (t < 1.0 / 6.0) return p + (q - p) * 6.0 * t;
    if (t < 1.0 / 2.0) return q;
    if (t < 2.0 / 3.0) return p + (q - p) * (2.0 / 3.0 - t) * 6.0;
    return p;
}

pub fn hslToRgb(h: f32, s: f32, l: f32) RGB {
    var r: f32 = 0.0;
    var g: f32 = 0.0;
    var b: f32 = 0.0;

    if (s == 0.0) {
        r = l;
        g = l;
        b = l;
    } else {
        const q = if (l < 0.5) l * (1.0 + s) else l + s - l * s;
        const p = 2.0 * l - q;
        r = hueToRgb(p, q, h + 1.0 / 3.0);
        g = hueToRgb(p, q, h);
        b = hueToRgb(p, q, h - 1.0 / 3.0);
    }

    return RGB{
        .r = @as(u8, @intFromFloat(@round(r * 255.0))),
        .g = @as(u8, @intFromFloat(@round(g * 255.0))),
        .b = @as(u8, @intFromFloat(@round(b * 255.0))),
    };
}

pub fn rgbToHex(r: u8, g: u8, b: u8) []const u8 {
    return formatHex(r, g, b);
}

pub fn rgbToHexAlloc(allocator: std.mem.Allocator, r: u8, g: u8, b: u8) ![]const u8 {
    return try std.fmt.allocPrint(allocator, "#{X:0>2}{X:0>2}{X:0>2}", .{ r, g, b });
}

pub fn colorf32ToRgb(r: f32, g: f32, b: f32) RGB {
    return RGB{
        .r = @intFromFloat(@min(255.0, @max(0.0, r * 255.0))),
        .g = @intFromFloat(@min(255.0, @max(0.0, g * 255.0))),
        .b = @intFromFloat(@min(255.0, @max(0.0, b * 255.0))),
    };
}

pub fn colorf32ToHex(allocator: std.mem.Allocator, r: f32, g: f32, b: f32) ![]const u8 {
    const rgb = colorf32ToRgb(r, g, b);
    return try rgbToHexAlloc(allocator, rgb.r, rgb.g, rgb.b);
}

/// Converts RGB to CIE LAB color space for perceptually uniform color comparisons.
/// Uses sRGB with D65 illuminant as the reference white point.
fn rgbToLab(rgb: RGB) LAB {
    var r = @as(f32, @floatFromInt(rgb.r)) / 255.0;
    var g = @as(f32, @floatFromInt(rgb.g)) / 255.0;
    var b = @as(f32, @floatFromInt(rgb.b)) / 255.0;

    r = if (r > 0.04045) std.math.pow(f32, (r + 0.055) / 1.055, 2.4) else r / 12.92;
    g = if (g > 0.04045) std.math.pow(f32, (g + 0.055) / 1.055, 2.4) else g / 12.92;
    b = if (b > 0.04045) std.math.pow(f32, (b + 0.055) / 1.055, 2.4) else b / 12.92;

    const x = (r * 0.4124564 + g * 0.3575761 + b * 0.1804375) / 0.95047;
    const y = (r * 0.2126729 + g * 0.7151522 + b * 0.0721750) / 1.00000;
    const z = (r * 0.0193339 + g * 0.1191920 + b * 0.9503041) / 1.08883;

    const epsilon: f32 = 0.008856;
    const kappa: f32 = 903.3;

    const fx = if (x > epsilon) std.math.pow(f32, x, 1.0 / 3.0) else (kappa * x + 16.0) / 116.0;
    const fy = if (y > epsilon) std.math.pow(f32, y, 1.0 / 3.0) else (kappa * y + 16.0) / 116.0;
    const fz = if (z > epsilon) std.math.pow(f32, z, 1.0 / 3.0) else (kappa * z + 16.0) / 116.0;

    return LAB{
        .l = 116.0 * fy - 16.0,
        .a = 500.0 * (fx - fy),
        .b = 200.0 * (fy - fz),
    };
}

/// CIE94 color difference formula - measures perceptual distance between two LAB colors.
/// Simpler than CIEDE2000 but still provides good perceptual uniformity for theme generation.
fn deltaE94(lab1: LAB, lab2: LAB) f32 {
    const dl = lab1.l - lab2.l;
    const da = lab1.a - lab2.a;
    const db = lab1.b - lab2.b;

    const c1 = @sqrt(lab1.a * lab1.a + lab1.b * lab1.b);
    const c2 = @sqrt(lab2.a * lab2.a + lab2.b * lab2.b);
    const dc = c1 - c2;

    const dh_sq = da * da + db * db - dc * dc;
    const dh = if (dh_sq > 0) @sqrt(dh_sq) else 0.0;

    const kl: f32 = 1.0;
    const kc: f32 = 1.0;
    const kh: f32 = 1.0;

    const sl: f32 = 1.0;
    const sc = 1.0 + 0.045 * c1;
    const sh = 1.0 + 0.015 * c1;

    const dl_term = dl / (kl * sl);
    const dc_term = dc / (kc * sc);
    const dh_term = dh / (kh * sh);

    return @sqrt(dl_term * dl_term + dc_term * dc_term + dh_term * dh_term);
}

fn deltaE94FromLab(lab1: LAB, lab2: LAB) f32 {
    const dl = lab1.l - lab2.l;
    const da = lab1.a - lab2.a;
    const db = lab1.b - lab2.b;

    const c1 = @sqrt(lab1.a * lab1.a + lab1.b * lab1.b);
    const c2 = @sqrt(lab2.a * lab2.a + lab2.b * lab2.b);
    const dc = c1 - c2;

    const dh_sq = da * da + db * db - dc * dc;
    const dh = if (dh_sq > 0) @sqrt(dh_sq) else 0.0;

    const sc = 1.0 + 0.045 * c1;
    const sh = 1.0 + 0.015 * c1;

    const dl_term = dl;
    const dc_term = dc / sc;
    const dh_term = dh / sh;

    return @sqrt(dl_term * dl_term + dc_term * dc_term + dh_term * dh_term);
}

/// Selects `count` maximally diverse colors using a greedy farthest-point sampling algorithm.
/// Starts with the color most distant from all others, then iteratively picks the color
/// with maximum minimum distance to already-selected colors.
/// Pre-computes LAB values to avoid repeated hex parsing.
pub fn selectDiverseColors(allocator: std.mem.Allocator, colors: []const []const u8, count: usize) ![][]const u8 {
    const n = colors.len;
    if (n == 0) return try allocator.alloc([]const u8, 0);

    const lab_values = try allocator.alloc(LAB, n);
    defer allocator.free(lab_values);
    for (colors, 0..) |color, i| {
        lab_values[i] = rgbToLab(parseHexToRgb(color));
    }

    const selected_indices = try allocator.alloc(usize, @min(count, n));
    defer allocator.free(selected_indices);
    var selected_count: usize = 0;

    var best_start_idx: usize = 0;
    var best_start_score: f32 = 0.0;
    for (0..n) |i| {
        var total_dist: f32 = 0.0;
        for (0..n) |j| {
            if (i != j) {
                total_dist += deltaE94FromLab(lab_values[i], lab_values[j]);
            }
        }
        if (total_dist > best_start_score) {
            best_start_score = total_dist;
            best_start_idx = i;
        }
    }
    selected_indices[selected_count] = best_start_idx;
    selected_count += 1;

    while (selected_count < count and selected_count < n) {
        var max_min_distance: f32 = 0.0;
        var best_candidate_idx: ?usize = null;

        for (0..n) |candidate_idx| {
            var already_selected = false;
            for (selected_indices[0..selected_count]) |sel_idx| {
                if (candidate_idx == sel_idx) {
                    already_selected = true;
                    break;
                }
            }
            if (already_selected) continue;

            var min_distance: f32 = std.math.floatMax(f32);
            for (selected_indices[0..selected_count]) |sel_idx| {
                const dist = deltaE94FromLab(lab_values[candidate_idx], lab_values[sel_idx]);
                if (dist < min_distance) {
                    min_distance = dist;
                }
            }

            if (min_distance > max_min_distance) {
                max_min_distance = min_distance;
                best_candidate_idx = candidate_idx;
            }
        }

        if (best_candidate_idx) |idx| {
            selected_indices[selected_count] = idx;
            selected_count += 1;
        } else {
            break;
        }
    }

    const result = try allocator.alloc([]const u8, selected_count);
    for (selected_indices[0..selected_count], 0..) |idx, i| {
        result[i] = colors[idx];
    }

    return result;
}

/// Generates a color based on color wheel harmony theory (complementary, triadic, etc.).
pub fn getHarmonicColor(base_color: []const u8, scheme: HarmonyScheme) []const u8 {
    const hsl = hexToHsl(base_color);

    const target_h: f32 = switch (scheme) {
        .complementary => @mod(hsl.h + 0.5, 1.0),
        .triadic => @mod(hsl.h + 1.0 / 3.0, 1.0),
        .analogous => @mod(hsl.h + 1.0 / 12.0 + 1.0, 1.0),
        .@"split-complementary" => @mod(hsl.h + 0.5 - 1.0 / 12.0 + 1.0, 1.0),
    };

    const rgb = hslToRgb(target_h, hsl.s, hsl.l);
    return rgbToHex(rgb.r, rgb.g, rgb.b);
}

/// Scores a color's suitability as a background based on low saturation and
/// appropriate luminance (dark for dark themes, light for light themes).
fn calculateBackgroundScore(hex: []const u8, prefer_dark: bool) f32 {
    const hsl = hexToHsl(hex);
    const luminance = getLuminance(hex);

    const saturation_score = 1.0 - hsl.s;

    var luminance_score: f32 = 0.0;
    if (prefer_dark) {
        if (luminance < 0.15) {
            luminance_score = 1.0;
        } else if (luminance < 0.4) {
            luminance_score = 0.7 - (luminance - 0.15) * 2.0;
        } else {
            luminance_score = 0.2;
        }
    } else {
        if (luminance > 0.85) {
            luminance_score = 1.0;
        } else if (luminance > 0.6) {
            luminance_score = 0.7 + (luminance - 0.6) * 1.2;
        } else {
            luminance_score = 0.2;
        }
    }

    return saturation_score * 0.6 + luminance_score * 0.4;
}

pub fn selectBackgroundColor(colors: []const []const u8, prefer_dark: bool) usize {
    var best_index: usize = 0;
    var best_score: f32 = 0.0;

    for (colors, 0..) |color, i| {
        const score = calculateBackgroundScore(color, prefer_dark);
        if (score > best_score) {
            best_score = score;
            best_index = i;
        }
    }

    return best_index;
}

pub fn selectForegroundColor(colors: []const []const u8, background: []const u8, exclude_index: usize) usize {
    var best_index: usize = if (exclude_index == 0) 1 else 0;
    var best_contrast: f32 = 0.0;

    for (colors, 0..) |color, i| {
        if (i == exclude_index) continue;

        const contrast = contrastRatio(color, background);
        if (contrast > best_contrast) {
            best_contrast = contrast;
            best_index = i;
        }
    }

    return best_index;
}

pub fn selectBackgroundAndForeground(allocator: std.mem.Allocator, colors: []const []const u8, prefer_dark: bool) !BackgroundForegroundSelection {
    const bg_index = selectBackgroundColor(colors, prefer_dark);
    const fg_index = selectForegroundColor(colors, colors[bg_index], bg_index);

    var remaining = std.ArrayList(usize){};
    defer remaining.deinit(allocator);

    for (0..colors.len) |i| {
        if (i != bg_index and i != fg_index) {
            try remaining.append(allocator, i);
        }
    }

    const remaining_slice = try allocator.dupe(usize, remaining.items);

    return BackgroundForegroundSelection{
        .background_index = bg_index,
        .foreground_index = fg_index,
        .remaining_indices = remaining_slice,
    };
}

test "hexToHsl and hslToRgb roundtrip" {
    const hsl = hexToHsl("#336699");
    const rgb = hslToRgb(hsl.h, hsl.s, hsl.l);
    try std.testing.expectEqual(@as(u8, 0x33), rgb.r);
    try std.testing.expectEqual(@as(u8, 0x66), rgb.g);
    try std.testing.expectEqual(@as(u8, 0x99), rgb.b);
}

test "luminance and contrast ratio basics" {
    const black = "#000000";
    const white = "#FFFFFF";
    try std.testing.expectApproxEqAbs(@as(f32, 0.0), getLuminance(black), 0.0001);
    try std.testing.expectApproxEqAbs(@as(f32, 1.0), getLuminance(white), 0.0001);
    try std.testing.expectApproxEqAbs(@as(f32, 21.0), contrastRatio(black, white), 0.01);
}

test "ensureReadableContrast returns color meeting minimum" {
    const background = "#000000";
    const proposed = "#111111";
    const adjusted = ensureReadableContrast(proposed, background, 7.0);
    try std.testing.expect(contrastRatio(adjusted, background) >= 7.0);
}

test "selectDiverseColors returns requested count" {
    const allocator = std.testing.allocator;
    const colors = [_][]const u8{ "#112233", "#445566", "#778899", "#AABBCC" };
    const selected = try selectDiverseColors(allocator, &colors, 2);
    defer allocator.free(selected);
    try std.testing.expectEqual(@as(usize, 2), selected.len);
}
