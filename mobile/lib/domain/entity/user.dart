/// User entity — pure Dart, no framework imports.
class User {
  final String id;
  final String email;
  final String name;
  final String? avatarUrl;
  final String createdAt;

  const User({
    required this.id,
    required this.email,
    required this.name,
    this.avatarUrl,
    required this.createdAt,
  });
}
