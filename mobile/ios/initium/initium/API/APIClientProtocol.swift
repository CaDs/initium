import Foundation

/// Behavior surface of `APIClient`, exposed as a protocol so tests can
/// substitute an in-memory fake without spinning up `URLSession`.
///
/// The concrete `APIClient` is an actor; the protocol intentionally is
/// not actor-bound so a non-actor test fake (e.g. a struct/class that
/// stores closures) can conform without becoming an actor itself. All
/// methods are `async`, so calls into the real actor or a class fake
/// are equivalently safe.
///
/// Production code should depend on `any APIClientProtocol`, not the
/// concrete class. Wiring at app launch (`initiumApp.swift`) supplies
/// the real `APIClient`.
protocol APIClientProtocol: Sendable {
    func requestMagicLink(email: String) async throws -> MessageResponse
    func verifyMagicLink(token: String) async throws -> TokenPair
    func verifyGoogleIDToken(_ idToken: String) async throws -> TokenPair
    func me() async throws -> User
    func logout() async
}

extension APIClient: APIClientProtocol {}
