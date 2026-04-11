import '../entity/session.dart';
import '../error/domain_error.dart';

/// Auth repository interface — pure Dart.
abstract class AuthRepository {
  Future<(Session?, DomainError?)> loginWithGoogle(String idToken);
  Future<(bool, DomainError?)> requestMagicLink(String email);
  Future<(Session?, DomainError?)> verifyMagicLink(String token);
  Future<(Session?, DomainError?)> refreshTokens(String refreshToken);
  Future<DomainError?> logout(String refreshToken);
}
