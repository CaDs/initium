import 'package:dio/dio.dart';
import '../../domain/entity/user.dart';
import '../../domain/error/domain_error.dart';
import '../../domain/repository/user_repository.dart';
import '../remote/api_client.dart';
import '../remote/dto/user_dto.dart';
import '../remote/mapper/user_mapper.dart';

class UserRepositoryImpl implements UserRepository {
  final ApiClient _client;

  UserRepositoryImpl(this._client);

  @override
  Future<(User?, DomainError?)> getProfile(String accessToken) async {
    try {
      final response = await _client.dio.get('/api/me');
      final dto = UserDto.fromJson(response.data);
      return (dto.toDomain(), null);
    } on DioException catch (e) {
      return (null, _mapError(e));
    }
  }

  @override
  Future<(User?, DomainError?)> updateProfile(String accessToken, String name) async {
    try {
      final response = await _client.dio.patch('/api/me', data: {'name': name});
      final dto = UserDto.fromJson(response.data);
      return (dto.toDomain(), null);
    } on DioException catch (e) {
      return (null, _mapError(e));
    }
  }

  DomainError _mapError(DioException e) {
    if (e.response?.statusCode == 401) return const Unauthorized();
    return UnknownError(e.message ?? 'Unknown error');
  }
}
