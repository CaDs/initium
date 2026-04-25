import SwiftUI

@main
struct initiumApp: App {

    /// App-level dependency graph. Constructed once; shared through
    /// SwiftUI's environment. No DI container — the graph is tiny.
    @State private var authStore: AuthStore

    init() {
        let tokenStorage = TokenStorage()

        // `api` refers to `authStore` inside `onUnauthorized` but we
        // haven't constructed the store yet. Work around by capturing
        // a weak reference after both are built.
        let api = APIClient(
            baseURL: Config.apiBaseURL,
            tokenStorage: tokenStorage,
            onUnauthorized: { /* wired below */ }
        )
        let store = AuthStore(api: api, tokenStorage: tokenStorage)
        _authStore = State(initialValue: store)

        // Release-build guard: DEV_BYPASS_AUTH must never be true in a
        // shipped build. Matches the old Flutter check.
        #if !DEBUG
        if Config.devBypassAuth {
            fatalError("DEV_BYPASS_AUTH must not be enabled in release builds.")
        }
        #endif
    }

    var body: some Scene {
        WindowGroup {
            RootView()
                .environment(authStore)
                .task {
                    await authStore.bootstrap()
                }
                .onOpenURL { url in
                    handleDeepLink(url)
                }
        }
    }

    /// Handles `initium://auth/verify?token=...` deep links. Anything
    /// else is ignored silently so third-party scheme collisions don't
    /// crash the app.
    private func handleDeepLink(_ url: URL) {
        guard let token = parseMagicLinkToken(from: url) else { return }
        Task {
            await authStore.verifyMagicLink(token: token)
        }
    }
}
