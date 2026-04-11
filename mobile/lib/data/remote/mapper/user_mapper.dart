import '../../../domain/entity/user.dart';
import '../dto/user_dto.dart';

extension UserDtoMapper on UserDto {
  User toDomain() => User(
        id: id,
        email: email,
        name: name,
        avatarUrl: avatarUrl,
        createdAt: createdAt,
      );
}
