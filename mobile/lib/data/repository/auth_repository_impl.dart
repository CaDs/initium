import 'package:dio/dio.dart';
import '../../domain/entity/session.dart';
import '../../domain/error/domain_error.dart';
import '../../domain/repository/auth_repository.dart';
import '../remote/api_client.dart';
import '../remote/dto/auth_dto.dart';
import '../remote/mapper/auth_mapper.dart';

class AuthRepositoryImpl implements AuthRepository {
  final ApiClient _client;

  AuthRepositoryImpl(this._client);

  @override
  Future<(Session?, DomainError?)> loginWithGoogle(String idToken) async {
    try {
      final response = await _client.dio.post(
        '/api/auth/mobile/google',
        data: {'id_token': idToken},
      );
      final dto = AuthResponseDto.fromJson(response.data);
      return (dto.toDomain(), null);
    } on DioException catch (e) {
      return (null, _mapError(e));
    }
  }

  @override
  Future<(bool, DomainError?)> requestMagicLink(String email) async {
    try {
      await _client.dio.post('/api/auth/magic-link', data: {'email': email});
      return (true, null);
    } on DioException catch (e) {
      return (false, _mapError(e));
    }
  }

  @override
  Future<(Session?, DomainError?)> verifyMagicLink(String token) async {
    try {
      final response = await _client.dio.post(
        '/api/auth/mobile/verify',
        data: {'token': token},
      );
      final dto = AuthResponseDto.fromJson(response.data);
      return (dto.toDomain(), null);
    } on DioException catch (e) {
      return (null, _mapError(e));
    }
  }

  @override
  Future<(Session?, DomainError?)> refreshTokens(String refreshToken) async {
    try {
      final response = await _client.dio.post(
        '/api/auth/refresh',
        data: {'refresh_token': refreshToken},
      );
      final dto = AuthResponseDto.fromJson(response.data);
      return (dto.toDomain(), null);
    } on DioException catch (e) {
      return (null, _mapError(e));
    }
  }

  @override
  Future<DomainError?> logout(String refreshToken) async {
    try {
      await _client.dio.post('/api/auth/logout');
      return null;
    } on DioException catch (e) {
      return _mapError(e);
    }
  }

  DomainError _mapError(DioException e) {
    if (e.response?.statusCode == 401) return const Unauthorized();
    if (e.type == DioExceptionType.connectionTimeout ||
        e.type == DioExceptionType.receiveTimeout) {
      return const NetworkError('Connection timeout');
    }
    final data = e.response?.data;
    if (data is Map<String, dynamic>) {
      return ServerError(
        data['code'] as String? ?? 'UNKNOWN',
        data['message'] as String? ?? 'Unknown error',
      );
    }
    return UnknownError(e.message ?? 'Unknown error');
  }
}
