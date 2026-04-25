import Foundation

/// Behavior surface of `TokenStorage`, exposed as a protocol so tests
/// can substitute an in-memory implementation without touching the
/// real Keychain. Production code should accept `any TokenStorageProtocol`,
/// and the concrete `TokenStorage` instance is constructed at app launch.
///
/// The protocol is `Sendable` because the production type is — both
/// `APIClient` (actor) and `AuthStore` (`@MainActor`) need to hold a
/// storage reference safely across isolation boundaries.
protocol TokenStorageProtocol: Sendable {
    func save(tokens: TokenPair)
    func accessToken() -> String?
    func refreshToken() -> String?
    func clear()
}

extension TokenStorage: TokenStorageProtocol {}
