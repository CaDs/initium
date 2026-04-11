import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Secure token storage using Keychain (iOS) / EncryptedSharedPreferences (Android).
/// Handles iOS keychain persistence across app reinstall.
class TokenStorage {
  static const _accessTokenKey = 'access_token';
  static const _refreshTokenKey = 'refresh_token';
  static const _firstLaunchKey = 'initium_first_launch_done';

  final FlutterSecureStorage _storage;

  TokenStorage() : _storage = const FlutterSecureStorage(
    aOptions: AndroidOptions(encryptedSharedPreferences: true),
    iOptions: IOSOptions(accessibility: KeychainAccessibility.first_unlock),
  );

  /// Must be called on app startup. Wipes stale keychain data on iOS reinstall.
  Future<void> initialize() async {
    final prefs = await SharedPreferences.getInstance();
    final firstLaunchDone = prefs.getBool(_firstLaunchKey) ?? false;

    if (!firstLaunchDone) {
      await _storage.deleteAll();
      await prefs.setBool(_firstLaunchKey, true);
    }
  }

  Future<void> saveTokens(String accessToken, String refreshToken) async {
    await _storage.write(key: _accessTokenKey, value: accessToken);
    await _storage.write(key: _refreshTokenKey, value: refreshToken);
  }

  Future<String?> getAccessToken() => _storage.read(key: _accessTokenKey);
  Future<String?> getRefreshToken() => _storage.read(key: _refreshTokenKey);

  Future<void> clear() async {
    await _storage.delete(key: _accessTokenKey);
    await _storage.delete(key: _refreshTokenKey);
  }
}
