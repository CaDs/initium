import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../data/remote/api_client.dart';
import '../data/local/token_storage.dart';
import '../data/local/session_manager.dart';
import '../data/repository/auth_repository_impl.dart';
import '../data/repository/user_repository_impl.dart';
import 'auth_provider.dart';

const _apiBaseUrl = String.fromEnvironment(
  'API_BASE_URL',
  defaultValue: 'http://localhost:8000',
);

const _devBypassAuth = bool.fromEnvironment('DEV_BYPASS_AUTH');

final tokenStorageProvider = Provider<TokenStorage>((ref) => TokenStorage());

final apiClientProvider = Provider<ApiClient>((ref) {
  final tokenStorage = ref.read(tokenStorageProvider);
  return ApiClient(baseUrl: _apiBaseUrl, tokenStorage: tokenStorage);
});

final authRepositoryProvider = Provider((ref) {
  return AuthRepositoryImpl(ref.read(apiClientProvider));
});

final userRepositoryProvider = Provider((ref) {
  return UserRepositoryImpl(ref.read(apiClientProvider));
});

final sessionManagerProvider = Provider((ref) {
  return SessionManager(
    tokenStorage: ref.read(tokenStorageProvider),
    authRepo: ref.read(authRepositoryProvider),
    userRepo: ref.read(userRepositoryProvider),
  );
});

final authProvider = StateNotifierProvider<AuthNotifier, AuthState>((ref) {
  return AuthNotifier(ref.read(sessionManagerProvider));
});

/// Whether dev bypass auth is enabled (compile-time constant).
bool get isDevBypassAuth => _devBypassAuth;
