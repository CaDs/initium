import Foundation

/// Pulls the magic-link token out of an `initium://auth/verify?token=...`
/// URL. Returns nil for any URL that doesn't match exactly so third-party
/// scheme collisions don't accidentally trigger an auth attempt.
///
/// Lives as a free function so unit tests can exercise the URL parsing
/// without instantiating the `App` value or its `@MainActor` graph.
func parseMagicLinkToken(from url: URL) -> String? {
    guard url.scheme == "initium",
          url.host == "auth",
          url.path == "/verify",
          let components = URLComponents(url: url, resolvingAgainstBaseURL: false),
          let token = components.queryItems?.first(where: { $0.name == "token" })?.value,
          !token.isEmpty
    else { return nil }
    return token
}
