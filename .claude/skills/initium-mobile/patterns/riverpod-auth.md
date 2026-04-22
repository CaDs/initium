# Riverpod auth pattern

`AuthState` is a sealed class in `providers/auth_provider.dart` with four
variants: `AuthLoading`, `AuthAuthenticated`, `AuthUnauthenticated`, `AuthError`.
It is a **UI concern** — do not import `AuthState` into `domain/`.

## Consumption in a screen

```dart
class HomeScreen extends ConsumerWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final authState = ref.watch(authProvider);

    if (authState is! AuthAuthenticated) {
      return const Scaffold(body: Center(child: CircularProgressIndicator.adaptive()));
    }

    final user = authState.user;
    // render profile...
  }
}
```

## Triggering actions

```dart
// Logout
ref.read(authProvider.notifier).logout();

// Magic link
final success = await ref.read(authProvider.notifier).requestMagicLink(email);

// Google login
await ref.read(authProvider.notifier).loginWithGoogle(idToken);
```

## Rules

- `authProvider` (StateNotifierProvider) is declared in
  `providers/api_provider.dart`. The `AuthState` sealed class and its variants
  (`AuthLoading`, `AuthAuthenticated`, `AuthUnauthenticated`, `AuthError`) and
  the `AuthNotifier` class live in `providers/auth_provider.dart`.
- **Screens that read `authProvider` must import BOTH files** — the provider
  from `api_provider.dart`, the `AuthState` variants from `auth_provider.dart`
  (so pattern matching resolves). Copying either import alone will not compile.
- Match on variants with `switch` + pattern matching, not nullable checks.
- Never store `AuthState` in a domain entity. Never pass it to a repository.
  Only UI code reads it.
- Router redirects read `authProvider` through a `_AuthNotifier` that bridges
  to `go_router`'s `refreshListenable`. Do not duplicate redirect logic
  inside screens.

## When adding a new auth-sensitive screen

1. Add the route to `presentation/router/app_router.dart`.
2. Add a redirect rule if this screen should only show when authenticated
   (or unauthenticated).
3. In the screen, `ref.watch(authProvider)` and handle `AuthLoading` +
   `AuthError` explicitly — never render assuming `AuthAuthenticated`.
