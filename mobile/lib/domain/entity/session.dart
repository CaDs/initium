/// Session entity — represents a token pair from the backend.
class Session {
  final String accessToken;
  final String refreshToken;

  const Session({
    required this.accessToken,
    required this.refreshToken,
  });
}
