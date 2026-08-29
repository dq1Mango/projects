const std = @import("std");
const http = std.http;
const print = std.debug.print;

const httpz = @import("httpz");

const c = @cImport({
    @cInclude("time.h");
});

const APP_NAME = "zigclaude";
const CREDS_FILE = "creds.json";

const AUTH_ENDPOINT = "https://claude.ai/oauth/authorize";
const client_id = "9d1c250a-e61b-44d9-88ed-5944d1962f5e";
const oauth_endpoint = "https://platform.claude.com/v1/oauth/token";

// const default_refresh_expiry: usize = 2460053;

const Creds = struct {
    token: []const u8,
    refresh: []const u8,
    // bofa deez r in unix seconds
    expires: usize,
    refresh_expires: usize,
};

fn base64UrlEncode(input: []const u8, output: []u8) []const u8 {
    const encoder = std.base64.Base64Encoder.init(std.base64.url_safe_alphabet_chars, null);
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

const Orginization = struct {
    uuid: []const u8,
    name: []const u8,
};

const Account = struct {
    uuid: []const u8,
    email_address: []const u8,
};

const TokenResponse = struct {
    token_type: []const u8,
    access_token: []const u8,
    expires_in: usize,
    refresh_token: []const u8,
    scope: []const u8,
    token_uuid: []const u8,
    refresh_token_expires_in: usize,
    organization: Orginization,
    account: Account,
};

const AuthenticationError = error{
    InvalidToken,
    ServerError,
};

const App = struct {
    credsDir: std.Io.Dir,
    gpa: std.mem.Allocator,
    io: std.Io,
    creds: ?Creds,
    rand: std.Random,
    client: *http.Client,

    fn init(gpa: std.mem.Allocator, io: std.Io, home: []const u8, client: *std.http.Client) !App {
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
            .client = client,
        };

        app.readCreds() catch {};

        std.debug.print("is authed?: {}\n", .{app.isAuthenticated()});
        if (!app.isAuthenticated()) {
            app.authenticate() catch |e| {
                app.deinit();
                return e;
            };
        }

        return app;
    }

    fn deinit(self: *App) void {
        self.credsDir.close(self.io);
    }

    fn generateVerifier(self: *App, verifier: []u8) []const u8 {
        var buf: [32]u8 = @splat(0);
        self.rand.bytes(&buf);
        return base64UrlEncode(&buf, verifier);
    }

    fn readCreds(self: *App) !void {
        _ = try self.credsDir.statFile(self.io, CREDS_FILE, .{});

        const contents =
            try self.credsDir.readFileAlloc(self.io, CREDS_FILE, self.gpa, std.Io.Limit.unlimited);
        defer self.gpa.free(contents);

        const creds = try std.json.parseFromSlice(Creds, self.gpa, contents, std.json.ParseOptions{});
        defer self.gpa.free(contents);

        self.creds = creds.value;
    }

    fn writeCreds(self: *App, creds: Creds) !void {
        self.creds = creds;

        const formatter = std.json.fmt(creds, .{ .whitespace = .indent_2 });

        var contents: std.Io.Writer.Allocating = .init(self.gpa);
        defer contents.deinit();

        try contents.writer.print("{f}", .{formatter});

        const file = try self.credsDir.openFile(self.io, CREDS_FILE, .{ .lock = .exclusive });
        defer file.close(self.io);

        try file.writeStreamingAll(self.io, contents.written());
    }

    fn isAuthenticated(self: *App) bool {
        if (self.creds) |crds| {
            const now: usize = @intCast(c.time(null));
            return (now < crds.expires);
        } else {
            return false;
        }
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

        var state_buf: [64]u8 = undefined;
        const state = self.generateVerifier(&state_buf);

        print("verifier: {s}\nchallenge: {s}\nstate: {s}\n", .{ verifier, challenge, state });

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

        std.debug.print("we gyatt a key: '{s}'\n", .{key});

        var spliterator = std.mem.splitScalar(u8, key, '#');
        var parts: [2][]const u8 = undefined;

        var i: u8 = 0;
        while (i < 2) {
            const next = spliterator.next() orelse {
                return AuthenticationError.InvalidToken;
            };
            parts[i] = next;
            i += 1;
        }

        print("part 1: {s}; part 2: {s}\n", .{ parts[0], parts[1] });

        if (spliterator.next() != null) {
            return AuthenticationError.InvalidToken;
        }

        const AuthRequest = struct {
            grant_type: []const u8,
            client_id: []const u8,
            code: []const u8,
            state: []const u8,
            redirect_uri: []const u8,
            code_verifier: []const u8,
        };

        const formatter = std.json.fmt(AuthRequest{
            .grant_type = "authorization_code",
            .client_id = client_id,
            .code = parts[0],
            .state = parts[1],
            .redirect_uri = redirect,
            .code_verifier = verifier,
        }, .{});

        var payload: std.Io.Writer.Allocating = .init(self.gpa);
        defer payload.deinit();

        try payload.writer.print("{f}", .{formatter});

        print("heres our bodyody:\n{s}\n", .{payload.written()});

        const uri = try std.Uri.parse(oauth_endpoint);

        var response_buf: [65536]u8 = undefined;
        var w: std.Io.Writer = .fixed(&response_buf);

        const response = try self.client.fetch(
            std.http.Client.FetchOptions{
                .location = .{ .uri = uri },
                .method = .POST,
                .headers = .{
                    .content_type = .{ .override = "application/json" },
                    .user_agent = .{ .override = "axios/1.15.2" },
                },
                .payload = payload.written(),
                .response_writer = &w,
            },
        );

        // Occasionally, httpbin might time out, so we disregard cases
        // where the response status is not okay.
        if (response.status != .ok) {
            std.debug.print("got bad status: {d}\n", .{response.status});
            // return AuthenticationError.ServerError;
        }

        print("Body:\n{s}\n", .{w.buffered()});

        if (response.status != .ok) {
            return AuthenticationError.ServerError;
        }

        const parsed = try std.json.parseFromSlice(
            TokenResponse,
            self.gpa,
            w.buffered(),
            .{ .ignore_unknown_fields = true },
        );
        defer parsed.deinit();

        print("heey it worked: {any}", .{parsed.value});

        const now = c.time(null);
        const zig_now: usize = @intCast(now);

        print("deez r seconds right: ? {d}", .{now});

        const creds: Creds = .{
            .token = parsed.value.access_token,
            .refresh = parsed.value.refresh_token,
            .expires = zig_now + parsed.value.expires_in,
            .refresh_expires = zig_now + parsed.value.refresh_token_expires_in,
        };

        try self.writeCreds(creds);
    }

    fn refreshCreds(self: *App) !void {
        const body = .{
            .grant_type = "refresh_token",
            .client_id = client_id,
            .refresh_token = self.creds.refresh_token,
        };

        var formatter = std.json.fmt(body, .{});

        var payload: std.Io.Writer.Allocating = .init(self.gpa);
        defer payload.deinit();

        try formatter.format(&payload.writer);

        print("refreshing token: {s}", payload.written());

        var response_buf: [65536]u8 = undefined;
        var w: std.Io.Writer = .fixed(&response_buf);

        const response = try self.client.fetch(
            std.http.Client.FetchOptions{
                .location = .{ .url = oauth_endpoint },
                .method = .POST,
                .headers = .{
                    .content_type = .{ .override = "application/json" },
                    .user_agent = .{ .override = "axios/1.15.2" },
                },
                .payload = payload.written(),
                .response_writer = &w,
            },
        );

        if (response.status != .ok) {
            print("got bad status: {d}\n", .{response.status});
        }

        print("refresh response body: {s}\n", w.buffered());

        const parsed = try std.json.parseFromSlice(
            TokenResponse,
            self.gpa,
            w.buffered(),
            .{ .ignore_unknown_fields = true },
        );
        defer parsed.deinit();
        const value = parsed.value;

        const now: usize = @intCast(c.time(null));

        const creds: Creds = .{
            .token = value.access_token,
            .refresh = value.refresh_token,
            .expires = now + value.expires_in,
            .refresh_expires = now + value.refresh_token_expires_in,
        };

        try self.writeCreds(creds);
    }
};

const MainError = AuthenticationError || error{NoHome};

pub fn main(init: std.process.Init) !void {
    const allocator = init.gpa;
    const addr = httpz.Config.Address.localhost(5882);

    const home = init.environ_map.get("HOME") orelse {
        return MainError.NoHome;
    };

    var client: http.Client = .{ .allocator = allocator, .io = init.io };
    defer client.deinit();

    try client.initDefaultProxies(init.arena.allocator(), init.environ_map);

    var app = App.init(init.gpa, init.io, home, &client) catch |e| {
        switch (e) {
            AuthenticationError.InvalidToken, AuthenticationError.ServerError => {
                std.debug.print(
                    "welp we couldn't authenticate, heres the error that means nothing: {any}\n",
                    .{e},
                );
                std.debug.print("You are welcome to try again :)\n", .{});
                return;
            },
            else => {
                return e;
            },
        }
    };
    defer app.deinit();

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
    router.get("/", proxyAnthropic, .{});

    std.debug.print("listening on -- {f}\n", .{addr});

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

fn proxyAnthropic(_: *App, req: *httpz.Request, res: *httpz.Response) !void {
    res.status = 200;
    // req.headers.
    try res.json(.{ .id = req.param("id").?, .name = "Teg" }, .{});
}
