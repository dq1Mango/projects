const std = @import("std");
const http = std.http;
const print = std.debug.print;
const log = std.log;
const mem = std.mem;
const Io = std.Io;
const net = Io.net;
const awake = Io.Clock.awake;
const IpAddress = net.IpAddress;

// const httpz = @import("httpz");

const APP_NAME = "zigclaude";
const CREDS_FILE = "creds.json";

// const AUTH_ENDPOINT = "https://claude.ai/oauth/authorize";
const client_id = "9d1c250a-e61b-44d9-88ed-5944d1962f5e";
const oauth_endpoint = "https://platform.claude.com/v1/oauth/token";
const token_endpoint = "https://api.anthropic.com";

// const default_refresh_expiry: usize = 2460053;

const Creds = struct {
    mewing: std.Io.Mutex = .init,

    token: []const u8,
    refresh: []const u8,
    // bofa deez r in unix seconds
    expires: Io.Timestamp,
    refresh_expires: Io.Timestamp,
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
    expires_in: i64,
    refresh_token: []const u8,
    scope: []const u8,
    token_uuid: []const u8,
    refresh_token_expires_in: i64,
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
    rand: std.Random,
    client: *http.Client,

    creds: ?Creds,
    refreshTask: ?Io.Future(anyerror!void),

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
            .refreshTask = null,
        };

        app.readCreds() catch {};

        std.debug.print("is authed?: {}\n", .{app.isAuthenticated()});
        if (!app.isAuthenticated()) {
            app.authenticate() catch |e| {
                app.deinit();
                return e;
            };
        }

        const refreshTask = try io.concurrent(App.continualRefresh, .{&app});
        app.refreshTask = refreshTask;

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
            const now = awake.now(self.io);
            return (now.toSeconds() < crds.expires.toSeconds());
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

        const now = awake.now(self.io);

        print("deez r seconds right: ? {d}", .{now});

        const creds: Creds = .{
            .token = parsed.value.access_token,
            .refresh = parsed.value.refresh_token,
            .expires = now.addDuration(
                Io.Duration.fromSeconds(parsed.value.expires_in),
            ),
            .refresh_expires = now.addDuration(
                Io.Duration.fromSeconds(parsed.value.refresh_token_expires_in),
            ),
        };

        try self.writeCreds(creds);
    }

    const RefreshError = error{NoExistingRefreshToken};

    fn refreshCreds(self: *App) !void {
        const refreshToken = self.creds orelse {
            return RefreshError.NoExistingRefreshToken;
        };
        const body = .{
            .grant_type = "refresh_token",
            .client_id = client_id,
            .refresh_token = refreshToken,
        };

        var formatter = std.json.fmt(body, .{});

        var payload: std.Io.Writer.Allocating = .init(self.gpa);
        defer payload.deinit();

        try formatter.format(&payload.writer);

        print("refreshing token: {s}", .{payload.written()});

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

        print("refresh response body: {s}\n", .{w.buffered()});

        const parsed = try std.json.parseFromSlice(
            TokenResponse,
            self.gpa,
            w.buffered(),
            .{ .ignore_unknown_fields = true },
        );
        defer parsed.deinit();
        const value = parsed.value;

        const now = awake.now(self.io);

        const creds: Creds = .{
            .token = value.access_token,
            .refresh = value.refresh_token,
            .expires = now.addDuration(.fromSeconds(value.expires_in)),
            .refresh_expires = now.addDuration(.fromSeconds(value.refresh_token_expires_in)),
        };

        try self.writeCreds(creds);
    }

    pub fn continualRefresh(app: *App) anyerror!void {
        const grace_time = 10 * 60;
        while (true) {
            const now = awake.now(app.io);
            try app.io.sleep(.fromSeconds(
                app.creds.?.expires.toSeconds() - now.toSeconds() - grace_time,
            ), .real);

            try app.refreshCreds();
        }
    }
};

const MainError = AuthenticationError || error{NoHome};

