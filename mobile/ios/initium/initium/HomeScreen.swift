import SwiftUI

/// Authenticated Home tab. Shows the profile card mirroring what
/// `web/src/app/home/page.tsx` and the Android `HomeProfile` render
/// (email / name / role / id), plus a logout button. The card uses the
/// Liquid Glass treatment on iOS 26+ and a regularMaterial fallback
/// on earlier OSes — same opt-in as the starter template's MVP.
struct HomeScreen: View {
    @Environment(AuthStore.self) private var authStore

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 24) {
                    header

                    if case .authenticated(let user) = authStore.state {
                        profileCard(user: user)
                            .padding(.horizontal)

                        logoutButton
                            .padding(.horizontal)
                    }
                }
                .padding(.vertical, 32)
            }
            .navigationTitle("Home")
            .navigationBarTitleDisplayMode(.inline)
        }
    }

    private var header: some View {
        Text("Home")
            .font(.largeTitle.weight(.semibold))
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(.horizontal)
    }

    private func profileCard(user: User) -> some View {
        VStack(alignment: .leading, spacing: 16) {
            profileRow(label: "Email", value: user.email)
            Divider()
            profileRow(label: "Name", value: user.name.isEmpty ? "—" : user.name)
            Divider()
            profileRow(label: "Role", value: user.role.rawValue)
            Divider()
            profileRow(label: "ID", value: user.id, monospace: true)
        }
        .padding(20)
        .frame(maxWidth: .infinity, alignment: .leading)
        .liquidGlassCard()
    }

    private func profileRow(label: String, value: String, monospace: Bool = false) -> some View {
        VStack(alignment: .leading, spacing: 4) {
            Text(label)
                .font(.caption.weight(.medium))
                .foregroundStyle(.secondary)
            Text(value)
                .font(monospace ? .system(.caption, design: .monospaced) : .subheadline)
                .textSelection(.enabled)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }

    private var logoutButton: some View {
        Button {
            Task { await authStore.logout() }
        } label: {
            Text("Logout")
                .font(.headline)
                .frame(maxWidth: .infinity, minHeight: 44)
        }
        .buttonStyle(.bordered)
    }
}

#Preview {
    let tokenStorage = TokenStorage()
    let api = APIClient(baseURL: URL(string: "http://localhost:8000")!, tokenStorage: tokenStorage)
    let store = AuthStore(api: api, tokenStorage: tokenStorage, devBypass: true)
    HomeScreen().environment(store)
}
