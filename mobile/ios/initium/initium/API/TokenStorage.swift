import Foundation
import Security

/// Persists the access + refresh token pair in Keychain.
///
/// - The keychain item survives app uninstall. We pair it with a
///   `UserDefaults` first-launch marker so that after a reinstall,
///   residual keychain entries from a prior install are wiped on the
///   first launch. This prevents a reused device handing out stale
///   tokens to a fresh install.
/// - Not thread-safe on its own; `APIClient` serializes concurrent
///   refresh attempts with its refresh actor.
///
/// Keychain access group is the default (app bundle); fork owners can
/// change `service` to match their reverse-DNS if they want to share
/// tokens across a suite of apps.
final class TokenStorage: @unchecked Sendable {

    private let service: String
    private let accessAccount = "access_token"
    private let refreshAccount = "refresh_token"
    private let installMarkerKey = "com.initium.keychain.install-marker"

    init(service: String = Bundle.main.bundleIdentifier ?? "cads.initium") {
        self.service = service
        wipeOnFirstLaunchIfNeeded()
    }

    func save(tokens: TokenPair) {
        set(value: tokens.accessToken, account: accessAccount)
        set(value: tokens.refreshToken, account: refreshAccount)
    }

    func accessToken() -> String? {
        read(account: accessAccount)
    }

    func refreshToken() -> String? {
        read(account: refreshAccount)
    }

    func clear() {
        delete(account: accessAccount)
        delete(account: refreshAccount)
    }

    // MARK: - Keychain helpers

    private func query(account: String) -> [String: Any] {
        [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: account,
        ]
    }

    private func set(value: String, account: String) {
        let data = Data(value.utf8)
        var query = query(account: account)
        SecItemDelete(query as CFDictionary)
        query[kSecValueData as String] = data
        query[kSecAttrAccessible as String] = kSecAttrAccessibleAfterFirstUnlockThisDeviceOnly
        SecItemAdd(query as CFDictionary, nil)
    }

    private func read(account: String) -> String? {
        var query = query(account: account)
        query[kSecReturnData as String] = true
        query[kSecMatchLimit as String] = kSecMatchLimitOne

        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)
        guard status == errSecSuccess, let data = result as? Data else { return nil }
        return String(data: data, encoding: .utf8)
    }

    private func delete(account: String) {
        SecItemDelete(query(account: account) as CFDictionary)
    }

    private func wipeOnFirstLaunchIfNeeded() {
        let defaults = UserDefaults.standard
        if defaults.bool(forKey: installMarkerKey) { return }
        delete(account: accessAccount)
        delete(account: refreshAccount)
        defaults.set(true, forKey: installMarkerKey)
    }
}
