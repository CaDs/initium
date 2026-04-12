import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../domain/entity/user.dart';
import '../data/local/session_manager.dart';

/// AuthState — UI concern, lives here in providers, NOT in domain.
sealed class AuthState {
  const AuthState();
}

class AuthLoading extends AuthState {
  const AuthLoading();
}

class AuthAuthenticated extends AuthState {
  final User user;
  const AuthAuthenticated(this.user);
}

class AuthUnauthenticated extends AuthState {
  const AuthUnauthenticated();
}

class AuthError extends AuthState {
  final String message;
  const AuthError(this.message);
}

/// Auth notifier wrapping SessionManager.
class AuthNotifier extends StateNotifier<AuthState> {
  final SessionManager _sessionManager;
  final bool _devBypassAuth;

  AuthNotifier(this._sessionManager, {bool devBypassAuth = false})
      : _devBypassAuth = devBypassAuth,
        super(const AuthLoading()) {
    _init();
  }

  Future<void> _init() async {
    if (_devBypassAuth) {
      state = AuthAuthenticated(User.stub());
      return;
    }

    final user = await _sessionManager.restoreSession();
    if (user != null) {
      state = AuthAuthenticated(user);
    } else {
      state = const AuthUnauthenticated();
    }
  }

  Future<void> loginWithGoogle(String idToken) async {
    state = const AuthLoading();
    final user = await _sessionManager.loginWithGoogle(idToken);
    if (user != null) {
      state = AuthAuthenticated(user);
    } else {
      state = const AuthError('Google login failed');
    }
  }

  Future<bool> requestMagicLink(String email) async {
    return _sessionManager.requestMagicLink(email);
  }

  Future<void> verifyMagicLink(String token) async {
    state = const AuthLoading();
    final user = await _sessionManager.verifyMagicLink(token);
    if (user != null) {
      state = AuthAuthenticated(user);
    } else {
      state = const AuthError('Magic link verification failed');
    }
  }

  Future<void> logout() async {
    await _sessionManager.logout();
    state = const AuthUnauthenticated();
  }
}
