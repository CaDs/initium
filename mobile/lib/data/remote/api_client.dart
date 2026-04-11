import 'dart:async';
import 'package:dio/dio.dart';
import '../local/token_storage.dart';

/// Dio HTTP client with auth interceptor and token refresh lock.
class ApiClient {
  final Dio dio;
  final TokenStorage _tokenStorage;
  Completer<void>? _refreshLock;

  ApiClient({
    required String baseUrl,
    required TokenStorage tokenStorage,
  })  : _tokenStorage = tokenStorage,
        dio = Dio(BaseOptions(
          baseUrl: baseUrl,
          connectTimeout: const Duration(seconds: 10),
          receiveTimeout: const Duration(seconds: 30),
          headers: {'Content-Type': 'application/json'},
        )) {
    dio.interceptors.add(InterceptorsWrapper(
      onRequest: (options, handler) async {
        final token = await _tokenStorage.getAccessToken();
        if (token != null) {
          options.headers['Authorization'] = 'Bearer $token';
        }
        handler.next(options);
      },
      onError: (error, handler) async {
        if (error.response?.statusCode == 401) {
          try {
            await _refreshAccessToken();
            final token = await _tokenStorage.getAccessToken();
            if (token != null) {
              error.requestOptions.headers['Authorization'] = 'Bearer $token';
              final response = await dio.fetch(error.requestOptions);
              return handler.resolve(response);
            }
          } catch (_) {
            // Refresh failed — propagate original 401
          }
        }
        handler.next(error);
      },
    ));
  }

  Future<void> _refreshAccessToken() async {
    // Serialize concurrent refresh attempts
    if (_refreshLock != null) {
      await _refreshLock!.future;
      return;
    }

    _refreshLock = Completer<void>();
    try {
      final refreshToken = await _tokenStorage.getRefreshToken();
      if (refreshToken == null) throw Exception('No refresh token');

      final response = await Dio(BaseOptions(
        baseUrl: dio.options.baseUrl,
        headers: {'Content-Type': 'application/json'},
      )).post('/api/auth/refresh', data: {'refresh_token': refreshToken});

      final accessToken = response.data['access_token'] as String;
      final newRefreshToken = response.data['refresh_token'] as String;
      await _tokenStorage.saveTokens(accessToken, newRefreshToken);

      _refreshLock!.complete();
    } catch (e) {
      _refreshLock!.completeError(e);
      await _tokenStorage.clear();
      rethrow;
    } finally {
      _refreshLock = null;
    }
  }
}
