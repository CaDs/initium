/// Domain errors — pure Dart sealed class.
sealed class DomainError {
  const DomainError();
}

class Unauthorized extends DomainError {
  const Unauthorized();
}

class TokenExpired extends DomainError {
  const TokenExpired();
}

class NetworkError extends DomainError {
  final String message;
  const NetworkError(this.message);
}

class ServerError extends DomainError {
  final String code;
  final String message;
  const ServerError(this.code, this.message);
}

class UnknownError extends DomainError {
  final String message;
  const UnknownError(this.message);
}
