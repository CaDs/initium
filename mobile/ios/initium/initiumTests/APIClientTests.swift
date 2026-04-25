import Testing
import Foundation
@testable import initium

// APIClient is the concrete actor that owns: bearer attachment, JSON
// encode/decode, 401 → refresh → retry once, and single-flight coalescing
// of refresh under concurrent 401s. The state machine in AuthStore is
// already covered by AuthStoreTests; THIS file exercises the wire-level
// behavior with a URLProtocol stub so no real server is needed.
//
// Why URLProtocol: APIClient is constructed with `URLSession.shared`'s
// configuration by default. A custom `URLSessionConfiguration` with our
// stub class registered intercepts every URLRequest and replies with a
// canned (Data, HTTPURLResponse), letting tests script the response
// sequence (first 401, then refresh-200, then retry-200) deterministically.

// MARK: - URLProtocol stub

/// Drives a configurable script of HTTP responses keyed by request order.
/// Tests enqueue responses; the protocol pops one per request. Captures
/// every received URLRequest for post-hoc assertions.
final class StubURLProtocol: URLProtocol, @unchecked Sendable {
    /// Single shared queue; `setUp()` clears it. Concurrency-safe via lock.
    private static let lock = NSLock()
    private static var queue: [Response] = []
    private static var requests: [URLRequest] = []

    struct Response {
        let status: Int
        let body: Data
    }

    static func reset() {
        lock.lock(); defer { lock.unlock() }
        queue.removeAll()
        requests.removeAll()
    }

    static func enqueue(_ responses: Response...) {
        lock.lock(); defer { lock.unlock() }
        queue.append(contentsOf: responses)
    }

    static var receivedRequests: [URLRequest] {
        lock.lock(); defer { lock.unlock() }
        return requests
    }

    override class func canInit(with _: URLRequest) -> Bool { true }
    override class func canonicalRequest(for r: URLRequest) -> URLRequest { r }

    override func startLoading() {
        Self.lock.lock()
        // URLProtocol re-issues the request without a body when `httpBody` is
        // a stream; the original Data body lives on `httpBodyStream`. We
        // capture both forms so assertions can inspect either.
        var captured = request
        if captured.httpBody == nil, let stream = captured.httpBodyStream {
            captured.httpBody = StubURLProtocol.drain(stream)
        }
        Self.requests.append(captured)
        let response = Self.queue.isEmpty ? nil : Self.queue.removeFirst()
        Self.lock.unlock()

        guard let response else {
            client?.urlProtocol(self, didFailWithError: NSError(
                domain: "StubURLProtocol", code: -1,
                userInfo: [NSLocalizedDescriptionKey: "no enqueued response"]
            ))
            return
        }

        let http = HTTPURLResponse(
            url: request.url!, statusCode: response.status,
            httpVersion: "HTTP/1.1", headerFields: ["Content-Type": "application/json"]
        )!
        client?.urlProtocol(self, didReceive: http, cacheStoragePolicy: .notAllowed)
        client?.urlProtocol(self, didLoad: response.body)
        client?.urlProtocolDidFinishLoading(self)
    }

    override func stopLoading() {}

    private static func drain(_ stream: InputStream) -> Data {
        stream.open()
        defer { stream.close() }
        var data = Data()
        let buf = UnsafeMutablePointer<UInt8>.allocate(capacity: 4096)
        defer { buf.deallocate() }
        while stream.hasBytesAvailable {
            let n = stream.read(buf, maxLength: 4096)
            if n <= 0 { break }
            data.append(buf, count: n)
        }
        return data
    }
}

// MARK: - Helpers

private func stubbedSession() -> URLSession {
    let cfg = URLSessionConfiguration.ephemeral
    cfg.protocolClasses = [StubURLProtocol.self]
    return URLSession(configuration: cfg)
}

