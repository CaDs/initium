/// Auth response DTO from backend.
class AuthResponseDto {
  final String accessToken;
  final String refreshToken;

  AuthResponseDto({required this.accessToken, required this.refreshToken});

  factory AuthResponseDto.fromJson(Map<String, dynamic> json) {
    return AuthResponseDto(
      accessToken: json['access_token'] as String,
      refreshToken: json['refresh_token'] as String,
    );
  }
}

/// Magic link request response.
class MessageResponseDto {
  final String message;

  MessageResponseDto({required this.message});

  factory MessageResponseDto.fromJson(Map<String, dynamic> json) {
    return MessageResponseDto(message: json['message'] as String);
  }
}
