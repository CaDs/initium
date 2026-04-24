import SwiftUI

/// Root view, bound to the app-level `AuthStore`. Switches between
/// the unauthenticated login flow and the authenticated tab shell
/// based on auth state. Loading and error states surface while we
/// hydrate / recover.
struct RootView: View {
    @Environment(AuthStore.self) private var authStore

    var body: some View {
        switch authStore.state {
        case .loading:
            LoadingView()
        case .authenticated:
            TabsRootView()
        case .unauthenticated:
            NavigationStack { LoginScreen() }
        case .error(let message):
            NavigationStack {
                LoginScreen()
                    .onAppear { /* The message already surfaces inline. */ _ = message }
            }
        }
    }
}

/// The previously-top-level 3-tab TabView, now nested inside the
/// authenticated branch of `RootView`.
struct TabsRootView: View {
    @State private var selection: AppTab = .home

    var body: some View {
        TabView(selection: $selection) {
            HomeScreen()
                .tabItem { Label(AppTab.home.title, systemImage: AppTab.home.systemImage) }
                .tag(AppTab.home)

            MainScreen()
                .tabItem { Label(AppTab.main.title, systemImage: AppTab.main.systemImage) }
                .tag(AppTab.main)

            SettingsScreen()
                .tabItem { Label(AppTab.settings.title, systemImage: AppTab.settings.systemImage) }
                .tag(AppTab.settings)
        }
    }
}

struct LoadingView: View {
    var body: some View {
        VStack(spacing: 16) {
            ProgressView()
            Text("Loading")
                .font(.footnote)
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}

#Preview("Authenticated") {
    let tokenStorage = TokenStorage()
    let api = APIClient(baseURL: URL(string: "http://localhost:8000")!, tokenStorage: tokenStorage)
    let store = AuthStore(api: api, tokenStorage: tokenStorage, devBypass: true)
    RootView().environment(store)
}
