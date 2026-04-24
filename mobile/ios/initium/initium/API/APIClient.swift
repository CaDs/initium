import Foundation

/// Thin hand-written API client for the auth + profile endpoints.
///
/// Concerns:
///  - Attaches `Authorization: Bearer <access>` to every request except
///    those that opt out via the `skipAuth` argument (login / magic link
///    request / refresh itself).
///  - On a 401 response, the single-flight [refreshActor] swaps in a
///    fresh token pair from `/api/auth/refresh` and retries the original
///    request once. Concurrent requests racing a 401 share the same
///    refresh call.
///  - JSON uses a single shared `JSONDecoder`/`JSONEncoder`. No
///    custom wire conventions beyond snake_case handled per-model via
///    explicit `CodingKeys`.
///
/// Instances are cheap — create one at app launch, pass it through the
/// view tree via SwiftUI's `Environment`. Don't construct per-call.
actor APIClient {

    enum Error: Swift.Error, Sendable {
        case network(underlying: Swift.Error)
        case http(status: Int, envelope: ErrorResponse?)
        case decode(message: String)
        case unauthorized
    }

    private let baseURL: URL
    private let session: URLSession
    private let tokenStorage: TokenStorage
    private let onUnauthorized: @Sendable () -> Void

    /// Single-flight refresh guard — only one refresh runs at a time.
    /// Concurrent callers reuse the in-flight task's result.
    private var refreshInFlight: Task<TokenPair?, Never>?

    init(
        baseURL: URL,
        tokenStorage: TokenStorage,
        session: URLSession = .shared,
        onUnauthorized: @Sendable @escaping () -> Void = {}
    ) {
        self.baseURL = baseURL
        self.session = session
        self.tokenStorage = tokenStorage
        self.onUnauthorized = onUnauthorized
    }

    // MARK: - Public API (one method per endpoint)

    func requestMagicLink(email: String) async throws -> MessageResponse {
        try await post("/api/auth/magic-link", body: MagicLinkRequest(email: email), skipAuth: true)
    }

    func verifyMagicLink(token: String) async throws -> TokenPair {
        let pair: TokenPair = try await post(
            "/api/auth/mobile/verify",
            body: MobileVerifyRequest(token: token),
            skipAuth: true
        )
        tokenStorage.save(tokens: pair)
        return pair
    }

    func verifyGoogleIDToken(_ idToken: String) async throws -> TokenPair {
        let pair: TokenPair = try await post(
            "/api/auth/mobile/google",
            body: MobileGoogleRequest(idToken: idToken),
            skipAuth: true
        )
        tokenStorage.save(tokens: pair)
        return pair
    }

    func me() async throws -> User {
        try await get("/api/me")
    }

    /// Best-effort logout: POSTs to the backend to revoke the session
    /// server-side, then clears local tokens regardless of the server
    /// response.
    func logout() async {
        _ = try? await post("/api/auth/logout", body: EmptyBody(), as: MessageResponse.self)
        tokenStorage.clear()
    }

    // MARK: - Internal plumbing

    private struct EmptyBody: Encodable {}

    private func get<T: Decodable>(_ path: String) async throws -> T {
        var request = URLRequest(url: baseURL.appendingPathComponent(path.trimmingLeading("/")))
        request.httpMethod = "GET"
        return try await send(request, skipAuth: false)
    }

    private func post<B: Encodable, T: Decodable>(
        _ path: String,
        body: B,
        skipAuth: Bool = false
    ) async throws -> T {
        try await post(path, body: body, as: T.self, skipAuth: skipAuth)
    }

    private func post<B: Encodable, T: Decodable>(
        _ path: String,
        body: B,
        as _: T.Type,
        skipAuth: Bool = false
    ) async throws -> T {
        var request = URLRequest(url: baseURL.appendingPathComponent(path.trimmingLeading("/")))
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try JSONEncoder().encode(body)
        return try await send(request, skipAuth: skipAuth)
    }

    private func send<T: Decodable>(_ original: URLRequest, skipAuth: Bool) async throws -> T {
        let firstTry = try await attempt(original, skipAuth: skipAuth)
        if firstTry.status == 401, !skipAuth {
            let refreshed = await runRefresh()
            guard refreshed != nil else {
                onUnauthorized()
                throw Error.unauthorized
            }
            let secondTry = try await attempt(original, skipAuth: skipAuth)
            return try decode(secondTry)
        }
        return try decode(firstTry)
    }

    private struct RawResponse {
        let status: Int
        let body: Data
    }

    private func attempt(_ request: URLRequest, skipAuth: Bool) async throws -> RawResponse {
        var request = request
        if !skipAuth, let token = tokenStorage.accessToken() {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
        do {
            let (data, response) = try await session.data(for: request)
            let status = (response as? HTTPURLResponse)?.statusCode ?? -1
            return RawResponse(status: status, body: data)
        } catch {
            throw Error.network(underlying: error)
        }
    }

    private func decode<T: Decodable>(_ raw: RawResponse) throws -> T {
        guard (200..<300).contains(raw.status) else {
            let envelope = try? JSONDecoder().decode(ErrorResponse.self, from: raw.body)
            throw Error.http(status: raw.status, envelope: envelope)
        }
        // `Unit`-style empty body: surface a default-decoded instance for
        // types whose only representation is empty JSON.
        if raw.body.isEmpty, let empty = "{}".data(using: .utf8),
           let decoded = try? JSONDecoder().decode(T.self, from: empty) {
            return decoded
        }
        do {
            return try JSONDecoder().decode(T.self, from: raw.body)
        } catch {
            throw Error.decode(message: "failed to decode \(T.self): \(error)")
        }
    }

    /// Single-flight refresh. Subsequent callers await the in-flight
    /// task instead of starting a new refresh.
    private func runRefresh() async -> TokenPair? {
        if let existing = refreshInFlight {
            return await existing.value
        }
        let task = Task { [tokenStorage, baseURL, session] () -> TokenPair? in
            guard let refresh = tokenStorage.refreshToken() else { return nil }
            var request = URLRequest(url: baseURL.appendingPathComponent("api/auth/refresh"))
            request.httpMethod = "POST"
            request.setValue("application/json", forHTTPHeaderField: "Content-Type")
            request.httpBody = try? JSONEncoder().encode(RefreshRequest(refreshToken: refresh))
            do {
                let (data, response) = try await session.data(for: request)
                guard let http = response as? HTTPURLResponse, (200..<300).contains(http.statusCode) else {
                    tokenStorage.clear()
                    return nil
                }
                let pair = try JSONDecoder().decode(TokenPair.self, from: data)
                tokenStorage.save(tokens: pair)
                return pair
            } catch {
                return nil
            }
        }
        refreshInFlight = task
        let result = await task.value
        refreshInFlight = nil
        return result
    }
}

private extension String {
    /// Trims a single leading occurrence of `prefix` so that
    /// `URL.appendingPathComponent` doesn't double-slash.
    func trimmingLeading(_ prefix: String) -> String {
        hasPrefix(prefix) ? String(dropFirst(prefix.count)) : self
    }
}