private let userJSON = #"""
{"id":"u-1","email":"u@example.com","name":"U","avatar_url":"","role":"user","created_at":"2026-01-01T00:00:00Z"}
"""#.data(using: .utf8)!

private let pairJSON = #"""
{"access_token":"NEW-ACCESS","refresh_token":"NEW-REFRESH"}
"""#.data(using: .utf8)!

private let errorJSON = #"""
{"code":"INVALID_CREDENTIALS","message":"unauthorized"}
"""#.data(using: .utf8)!

// MARK: - Tests

// Serialized: StubURLProtocol uses static (process-global) queue + capture
// state, which doesn't compose with Swift Testing's default parallel
// execution. The cost is small (each test is ~10ms) and the alternative
// — per-instance protocol classes — fights URLProtocol's API design.
@Suite(.serialized)
struct APIClientTests {

    @Test func meAttachesBearerHeaderFromTokenStorage() async throws {
        StubURLProtocol.reset()
        StubURLProtocol.enqueue(.init(status: 200, body: userJSON))

        let storage = MockTokenStorage(access: "ABC", refresh: "REF")
        let client = APIClient(
            baseURL: URL(string: "https://example.test")!,
            tokenStorage: storage,
            session: stubbedSession()
        )

        _ = try await client.me()

        let req = StubURLProtocol.receivedRequests[0]
        #expect(req.value(forHTTPHeaderField: "Authorization") == "Bearer ABC")
        #expect(req.url?.path == "/api/me")
    }

    @Test func requestMagicLinkSkipsAuth() async throws {
        StubURLProtocol.reset()
        StubURLProtocol.enqueue(.init(status: 200, body: #"{"message":"sent"}"#.data(using: .utf8)!))

        let storage = MockTokenStorage(access: "ABC", refresh: "REF")
        let client = APIClient(
            baseURL: URL(string: "https://example.test")!,
            tokenStorage: storage,
            session: stubbedSession()
        )

        _ = try await client.requestMagicLink(email: "u@example.com")

        let req = StubURLProtocol.receivedRequests[0]
        // Magic link request has no session yet — bearer must NOT attach.
        #expect(req.value(forHTTPHeaderField: "Authorization") == nil)
    }

    @Test func unauthorized_thenRefresh_thenRetrySucceeds() async throws {
        StubURLProtocol.reset()
        // 1) GET /me with OLD → 401
        // 2) POST /api/auth/refresh → 200 with new pair
        // 3) GET /me retried with NEW → 200
        StubURLProtocol.enqueue(
            .init(status: 401, body: errorJSON),
            .init(status: 200, body: pairJSON),
            .init(status: 200, body: userJSON)
        )

        let storage = MockTokenStorage(access: "OLD", refresh: "REFRESH")
        let client = APIClient(
            baseURL: URL(string: "https://example.test")!,
            tokenStorage: storage,
            session: stubbedSession()
        )

        let user = try await client.me()
        #expect(user.email == "u@example.com")

        let reqs = StubURLProtocol.receivedRequests
        #expect(reqs.count == 3)
        #expect(reqs[0].value(forHTTPHeaderField: "Authorization") == "Bearer OLD")
        #expect(reqs[1].url?.path == "/api/auth/refresh")
        // After refresh succeeded, storage holds new tokens AND the retry
        // sent the new bearer.
        #expect(storage.accessToken() == "NEW-ACCESS")
        #expect(storage.refreshToken() == "NEW-REFRESH")
        #expect(reqs[2].value(forHTTPHeaderField: "Authorization") == "Bearer NEW-ACCESS")
    }

    @Test func refreshFailure_throwsUnauthorized_andFiresOnUnauthorized() async throws {
        StubURLProtocol.reset()
        // 1) GET /me with OLD → 401
        // 2) POST /api/auth/refresh → 401 (refresh itself rejected)
        StubURLProtocol.enqueue(
            .init(status: 401, body: errorJSON),
            .init(status: 401, body: errorJSON)
        )

        let storage = MockTokenStorage(access: "OLD", refresh: "REFRESH")

        // Capture the callback. APIClient's `onUnauthorized` is `@Sendable`,
        // so we use an actor-isolated counter.
        actor Counter { var n = 0; func inc() { n += 1 } }
        let counter = Counter()

        let client = APIClient(
            baseURL: URL(string: "https://example.test")!,
            tokenStorage: storage,
            session: stubbedSession(),
            onUnauthorized: { Task { await counter.inc() } }
        )

        do {
            _ = try await client.me()
            Issue.record("expected APIClient.Error.unauthorized")
        } catch APIClient.Error.unauthorized {
            // ok
        } catch {
            Issue.record("wrong error: \(error)")
        }

        // Refresh's internal Task fires onUnauthorized via Task { ... }; give
        // the runtime one tick to drain it before asserting.
        try? await Task.sleep(nanoseconds: 50_000_000)
        let calls = await counter.n
        #expect(calls == 1)
        // Tokens cleared so the app's auth state machine drops to login.
        #expect(storage.accessToken() == nil)
    }

    @Test func concurrentUnauthorized_coalescesIntoSingleRefresh() async throws {
        StubURLProtocol.reset()
        // Five concurrent /me calls each hit 401. The single-flight refresh
        // actor must collapse them into ONE /api/auth/refresh call. Then
        // each retried /me succeeds with the new bearer.
        //
        // Sequence (order is enforced by URLProtocol queue, not by the
        // client — APIClient's actor serializes refresh, which the queue
        // matches up by happenstance: 5×401 then 1×refresh-200 then 5×200).
        StubURLProtocol.enqueue(
            .init(status: 401, body: errorJSON),
            .init(status: 401, body: errorJSON),
            .init(status: 401, body: errorJSON),
            .init(status: 401, body: errorJSON),
            .init(status: 401, body: errorJSON),
            .init(status: 200, body: pairJSON),
            .init(status: 200, body: userJSON),
            .init(status: 200, body: userJSON),
            .init(status: 200, body: userJSON),
            .init(status: 200, body: userJSON),
            .init(status: 200, body: userJSON)
        )

        let storage = MockTokenStorage(access: "OLD", refresh: "REFRESH")
        let client = APIClient(
            baseURL: URL(string: "https://example.test")!,
            tokenStorage: storage,
            session: stubbedSession()
        )

        try await withThrowingTaskGroup(of: User.self) { group in
            for _ in 0..<5 {
                group.addTask { try await client.me() }
            }
            for try await _ in group {}
        }

        // 5 initial /me + 1 refresh + 5 retries = 11 total. If single-flight
        // broke and each 401 triggered its own refresh, count would be 15+.
        let reqs = StubURLProtocol.receivedRequests
        #expect(reqs.count == 11)

        let refreshCalls = reqs.filter { $0.url?.path == "/api/auth/refresh" }.count
        #expect(refreshCalls == 1, "single-flight refresh must collapse N concurrent 401s into 1 refresh; got \(refreshCalls)")
    }

    @Test func httpErrorEnvelopeSurfacesAsAPIClientError() async throws {
        StubURLProtocol.reset()
        let usedTokenJSON = #"{"code":"TOKEN_USED","message":"token already used"}"#.data(using: .utf8)!
        StubURLProtocol.enqueue(.init(status: 409, body: usedTokenJSON))

        let storage = MockTokenStorage()
        let client = APIClient(
            baseURL: URL(string: "https://example.test")!,
            tokenStorage: storage,
            session: stubbedSession()
        )

        do {
            _ = try await client.verifyMagicLink(token: "used")
            Issue.record("expected http error")
        } catch APIClient.Error.http(let status, let envelope) {
            #expect(status == 409)
            #expect(envelope?.code == "TOKEN_USED")
        } catch {
            Issue.record("wrong error: \(error)")
        }
    }
}
