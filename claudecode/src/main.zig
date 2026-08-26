const std = @import("std");
const httpz = @import("httpz");

pub fn main(init: std.process.Init) !void {
    const allocator = init.gpa;
    const addr = httpz.Config.Address.localhost(5882);

    // More advanced cases will use a custom "Handler" instead of "void".
    // The last parameter is our handler instance; since we have a "void"
    // handler, we passed a void ({}) value.
    var server = try httpz.Server(void).init(init.io, allocator, .{
        // use .all(5882) to bind to all interfaces, i.e. 0.0.0.0
        .address = addr,
    }, {});
    defer {
        // clean shutdown, finishes serving any live requests
        server.stop();
        server.deinit();
    }

    var router = try server.router(.{});
    router.get("/api/user/:id", getUser, .{});

    var buf: [16]u8 = undefined;
    // var w = std.Io.File.stdout().writer(init.io, &buf);
    // const stdout: *std.Io.Writer = &w.interface;
    // try std.Io.File.stdout().writeStreamingAll(init.io, &buf);
    //
    var fx: std.Io.Writer = .fixed(&buf);
    try addr.format(&fx);

    std.debug.print("listening on -- {s}\n", .{fx.buffered()});

    // try stdout.print("listening on -- ", .{});
    // try addr.format(stdout);
    // try stdout.print("\n", .{});
    // try stdout.flush();
    // addr.format(stdout.writer(io: Io, buffer: []u8));
    // std.debug.print("Listening on {any}\n", .{addr});

    // try stdout.writeStreamingAll(init.io, "Listening on {}", addr);

    // blocks
    try server.listen();
}

fn getUser(req: *httpz.Request, res: *httpz.Response) !void {
    res.status = 200;
    try res.json(.{ .id = req.param("id").?, .name = "Teg" }, .{});
}
