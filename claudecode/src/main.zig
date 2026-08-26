const std = @import("std");

const httpz = @import("httpz");

const APP_NAME = "zigclaude";
const CREDS_FILE = "creds.json";

const Creds = struct { token: []const u8, refresh: []const u8, expires: std.time.epoch };

const App = struct {
    credsDir: std.Io.Dir,
    gpa: std.mem.Allocator,
    io: std.Io,
    creds: ?Creds,

    fn init(gpa: std.mem.Allocator, io: std.Io, home: []const u8) !App {
        const path = std.Io.Dir.path;
        const homeDir = try std.Io.Dir.openDirAbsolute(io, home, std.Io.Dir.OpenOptions{});

        const credsPath = try path.join(gpa, &.{ ".local", "state", APP_NAME });
        try std.Io.Dir.createDirPath(homeDir, io, credsPath);

        const dir = try homeDir.openDir(io, credsPath, std.Io.Dir.OpenOptions{
            .follow_symlinks = false,
        });

        return App{
            .creds = null,
            .credsDir = dir,
            .io = io,
            .gpa = gpa,
        };
    }

    fn readCreds(self: App) !void {
        // const file = try self.credsDir.openFile(self.io, CREDS_FILE, std.Io.Dir.OpenFileOptions{ .mode = .read_write });
        const contents = try self.credsDir.readFileAlloc(self.io, CREDS_FILE, 1024 * 1024);
        const creds = try std.json.parseFromSlice(Creds, self.gpa, contents, std.json.ParseOptions{});
        self.creds = creds;
    }

    // fn writeCreds(self: App) !void {}

    fn refreshCreds() void {}
};

pub fn main(init: std.process.Init) !void {
    const allocator = init.gpa;
    const addr = httpz.Config.Address.localhost(5882);

    const home = try init.minimal.environ.getAlloc(init.gpa, "HOME");
    var app = try App.init(init.gpa, init.io, home);

    // More advanced cases will use a custom "Handler" instead of "void".
    // The last parameter is our handler instance; since we have a "void"
    // handler, we passed a void ({}) value.
    var server = try httpz.Server(*App).init(init.io, allocator, .{
        // use .all(5882) to bind to all interfaces, i.e. 0.0.0.0
        .address = addr,
    }, &app);
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

fn getUser(_: *App, req: *httpz.Request, res: *httpz.Response) !void {
    res.status = 200;
    try res.json(.{ .id = req.param("id").?, .name = "Teg" }, .{});
}