pub fn main(init: std.process.Init) !void {
    const allocator = init.gpa;
    // const addr = httpz.Config.Address.localhost(5882);

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

    // var server = try httpz.Server(*App).init(init.io, allocator, .{
    //     // use .all(5882) to bind to all interfaces, i.e. 0.0.0.0
    //     .address = addr,
    // }, &app);
    // defer {
    //     // clean shutdown, finishes serving any live requests
    //     server.stop();
    //     server.deinit();
    // }
    //
    //
    // var router = try server.router(.{});
    // router.get("/", proxyAnthropic, .{});
    //
    // std.debug.print("listening on -- {f}\n", .{addr});

    // try stdout.print("listening on -- ", .{});
    // try addr.format(stdout);
    // try stdout.print("\n", .{});
    // try stdout.flush();
    // addr.format(stdout.writer(io: Io, buffer: []u8));
    // std.debug.print("Listening on {any}\n", .{addr});

    // try stdout.writeStreamingAll(init.io, "Listening on {}", addr);

    // blocks
    // try server.listen();
    //
    const listen_port = 1234;

    const addr = try IpAddress.parseIp4("0.0.0.0", listen_port);
    var listener = try addr.listen(init.io, .{ .reuse_address = true });
    defer listener.deinit(init.io);

    var connections: Io.Group = .init;
    defer connections.cancel(init.io);

    print("proxy listening on http://0.0.0.0:{d}", .{listen_port});
    print("forwarding to https://{s}", .{token_endpoint});
    print("injecting Authorization: {s}", .{app.creds.?.token});

    while (true) {
        const stream = try listener.accept(init.io);
        connections.concurrent(
            init.io,
            handleConnection,
            .{ init.io, allocator, stream, &app.creds.? },
        ) catch |err| {
            print("connection error: {any}", .{err});
        };
    }
}

fn handleConnection(
    io: Io,
    allocator: mem.Allocator,
    stream: net.Stream,
    creds: *Creds,
) !void {
    defer stream.close(io);

    // Buffers for the incoming client connection.
    var in_buf: [65536]u8 = undefined;
    var out_buf: [8192]u8 = undefined;

    var stream_reader = net.Stream.Reader.init(stream, io, &in_buf);
    var stream_writer = net.Stream.Writer.init(stream, io, &out_buf);

    var server = http.Server.init(&stream_reader.interface, &stream_writer.interface);

    // Handle potentially multiple requests on the same connection (keep-alive).
    while (true) {
        var request = server.receiveHead() catch |err| switch (err) {
            error.HttpConnectionClosing => return,
            else => {
                print("failed to receive request head: {any}", .{err});
                return;
            },
        };

        proxyRequest(io, allocator, &request, creds) catch |err| {
            print("proxy error for {s} {s}: {any}", .{
                @tagName(request.head.method),
                request.head.target,
                err,
            });
            // Try to send an error response back to the client.
            request.respond("502 Bad Gateway\n", .{
                .status = .bad_gateway,
                .extra_headers = &.{
                    .{ .name = "content-type", .value = "text/plain" },
                },
            }) catch return;
        };
    }
}

