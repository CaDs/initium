import Foundation

/// Build-time configuration surface.
///
/// Values are read from Info.plist (set via pbxproj `INFOPLIST_KEY_*`
/// build settings OR a `.xcconfig` file). Falls back to dev-friendly
/// defaults so a fresh clone runs against a local backend without any
/// Xcode configuration.
///
/// Forks should override these by:
/// 1. Adding `API_BASE_URL`, `DEV_BYPASS_AUTH`, `GOOGLE_IOS_CLIENT_ID`
///    to the app target's build settings (Build Settings → Info.plist
///    Values → Custom iOS Target Properties).
/// 2. Or by adding a `.xcconfig` referenced by the build configuration.
enum Config {

    /// Backend URL for the auth + profile endpoints.
    /// iOS simulator reaches the host machine at `localhost`, so the
    /// dev default works without the Android `10.0.2.2` alias.
    static let apiBaseURL: URL = {
        if let raw = Bundle.main.object(forInfoDictionaryKey: "API_BASE_URL") as? String,
           !raw.isEmpty, let url = URL(string: raw) {
            return url
        }
        return URL(string: "http://localhost:8000")!
    }()

    /// When true, the app skips real auth and renders a stub user.
    /// Mirrors the backend's `DEV_BYPASS_AUTH` env var. Release builds
    /// hard-fail if this ever resolves to true (see `initiumApp.swift`).
    static let devBypassAuth: Bool = {
        let raw = Bundle.main.object(forInfoDictionaryKey: "DEV_BYPASS_AUTH") as? String
        return raw?.lowercased() == "true"
    }()

    /// Google iOS OAuth Client ID (ends in `.apps.googleusercontent.com`).
    /// Empty = Google button is disabled at runtime.
    static let googleClientID: String? = {
        let raw = Bundle.main.object(forInfoDictionaryKey: "GOOGLE_IOS_CLIENT_ID") as? String
        return (raw?.isEmpty == false) ? raw : nil
    }()
}
