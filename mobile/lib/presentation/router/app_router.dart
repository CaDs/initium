import 'package:go_router/go_router.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../providers/auth_provider.dart';
import '../../providers/api_provider.dart';
import '../landing/landing_screen.dart';
import '../login/login_screen.dart';
import '../home/home_screen.dart';

final routerProvider = Provider<GoRouter>((ref) {
  final authState = ref.watch(authProvider);

  return GoRouter(
    initialLocation: '/',
    redirect: (context, state) {
      final isAuthenticated = authState is AuthAuthenticated;
      final isLoading = authState is AuthLoading;
      final isOnLogin = state.matchedLocation == '/login';
      final isOnHome = state.matchedLocation == '/home';

      if (isLoading) return null;

      if (isAuthenticated && isOnLogin) return '/home';
      if (!isAuthenticated && isOnHome) return '/login';

      return null;
    },
    routes: [
      GoRoute(
        path: '/',
        builder: (context, state) => const LandingScreen(),
      ),
      GoRoute(
        path: '/login',
        builder: (context, state) => const LoginScreen(),
      ),
      GoRoute(
        path: '/home',
        builder: (context, state) => const HomeScreen(),
      ),
    ],
  );
});
