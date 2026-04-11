import '../../domain/entity/user.dart';
import '../../domain/repository/auth_repository.dart';
import '../../domain/repository/user_repository.dart';
import 'token_storage.dart';

/// Plain Dart class managing auth lifecycle. Wrapped by Riverpod in providers/.
class SessionManager {
  final TokenStorage _tokenStorage;
  final AuthRepository _authRepo;
  final UserRepository _userRepo;

  SessionManager({
    required TokenStorage tokenStorage,
    required AuthRepository authRepo,
    required UserRepository userRepo,
  })  : _tokenStorage = tokenStorage,
        _authRepo = authRepo,
        _userRepo = userRepo;

  /// Check if user has a valid session on startup.
  Future<User?> restoreSession() async {
    final accessToken = await _tokenStorage.getAccessToken();
    if (accessToken == null) return null;

    final (user, error) = await _userRepo.getProfile(accessToken);
    if (error != null || user == null) {
      // Try refresh
      final refreshToken = await _tokenStorage.getRefreshToken();
      if (refreshToken == null) return null;

      final (session, refreshError) = await _authRepo.refreshTokens(refreshToken);
      if (refreshError != null || session == null) {
        await _tokenStorage.clear();
        return null;
      }

      await _tokenStorage.saveTokens(session.accessToken, session.refreshToken);
      final (refreshedUser, _) = await _userRepo.getProfile(session.accessToken);
      return refreshedUser;
    }

    return user;
  }

  /// Login with Google ID token.
  Future<User?> loginWithGoogle(String idToken) async {
    final (session, error) = await _authRepo.loginWithGoogle(idToken);
    if (error != null || session == null) return null;

    await _tokenStorage.saveTokens(session.accessToken, session.refreshToken);
    final (user, _) = await _userRepo.getProfile(session.accessToken);
    return user;
  }

  /// Request magic link email.
  Future<bool> requestMagicLink(String email) async {
    final (success, _) = await _authRepo.requestMagicLink(email);
    return success;
  }

  /// Verify magic link token.
  Future<User?> verifyMagicLink(String token) async {
    final (session, error) = await _authRepo.verifyMagicLink(token);
    if (error != null || session == null) return null;

    await _tokenStorage.saveTokens(session.accessToken, session.refreshToken);
    final (user, _) = await _userRepo.getProfile(session.accessToken);
    return user;
  }

  /// Logout and clear tokens.
  Future<void> logout() async {
    final refreshToken = await _tokenStorage.getRefreshToken();
    if (refreshToken != null) {
      await _authRepo.logout(refreshToken);
    }
    await _tokenStorage.clear();
  }
}
