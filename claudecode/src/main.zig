const std = @import("std");

const httpz = @import("httpz");

const c = @cImport({
    @cInclude("time.h");
});

const APP_NAME = "zigclaude";
const CREDS_FILE = "creds.json";

const AUTH_ENDPOINT = "https://claude.ai/oauth/authorize";
const client_id = "9d1c250a-e61b-44d9-88ed-5944d1962f5e";

const Creds = struct {
    token: []const u8,
    refresh: []const u8,
    expires: c_long,
};

fn base64UrlEncode(input: []const u8, output: []u8) []const u8 {
    const encoder = std.base64.Base64Encoder.init(std.base64.standard_alphabet_chars, null);
    return encoder.encode(output, input);
}

fn sha256(input: []const u8) [256 / 8]u8 {
    const Sha256 = std.crypto.hash.sha2.Sha256;
    var sha = Sha256.init(Sha256.Options{});
    sha.update(input);
    return sha.finalResult();
}

fn generateChallenge(verifier: []const u8, challenge: []u8) []const u8 {
    return base64UrlEncode(&sha256(verifier), challenge);
}

const App = struct {
    credsDir: std.Io.Dir,
    gpa: std.mem.Allocator,
    io: std.Io,
    creds: ?Creds,
    rand: std.Random,

    fn init(gpa: std.mem.Allocator, io: std.Io, home: []const u8) !App {
        const path = std.Io.Dir.path;
        const homeDir = try std.Io.Dir.openDirAbsolute(io, home, std.Io.Dir.OpenOptions{});

        const credsPath = try path.join(gpa, &.{ ".local", "state", APP_NAME });
        defer gpa.free(credsPath);

        try std.Io.Dir.createDirPath(homeDir, io, credsPath);

        const dir = try homeDir.openDir(io, credsPath, std.Io.Dir.OpenOptions{
            .follow_symlinks = false,
        });

        const rng_impl: std.Random.IoSource = .{ .io = io };
        const secureRand = rng_impl.interface();

        var app = App{
            .creds = null,
            .credsDir = dir,
            .io = io,
            .gpa = gpa,
            .rand = secureRand,
        };

        app.readCreds() catch {
            try app.authenticate();
        };

        std.debug.print("is authed?: {}\n", .{app.isAuthenticated()});
        if (!app.isAuthenticated()) {
            try app.authenticate();
        }

        return app;
    }

    fn generateVerifier(self: *App, verifier: []u8) []const u8 {
        var buf: [32]u8 = @splat(0);
        self.rand.bytes(&buf);
        return base64UrlEncode(&buf, verifier);
    }

    fn readCreds(self: *App) !void {
        const contents =
            try self.credsDir.readFileAlloc(self.io, CREDS_FILE, self.gpa, std.Io.Limit.unlimited);
        defer self.gpa.free(contents);

        const creds = try std.json.parseFromSlice(Creds, self.gpa, contents, std.json.ParseOptions{});
        defer self.gpa.free(contents);

        self.creds = creds.value;
    }

    fn isAuthenticated(self: *App) bool {
        if (self.creds == null) {
            return false;
        }

        const now = c.time(null);

        return (now < self.creds.?.expires);
        // if (self.creds.?.expires.secs) {
        //
        // }
    }

    pub fn openUrl(self: *App, url: []const u8) !void {
        const argv = switch (@import("builtin").os.tag) {
            .windows => &[_][]const u8{ "rundll32", "url.dll,FileProtocolHandler", url },
            .macos => &[_][]const u8{ "open", url },
            else => &[_][]const u8{ "xdg-open", url },
        };

        const result = try std.process.run(self.gpa, self.io, std.process.RunOptions{ .argv = argv });
        // std.debug.print("xdg-open: {s}\n", .{result.stderr});
        self.gpa.free(result.stderr);
        self.gpa.free(result.stdout);
    }

    fn authenticate(self: *App) !void {
        var verifier_buf: [64]u8 = undefined;
        const verifier = self.generateVerifier(&verifier_buf);

        // there is 100% someway to comptime magic yourself into knowing how big
        // this slice *should* be, but i am not good enough to know it
        var challenge_buf: [64]u8 = undefined;
        const challenge = generateChallenge(verifier, &challenge_buf);

        const state = self.generateVerifier(&verifier_buf);

        const redirect = "https://console.anthropic.com/oauth/code/callback";
        const scope = "org:create_api_key user:profile user:inference";

        const query_string = try std.fmt.allocPrint(
            self.gpa,
            "client_id={[id]s}&response_type=code&redirect_uri={[redirect]s}&" ++
                "scope={[scope]s}&state={[state]s}&code_challenge={[challenge]s}&" ++
                "code_challenge_method=S256",
            .{
                .id = client_id,
                .redirect = redirect,
                .scope = scope,
                .state = state,
                .challenge = challenge,
            },
        );
        defer self.gpa.free(query_string);

        const url = std.Uri{
            .scheme = "https",
            .host = .{ .raw = "claude.ai" },
            .path = .{ .raw = "/oauth/authorize" },
            .query = .{ .raw = query_string },
        };

        var buf: [1024]u8 = undefined;
        var fixed: std.Io.Writer = .fixed(&buf);
        try url.format(&fixed);
        const auth_url = fixed.buffered();

        var out_buf: [1024]u8 = undefined;
        var stdout_writer = std.Io.File.stdout().writer(self.io, &out_buf);
        var stdout = &stdout_writer.interface;

        try stdout.print("Visit the following url in your browser:\n{s}\n", .{auth_url});
        try stdout.flush();

        try self.openUrl(auth_url);

        try stdout.print("\n", .{});
        try stdout.print("Paste Auth Key Here: ", .{});
        try stdout.flush();

        var in_buf: [1024]u8 = undefined;
        var stdin_reader = std.Io.File.stdin().reader(self.io, &in_buf);
        const stdin = &stdin_reader.interface;

        const key = try stdin.takeDelimiterExclusive('\n');

        std.debug.print("we gyatt a key: '{s}'", .{key});
    }

    // fn writeCreds(self: App) !void {}
    //
    // fn refreshCreds() void {}
};

pub fn main(init: std.process.Init) !void {
    const allocator = init.gpa;
    const addr = httpz.Config.Address.localhost(5882);

    const home = try init.minimal.environ.getAlloc(init.gpa, "HOME");
    defer init.gpa.free(home);
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
