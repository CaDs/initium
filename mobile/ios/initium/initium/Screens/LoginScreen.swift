import SwiftUI

/// The authentication entry point. Offers two paths:
///  - **Magic link**: user enters their email → backend emails a
///    one-use deep link (`initium://auth/verify?token=...`) → tapping
///    it elsewhere routes back to `initiumApp.onOpenURL`.
///  - **Google Sign-In**: disabled in this PR. The button appears
///    visually dimmed until the Google SDK integration lands.
struct LoginScreen: View {
    @Environment(AuthStore.self) private var authStore

    @State private var email: String = ""
    @State private var magicLinkState: MagicLinkState = .idle

    enum MagicLinkState: Equatable {
        case idle
        case sending
        case sent
        case failed(String)
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 24) {
                header

                googleButton

                divider

                magicLinkSection

                if case .error(let message) = authStore.state {
                    errorBanner(message)
                }
            }
            .padding(.horizontal, 24)
            .padding(.vertical, 40)
            .frame(maxWidth: 480)
            .frame(maxWidth: .infinity, alignment: .leading)
        }
        .navigationTitle("Sign In")
        .navigationBarTitleDisplayMode(.inline)
    }

    private var header: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Welcome to Initium")
                .font(.largeTitle.weight(.semibold))
            Text("Sign in with Google or request a magic link by email.")
                .font(.subheadline)
                .foregroundStyle(.secondary)
        }
    }

    private var googleButton: some View {
        Button {
            // Wired in the follow-up Google Sign-In PR.
        } label: {
            HStack(spacing: 10) {
                Image(systemName: "g.circle.fill")
                    .font(.title3)
                Text("Sign in with Google")
                    .font(.headline)
            }
            .frame(maxWidth: .infinity, minHeight: 44)
        }
        .buttonStyle(.bordered)
        .disabled(true)
        .opacity(0.5)
        .accessibilityHint("Coming soon")
    }

    private var divider: some View {
        HStack(spacing: 12) {
            Rectangle().fill(.secondary.opacity(0.3)).frame(height: 1)
            Text("or").font(.footnote).foregroundStyle(.secondary)
            Rectangle().fill(.secondary.opacity(0.3)).frame(height: 1)
        }
    }

    private var magicLinkSection: some View {
        VStack(alignment: .leading, spacing: 12) {
            TextField("Enter your email", text: $email)
                .textFieldStyle(.roundedBorder)
                .keyboardType(.emailAddress)
                .textContentType(.emailAddress)
                .autocorrectionDisabled()
                .textInputAutocapitalization(.never)
                .disabled(magicLinkState == .sending || magicLinkState == .sent)

            Button(action: submitMagicLink) {
                Text(magicLinkButtonLabel)
                    .font(.headline)
                    .frame(maxWidth: .infinity, minHeight: 44)
            }
            .buttonStyle(.borderedProminent)
            .disabled(!canSubmit)

            switch magicLinkState {
            case .sent:
                sentBanner
            case .failed(let message):
                errorBanner(message)
            case .idle, .sending:
                EmptyView()
            }
        }
    }

    private var magicLinkButtonLabel: String {
        switch magicLinkState {
        case .idle, .failed: return "Send Magic Link"
        case .sending:        return "Sending..."
        case .sent:           return "Sent"
        }
    }

    private var canSubmit: Bool {
        if magicLinkState == .sending || magicLinkState == .sent { return false }
        return email.contains("@") && email.count > 3
    }

    private var sentBanner: some View {
        VStack(alignment: .leading, spacing: 4) {
            Text("Magic link sent")
                .font(.subheadline.weight(.semibold))
            Text("Check your inbox — the link expires in 15 minutes.")
                .font(.footnote)
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(12)
        .background(.green.opacity(0.1), in: RoundedRectangle(cornerRadius: 12))
    }

    private func errorBanner(_ message: String) -> some View {
        Text(message)
            .font(.footnote)
            .foregroundStyle(.red)
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(12)
            .background(.red.opacity(0.1), in: RoundedRectangle(cornerRadius: 12))
    }

    private func submitMagicLink() {
        magicLinkState = .sending
        Task {
            do {
                try await authStore.requestMagicLink(email: email)
                magicLinkState = .sent
            } catch {
                let message = (error as? APIClient.Error).flatMap { err -> String? in
                    if case .http(_, let env) = err { return env?.message }
                    return nil
                } ?? "Could not send magic link. Try again."
                magicLinkState = .failed(message)
            }
        }
    }
}

#Preview {
    let tokenStorage = TokenStorage()
    let api = APIClient(baseURL: URL(string: "http://localhost:8000")!, tokenStorage: tokenStorage)
    let store = AuthStore(api: api, tokenStorage: tokenStorage, devBypass: false)
    NavigationStack {
        LoginScreen()
            .environment(store)
    }
}
