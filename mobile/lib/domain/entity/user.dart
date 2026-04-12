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

  /// Stub user for DEV_BYPASS_AUTH mode.
  factory User.stub() => const User(
        id: '00000000-0000-0000-0000-000000000001',
        email: 'dev@initium.local',
        name: 'Dev User',
        createdAt: '2026-01-01T00:00:00Z',
      );
}
