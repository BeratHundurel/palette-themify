const std = @import("std");
const httpz = @import("httpz");
const palette_api = @import("palette_themify_api").palette_api;

const DEFAULT_PORT: u16 = 8089;
const DEFAULT_MAX_BODY_SIZE_MB: usize = 50;

const ApiHandler = struct {
    pub fn notFound(_: *ApiHandler, _: *httpz.Request, res: *httpz.Response) !void {
        res.status = 404;
        res.content_type = .JSON;
        res.body = "{\"error\":\"Not found\"}";
    }

    pub fn uncaughtError(_: *ApiHandler, req: *httpz.Request, res: *httpz.Response, err: anyerror) void {
        std.log.err("Unhandled API error for {s}: {}", .{ req.url.path, err });
        res.status = 500;
        res.content_type = .JSON;
        res.body = "{\"error\":\"Internal server error\"}";
    }
};

fn health(_: *ApiHandler, _: *httpz.Request, res: *httpz.Response) !void {
    res.content_type = .JSON;
    res.body = "{\"status\":\"ok\"}";
}

fn extractPalette(_: *ApiHandler, req: *httpz.Request, res: *httpz.Response) !void {
    res.content_type = .JSON;

    const body = req.body() orelse {
        res.status = 400;
        res.body = "{\"error\":\"Missing request body\"}";
        return;
    };

    if (body.len == 0) {
        res.status = 400;
        res.body = "{\"error\":\"Request body is empty\"}";
        return;
    }

    const result = palette_api.handleExtractPalette(req.arena, body) catch |err| {
        res.status = 400;
        res.body = switch (err) {
            error.InvalidContentType => "{\"error\":\"Invalid content type, expected multipart/form-data  \"}",
            error.NoImageProvided => "{\"error\":\"No image file provided in request\"}",
            error.NotAnImage => "{\"error\":\"Could not decode image - unsupported format\"}",
            error.NoColors => "{\"error\":\"No colors found in image\"}",
            else => "{\"error\":\"Internal server error\"}",
        };
        return;
    };

    res.body = result;
}

fn generateTheme(_: *ApiHandler, req: *httpz.Request, res: *httpz.Response) !void {
    res.content_type = .JSON;

    const body = req.body() orelse {
        res.status = 400;
        res.body = "{\"error\":\"Missing request body\"}";
        return;
    };

    const result = palette_api.handleGenerateTheme(req.arena, body) catch |err| {
        res.status = 400;
        res.body = switch (err) {
            error.NoColors => "{\"error\":\"Not enough colors - please provide at least 5 colors\"}",
            else => std.fmt.allocPrint(
                req.arena,
                "{{\"error\":\"Failed to generate theme error is {s}\"}}",
                .{@errorName(err)},
            ) catch "{\"error\":\"Internal server error\"}",
        };
        return;
    };

    res.body = result;
}

fn generateOverridable(_: *ApiHandler, req: *httpz.Request, res: *httpz.Response) !void {
    res.content_type = .JSON;

    const body = req.body() orelse {
        res.status = 400;
        res.body = "{\"error\":\"Missing request body\"}";
        return;
    };

    const result = palette_api.handleGenerateOverridable(req.arena, body) catch |err| {
        res.status = 400;
        res.body = switch (err) {
            error.InvalidTheme => "{\"error\":\"Please provide a valid Zed or VSCode theme JSON\"}",
            else => std.fmt.allocPrint(
                req.arena,
                "{{\"error\":\"Failed to generate theme error is {s}\"}}",
                .{@errorName(err)},
            ) catch "{\"error\":\"Internal server error\"}",
        };
        return;
    };

    res.body = result;
}

pub fn main(init: std.process.Init) !void {
    const allocator = std.heap.smp_allocator;
    const port = try environmentUnsigned(u16, init.environ_map, "PORT", DEFAULT_PORT);
    const max_body_size_mb = try environmentUnsigned(usize, init.environ_map, "MAX_BODY_SIZE_MB", DEFAULT_MAX_BODY_SIZE_MB);
    if (max_body_size_mb == 0 or max_body_size_mb > 1024) return error.InvalidMaxBodySize;
    const max_body_size = max_body_size_mb * 1024 * 1024;
    const cors_origin = init.environ_map.get("ALLOWED_ORIGIN") orelse "*";

    std.log.info("Zig Palette API starting on port {d}", .{port});

    var handler = ApiHandler{};
    var server = try httpz.Server(*ApiHandler).init(init.io, allocator, .{
        .address = .all(port),
        .request = .{ .max_body_size = max_body_size },
    }, &handler);
    defer server.stop();
    defer server.deinit();

    const cors = try server.middleware(httpz.middleware.Cors, .{
        .origin = cors_origin,
        .methods = "GET, POST, PUT, DELETE, OPTIONS",
        .headers = "Content-Type, Authorization, Accept",
        .max_age = "86400",
    });

    var router = try server.router(.{ .middlewares = &.{cors} });
    router.get("/health", health, .{});
    router.post("/extract-palette", extractPalette, .{});
    router.post("/generate-theme", generateTheme, .{});
    router.post("/generate-overridable", generateOverridable, .{});

    try server.listen();
}

fn environmentUnsigned(
    comptime T: type,
    environment: *const std.process.Environ.Map,
    name: []const u8,
    fallback: T,
) !T {
    const raw_value = environment.get(name) orelse return fallback;
    return std.fmt.parseUnsigned(T, raw_value, 10) catch error.InvalidEnvironmentVariable;
}
