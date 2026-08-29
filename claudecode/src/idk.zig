const std = @import("std");
const print = std.debug.print;
const http = std.http;

// pub fn main(init: std.process.Init) !void {
//     const gpa = init.gpa;
//     const io = init.io;
//
//     var client: http.Client = .{ .allocator = gpa, .io = io };
//     defer client.deinit();
//     try client.initDefaultProxies(init.arena.child_allocator, init.environ_map);
//
//     const uri = try std.Uri.parse("https://example.com");
//
//     var req = try client.request(.POST, uri, .{
//         .extra_headers = &.{.{ .name = "Content-Type", .value = "application/json" }},
//     });
//     defer req.deinit();
//
//     var payload: [7]u8 = "[1,2,3]".*;
//     try req.sendBodyComplete(&payload);
//     var buf: [1024]u8 = undefined;
//     var response = try req.receiveHead(&buf);
//
//     // Occasionally, httpbin might time out, so we disregard cases
//     // where the response status is not okay.
//     if (response.head.status != .ok) {
//         std.debug.print("got bad status: {d}\n", .{response.head.status});
//     }
//
//     const body = try response.reader(&.{}).allocRemaining(gpa, .unlimited);
//     defer gpa.free(body);
//     print("Body:\n{s}\n", .{body});
// }
//
//

const Orginization = struct {
    uuid: []const u8,
    name: []const u8,
};

const Account = struct {
    uuid: []const u8,
    email_address: []const u8,
};

const AuthenticationRespons = struct {
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
pub fn main(init: std.process.Init) !void {
    const body =
        \\{"token_type": "Bearer", "access_token": "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", "expires_in": 28800, "refresh_token": "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", "scope": "user:inference user:profile", "token_uuid": "efc600cf-e461-4d9f-be8f-f1b06e4a56fd", "refresh_token_expires_in": 2460053, "organization": {"uuid": "dc5b47ad-fdb0-407f-98f5-5020aaadbcbb", "name": "Voxer"}, "account": {"uuid": "6c39d191-ccd2-4895-ac69-436b4479ef7d", "email_address": "thomas.ranney@voxer.com"}}
    ;

    const auth = try std.json.parseFromSlice(
        AuthenticationRespons,
        init.gpa,
        body,
        .{ .ignore_unknown_fields = true },
    );

    defer auth.deinit();

    print("heey it worked: {}", .{auth.value});
}
