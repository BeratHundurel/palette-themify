const std = @import("std");
const tk = @import("tokamak");
const palette_api = @import("palette_api.zig");

const PORT: u16 = 8089;
const MAX_BODY_SIZE: usize = 50 * 1024 * 1024;

const routes: []const tk.Route = &.{
    .{ .handler = &corsMiddleware, .children = &.{
        .get("/health", health),
        .post0("/extract-palette", extractPalette),
        .post0("/generate-theme", generateTheme),
        .post0("/generate-overridable", generateOverridable),
    } },
};

fn corsMiddleware(ctx: *tk.Context) anyerror!void {
    ctx.res.headers.add("Access-Control-Allow-Origin", "*");
    ctx.res.headers.add("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS");
    ctx.res.headers.add("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept");
    ctx.res.headers.add("Access-Control-Max-Age", "86400");

    if (ctx.req.method == .OPTIONS) {
        try ctx.res.write();
        return;
    }

    ctx.next() catch |err| {
        if (err == error.NotFound) {
            ctx.res.status = 404;
            ctx.res.content_type = .JSON;
            ctx.res.body = "{\"error\":\"Not found\"}";
            return;
        }
        return err;
    };
}

fn health() []const u8 {
    return "{\"status\":\"ok\"}";
}

fn extractPalette(ctx: *tk.Context, allocator: std.mem.Allocator) anyerror!void {
    ctx.res.content_type = .JSON;

    const body = ctx.req.body() orelse {
        ctx.res.status = 400;
        ctx.res.body = "{\"error\":\"Missing request body\"}";
        return;
    };

    if (body.len == 0) {
        ctx.res.status = 400;
        ctx.res.body = "{\"error\":\"Request body is empty\"}";
        return;
    }

    const result = palette_api.handleExtractPalette(allocator, body) catch |err| {
        ctx.res.status = 400;
        ctx.res.body = switch (err) {
            error.InvalidContentType => "{\"error\":\"Invalid content type, expected multipart/form-data  \"}",
            error.NoImageProvided => "{\"error\":\"No image file provided in request\"}",
            error.NotAnImage => "{\"error\":\"Could not decode image - unsupported format\"}",
            error.NoColors => "{\"error\":\"No colors found in image\"}",
            else => "{\"error\":\"Internal server error\"}",
        };
        return;
    };

    ctx.res.body = result;
}

fn generateTheme(ctx: *tk.Context, allocator: std.mem.Allocator) anyerror!void {
    ctx.res.content_type = .JSON;

    const body = ctx.req.body() orelse {
        ctx.res.status = 400;
        ctx.res.body = "{\"error\":\"Missing request body\"}";
        return;
    };

    const result = palette_api.handleGenerateTheme(allocator, body) catch |err| {
        ctx.res.status = 400;
        ctx.res.body = switch (err) {
            error.NotEnoughColors => "{\"error\":\"Not enough colors - please provide at least 5 colors\"}",
            else => "{\"error\":\"Failed to generate theme\"}",
        };
        return;
    };

    ctx.res.body = result;
}

fn generateOverridable(ctx: *tk.Context, allocator: std.mem.Allocator) anyerror!void {
    ctx.res.content_type = .JSON;

    const body = ctx.req.body() orelse {
        ctx.res.status = 400;
        ctx.res.body = "{\"error\":\"Missing request body\"}";
        return;
    };

    const result = palette_api.handleGenerateOverridable(allocator, body) catch |err| {
        ctx.res.status = 400;
        ctx.res.body = switch (err) {
            error.InvalidTheme => "{\"error\":\"Please provide a zed editor theme\"}",
            else => "{\"error\":\"Falied to generate theme\"}",
        };
        return;
    };

    ctx.res.body = result;
}

pub fn main() !void {
    var gpa: std.heap.DebugAllocator(.{}) = .init;
    defer {
        const deinit_status = gpa.deinit();
        if (deinit_status == .leak) {
            std.log.err("memory leak", .{});
        }
    }
    const allocator = gpa.allocator();

    std.log.info("Zig Palette API starting on http://localhost:{d}", .{PORT});

    var server = try tk.Server.init(allocator, routes, .{
        .listen = .{ .port = PORT },
        .request = .{ .max_body_size = MAX_BODY_SIZE },
    });
    defer server.deinit();

    try server.start();
}
