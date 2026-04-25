import Foundation
import Observation

/// App-wide auth state machine.
///
/// SwiftUI views observe this via `@Environment(AuthStore.self)` (or
/// `@Bindable` when they need write access). The root `RootView`
/// switches between login and the authenticated tabs based on
/// `state`.
///
/// The state enum intentionally mirrors the old Flutter `AuthState`
/// sealed class for parity with the web/Android sides:
///  - `loading`         — booting, waiting for `me()` to resolve
///  - `authenticated`   — we have a valid session + loaded user
///  - `unauthenticated` — no session / refresh failed
///  - `error(String)`   — a sign-in attempt failed with a message
@Observable
@MainActor
final class AuthStore {

    enum State: Equatable {
        case loading
        case authenticated(User)
        case unauthenticated
        case error(String)
    }

    private(set) var state: State = .loading

    private let api: APIClient
    private let tokenStorage: TokenStorage
    private let devBypass: Bool

    init(api: APIClient, tokenStorage: TokenStorage, devBypass: Bool = Config.devBypassAuth) {
        self.api = api
        self.tokenStorage = tokenStorage
        self.devBypass = devBypass
    }

    /// Called once at app launch. Hydrates from stored tokens or
    /// drops to the login screen.
    func bootstrap() async {
        if devBypass {
            state = .authenticated(User.devStub)
            return
        }
        guard tokenStorage.accessToken() != nil || tokenStorage.refreshToken() != nil else {
            state = .unauthenticated
            return
        }
        do {
            let user = try await api.me()
            state = .authenticated(user)
        } catch {
            // Refresh attempt inside APIClient already failed; drop to login.
            tokenStorage.clear()
            state = .unauthenticated
        }
    }

    /// Asks the backend to email a magic link. Does NOT change `state`;
    /// the caller surfaces success via its own UI (e.g. "Check your
    /// inbox"), and the real state change happens when the user taps
    /// the link and `verifyMagicLink(token:)` runs.
    func requestMagicLink(email: String) async throws {
        _ = try await api.requestMagicLink(email: email)
    }

    /// Exchanges a magic-link token for a session. Drives state into
    /// `.authenticated` on success or `.error` on failure.
    func verifyMagicLink(token: String) async {
        do {
            _ = try await api.verifyMagicLink(token: token)
            let user = try await api.me()
            state = .authenticated(user)
        } catch {
            state = .error(humanMessage(from: error))
        }
    }

    /// Exchanges a Google ID token for a session.
    func verifyGoogleIDToken(_ idToken: String) async {
        do {
            _ = try await api.verifyGoogleIDToken(idToken)
            let user = try await api.me()
            state = .authenticated(user)
        } catch {
            state = .error(humanMessage(from: error))
        }
    }

    /// Best-effort logout. Clears local tokens regardless of server outcome.
    func logout() async {
        await api.logout()
        state = .unauthenticated
    }

    /// Called by `APIClient` when its refresh attempt fails — drops to
    /// the login screen from any point in the app.
    func handleUnauthorized() {
        tokenStorage.clear()
        state = .unauthenticated
    }

    private func humanMessage(from error: Error) -> String {
        switch error {
        case APIClient.Error.http(_, let envelope):
            return envelope?.message ?? "Sign-in failed."
        case APIClient.Error.network:
            return "Network error. Check your connection and try again."
        case APIClient.Error.decode:
            return "Unexpected server response."
        case APIClient.Error.unauthorized:
            return "Your session expired. Please sign in again."
        default:
            return "Sign-in failed."
        }
    }
}

extension User {
    /// Stub user surfaced when `DEV_BYPASS_AUTH=true`. The UUID is
    /// deterministic so the app doesn't treat every relaunch as a new
    /// user.
    static let devStub = User(
        id: "00000000-0000-0000-0000-000000000001",
        email: "dev@initium.local",
        name: "Dev User",
        avatarURL: "",
        role: .user,
        createdAt: "2026-04-25T00:00:00Z"
    )
}
