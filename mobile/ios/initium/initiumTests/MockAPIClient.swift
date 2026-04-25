import Foundation
@testable import initium

/// In-memory fake `APIClientProtocol` for unit tests. Each method has a
/// configurable `Result` (or just a return value). Default behaviour
/// returns canned values; tests override only the methods they care
/// about.
///
/// Made a `final class` instead of a struct so tests can mutate
/// per-call counters without forcing immutability gymnastics.
final class MockAPIClient: APIClientProtocol, @unchecked Sendable {

    var requestMagicLinkResult: Result<MessageResponse, Error> = .success(.init(message: "magic link sent"))
    var verifyMagicLinkResult: Result<TokenPair, Error> = .success(.init(accessToken: "test-access", refreshToken: "test-refresh"))
    var verifyGoogleIDTokenResult: Result<TokenPair, Error> = .success(.init(accessToken: "test-access", refreshToken: "test-refresh"))
    var meResult: Result<User, Error> = .success(.devStub)

    private(set) var requestMagicLinkCalls: [String] = []
    private(set) var verifyMagicLinkCalls: [String] = []
    private(set) var verifyGoogleIDTokenCalls: [String] = []
    private(set) var meCallCount = 0
    private(set) var logoutCallCount = 0

    func requestMagicLink(email: String) async throws -> MessageResponse {
        requestMagicLinkCalls.append(email)
        return try requestMagicLinkResult.get()
    }

    func verifyMagicLink(token: String) async throws -> TokenPair {
        verifyMagicLinkCalls.append(token)
        return try verifyMagicLinkResult.get()
    }

    func verifyGoogleIDToken(_ idToken: String) async throws -> TokenPair {
        verifyGoogleIDTokenCalls.append(idToken)
        return try verifyGoogleIDTokenResult.get()
    }

    func me() async throws -> User {
        meCallCount += 1
        return try meResult.get()
    }

    func logout() async {
        logoutCallCount += 1
    }
}

/// In-memory `TokenStorageProtocol` fake. Mirrors the keychain-backed
/// storage's behaviour without the system call.
final class MockTokenStorage: TokenStorageProtocol, @unchecked Sendable {

    private var access: String?
    private var refresh: String?

    init(access: String? = nil, refresh: String? = nil) {
        self.access = access
        self.refresh = refresh
    }

    func save(tokens: TokenPair) {
        access = tokens.accessToken
        refresh = tokens.refreshToken
    }

    func accessToken() -> String? { access }
    func refreshToken() -> String? { refresh }

    func clear() {
        access = nil
        refresh = nil
    }
}
