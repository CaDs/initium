import '../../../domain/entity/session.dart';
import '../dto/auth_dto.dart';

extension AuthResponseDtoMapper on AuthResponseDto {
  Session toDomain() => Session(
        accessToken: accessToken,
        refreshToken: refreshToken,
      );
}
