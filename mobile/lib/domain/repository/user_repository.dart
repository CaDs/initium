import '../entity/user.dart';
import '../error/domain_error.dart';

/// User repository interface — pure Dart.
abstract class UserRepository {
  Future<(User?, DomainError?)> getProfile(String accessToken);
  Future<(User?, DomainError?)> updateProfile(String accessToken, String name);
}
