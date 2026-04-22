/// User response DTO from backend.
class UserDto {
  final String id;
  final String email;
  final String name;
  final String? avatarUrl;
  final String role;
  final String createdAt;

  UserDto({
    required this.id,
    required this.email,
    required this.name,
    this.avatarUrl,
    required this.role,
    required this.createdAt,
  });

  factory UserDto.fromJson(Map<String, dynamic> json) {
    return UserDto(
      id: json['id'] as String,
      email: json['email'] as String,
      name: json['name'] as String,
      avatarUrl: json['avatar_url'] as String?,
      role: json['role'] as String,
      createdAt: json['created_at'] as String,
    );
  }
}