fn proxyRequest(
    io: Io,
    gpa: mem.Allocator,
    request: *http.Server.Request,
    creds: *Creds,
) !void {
    log.info("{s} {s}", .{ @tagName(request.head.method), request.head.target });

    // ── 1. Collect incoming headers (skip hop-by-hop & authorization) ──

    // We'll build the list of extra headers for the upstream request.
    // Worst case we forward all incoming headers + 1 for Authorization.
    var header_list = try std.ArrayList(http.Header).initCapacity(gpa, 16);
    defer header_list.deinit(gpa);

    var it = request.iterateHeaders();
    while (it.next()) |hdr| {
        // Skip hop-by-hop headers that shouldn't be forwarded.
        if (std.ascii.eqlIgnoreCase(hdr.name, "connection")) continue;
        if (std.ascii.eqlIgnoreCase(hdr.name, "keep-alive")) continue;
        if (std.ascii.eqlIgnoreCase(hdr.name, "proxy-authorization")) continue;
        if (std.ascii.eqlIgnoreCase(hdr.name, "proxy-connection")) continue;
        if (std.ascii.eqlIgnoreCase(hdr.name, "te")) continue;
        if (std.ascii.eqlIgnoreCase(hdr.name, "transfer-encoding")) continue;
        if (std.ascii.eqlIgnoreCase(hdr.name, "upgrade")) continue;

        // Strip any existing Authorization – we'll add our own.
        if (std.ascii.eqlIgnoreCase(hdr.name, "authorization")) continue;

        // Skip headers the client library manages itself.
        if (std.ascii.eqlIgnoreCase(hdr.name, "host")) continue;
        if (std.ascii.eqlIgnoreCase(hdr.name, "user-agent")) continue;
        if (std.ascii.eqlIgnoreCase(hdr.name, "accept-encoding")) continue;
        if (std.ascii.eqlIgnoreCase(hdr.name, "content-length")) continue;

        try header_list.append(gpa, .{ .name = hdr.name, .value = hdr.value });
    }

    // ── 2. Read the request body (if any) ──

    var body: ?[]const u8 = null;
    defer if (body) |b| gpa.free(b);

    if (request.head.method.requestHasBody()) {
        if (request.head.content_length) |len| {
            if (len > 0 and len <= 10 * 1024 * 1024) { // 10 MB limit
                var body_buf: [4096]u8 = undefined;
                var reader = request.readerExpectNone(&body_buf);

                var collected = try std.ArrayList(u8).initCapacity(gpa, len);
                defer collected.deinit(gpa);

                var tmp: [8192]u8 = undefined;
                while (true) {
                    const n = reader.readSliceShort(&tmp) catch {
                        break;
                    };
                    if (n == 0) break;
                    try collected.appendSlice(gpa, tmp[0..n]);
                }

                body = try collected.toOwnedSlice(gpa);
            }
        }
    }

    // ── 3. Build the upstream URI ──

    // The target from the client is typically just a path like "/api/data".
    // We construct a full URI pointing at the upstream.
    var uri_buf: [8192]u8 = undefined;
    const uri_str = try std.fmt.bufPrint(&uri_buf, "{s}{s}", .{
        token_endpoint,
        request.head.target,
    });
    const uri = try std.Uri.parse(uri_str);

    // ── 4. Make the upstream request ──

    var client = http.Client{
        .allocator = gpa,
        .io = io,
    };
    defer client.deinit();

    try creds.mewing.lock(io);

    const auth_header = std.fmt.allocPrint(
        gpa,
        "Authorization {s}",
        .{creds.token},
    ) catch |err| {
        creds.mewing.unlock(io);
        return err;
    };

    creds.mewing.unlock(io);

    var upstream_req = try client.request(request.head.method, uri, .{
        .redirect_behavior = .unhandled,
        .keep_alive = false,
        .headers = .{
            // Let the client library set the default host from the URI.
            .host = .default,
            // Inject our Authorization header via the override mechanism.
            .authorization = .{ .override = auth_header },
            .user_agent = .default,
            .connection = .default,
            .accept_encoding = .default,
            .content_type = if (request.head.content_type) |ct|
                .{ .override = ct }
            else
                .default,
        },
        .extra_headers = header_list.items,
    });
    defer upstream_req.deinit();

    // Send the request (with body if present).
    if (body) |b| {
        upstream_req.transfer_encoding = .{ .content_length = b.len };
        var bw = try upstream_req.sendBodyUnflushed(&.{});
        bw.writer.end = b.len;
        try bw.end();
        try upstream_req.connection.?.flush();
    } else {
        try upstream_req.sendBodiless();
    }

    // ── 5. Receive the upstream response head ──

    var redirect_buf: [1]u8 = undefined;
    var response = try upstream_req.receiveHead(&redirect_buf);

    // ── 6. Collect upstream response headers to forward back ──

    var resp_headers = try std.ArrayList(http.Header).initCapacity(gpa, 16);
    defer resp_headers.deinit(gpa);

    var resp_it = response.head.iterateHeaders();
    while (resp_it.next()) |hdr| {
        // Skip hop-by-hop and headers the server library manages.
        if (std.ascii.eqlIgnoreCase(hdr.name, "connection")) continue;
        if (std.ascii.eqlIgnoreCase(hdr.name, "keep-alive")) continue;
        if (std.ascii.eqlIgnoreCase(hdr.name, "transfer-encoding")) continue;
        if (std.ascii.eqlIgnoreCase(hdr.name, "content-length")) continue;

        try resp_headers.append(gpa, .{ .name = hdr.name, .value = hdr.value });
    }

    // ── 7. Stream the upstream response body back to the client ──
    //
    // Use respondStreaming so bytes flow through without buffering the
    // entire body in memory.  If the upstream provided a content-length
    // we forward it so the client knows the size up front; otherwise
    // the server library falls back to chunked transfer-encoding.

    var stream_buf: [16384]u8 = undefined;
    var body_writer = try request.respondStreaming(&stream_buf, .{
        .content_length = response.head.content_length,
        .respond_options = .{
            .status = response.head.status,
            .reason = response.head.reason,
            .extra_headers = resp_headers.items,
        },
    });

    var transfer_buf: [16384]u8 = undefined;
    var body_reader = response.reader(&transfer_buf);

    var total_bytes: u64 = 0;
    var read_buf: [16384]u8 = undefined;
    while (true) {
        const n = body_reader.readSliceShort(&read_buf) catch {
            break;
        };
        if (n == 0) break;
        try body_writer.writer.writeAll(read_buf[0..n]);
        total_bytes += n;
    }

    try body_writer.end();
    try request.server.out.flush();

    log.info("  -> {d} {s} ({d} bytes streamed)", .{
        @intFromEnum(response.head.status),
        response.head.reason,
        total_bytes,
    });
}
