import Testing
import Foundation
@testable import initium

/// AuthStore is the iOS app's auth state machine. These tests cover the
/// state transitions every protected flow depends on:
///   - bootstrap with no tokens → .unauthenticated
///   - bootstrap with tokens + me() success → .authenticated(user)
///   - bootstrap with tokens + me() failure → .unauthenticated (clears tokens)
///   - bootstrap in dev-bypass → .authenticated(devStub)
///   - verifyMagicLink success → .authenticated, error → .error
///   - verifyGoogleIDToken success → .authenticated
///   - logout → .unauthenticated
///   - handleUnauthorized → .unauthenticated (clears tokens)
///
/// All tests are `@MainActor` because AuthStore is `@MainActor`-isolated.

@MainActor
struct AuthStoreTests {

    // MARK: bootstrap

    @Test func bootstrap_devBypass_setsAuthenticatedStub() async {
        let store = AuthStore(
            api: MockAPIClient(),
            tokenStorage: MockTokenStorage(),
            devBypass: true
        )

        await store.bootstrap()

        if case .authenticated(let user) = store.state {
            #expect(user.id == "00000000-0000-0000-0000-000000000001")
            #expect(user.email == "dev@initium.local")
        } else {
            Issue.record("expected .authenticated, got \(store.state)")
        }
    }

    @Test func bootstrap_noTokens_dropsToUnauthenticated() async {
        let store = AuthStore(
            api: MockAPIClient(),
            tokenStorage: MockTokenStorage(),
            devBypass: false
        )

        await store.bootstrap()

        #expect(store.state == .unauthenticated)
    }

    @Test func bootstrap_withTokens_andMeSucceeds_authenticates() async {
        let api = MockAPIClient()
        let storage = MockTokenStorage(access: "a", refresh: "r")

        let store = AuthStore(api: api, tokenStorage: storage, devBypass: false)
        await store.bootstrap()

        if case .authenticated(let user) = store.state {
            #expect(user.email == "dev@initium.local") // default mock returns devStub
        } else {
            Issue.record("expected .authenticated after successful me()")
        }
        #expect(api.meCallCount == 1)
    }

    @Test func bootstrap_withTokens_andMeFails_clearsTokensAndUnauthenticates() async {
        let api = MockAPIClient()
        api.meResult = .failure(APIClient.Error.unauthorized)
        let storage = MockTokenStorage(access: "a", refresh: "r")

        let store = AuthStore(api: api, tokenStorage: storage, devBypass: false)
        await store.bootstrap()

        #expect(store.state == .unauthenticated)
        // Tokens must be cleared so the next bootstrap doesn't loop.
        #expect(storage.accessToken() == nil)
        #expect(storage.refreshToken() == nil)
    }

    // MARK: verifyMagicLink

    @Test func verifyMagicLink_success_authenticates() async {
        let api = MockAPIClient()
        let store = AuthStore(api: api, tokenStorage: MockTokenStorage(), devBypass: false)

        await store.verifyMagicLink(token: "valid-token")

        #expect(api.verifyMagicLinkCalls == ["valid-token"])
        if case .authenticated = store.state {
            // ok
        } else {
            Issue.record("expected .authenticated after successful verifyMagicLink")
        }
    }

    @Test func verifyMagicLink_failure_setsErrorState() async {
        let api = MockAPIClient()
        api.verifyMagicLinkResult = .failure(
            APIClient.Error.http(status: 410, envelope: ErrorResponse(code: "TOKEN_EXPIRED", message: "token expired", requestId: nil))
        )
        let store = AuthStore(api: api, tokenStorage: MockTokenStorage(), devBypass: false)

        await store.verifyMagicLink(token: "expired-token")

        if case .error(let msg) = store.state {
            #expect(msg == "token expired")
        } else {
            Issue.record("expected .error state, got \(store.state)")
        }
    }

    // MARK: verifyGoogleIDToken

    @Test func verifyGoogleIDToken_success_authenticates() async {
        let api = MockAPIClient()
        let store = AuthStore(api: api, tokenStorage: MockTokenStorage(), devBypass: false)

        await store.verifyGoogleIDToken("google-id-token")

        #expect(api.verifyGoogleIDTokenCalls == ["google-id-token"])
        if case .authenticated = store.state {
            // ok
        } else {
            Issue.record("expected .authenticated after successful Google sign-in")
        }
    }

    @Test func verifyGoogleIDToken_networkError_humanizesMessage() async {
        let api = MockAPIClient()
        api.verifyGoogleIDTokenResult = .failure(
            APIClient.Error.network(underlying: NSError(domain: "test", code: -1))
        )
        let store = AuthStore(api: api, tokenStorage: MockTokenStorage(), devBypass: false)

        await store.verifyGoogleIDToken("bad-token")

        if case .error(let msg) = store.state {
            // Network errors get a human-friendly translation, not the raw NSError.
            #expect(msg.contains("Network error"))
        } else {
            Issue.record("expected .error, got \(store.state)")
        }
    }

    // MARK: logout / handleUnauthorized

    @Test func logout_dropsToUnauthenticated_andCallsAPI() async {
        let api = MockAPIClient()
        let storage = MockTokenStorage(access: "a", refresh: "r")
        let store = AuthStore(api: api, tokenStorage: storage, devBypass: false)

        await store.logout()

        #expect(store.state == .unauthenticated)
        #expect(api.logoutCallCount == 1)
    }

    @Test func handleUnauthorized_clearsTokens_andUnauthenticates() async {
        let storage = MockTokenStorage(access: "a", refresh: "r")
        let store = AuthStore(api: MockAPIClient(), tokenStorage: storage, devBypass: false)

        store.handleUnauthorized()

        #expect(store.state == .unauthenticated)
        #expect(storage.accessToken() == nil)
        #expect(storage.refreshToken() == nil)
    }
}
